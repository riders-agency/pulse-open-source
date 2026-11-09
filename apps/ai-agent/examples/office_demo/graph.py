from __future__ import annotations

from time import monotonic
from typing import Any, TypedDict

from langgraph.checkpoint.memory import InMemorySaver
from langgraph.graph import END, START, StateGraph
from langgraph.types import Command, interrupt

from .fixtures import MAIL, NODES
from .gateway import Gateway
from .model import Model
from .store import Store


class State(TypedDict, total=False):
    mail: dict[str, Any]
    received: bool
    classification: dict[str, Any]
    drafts: dict[str, Any]
    approved: bool
    result: dict[str, Any]


class Office:
    def __init__(self, store: Store, model: Model):
        self.store, self.model, self.gateway = store, model, Gateway(store)
        self.started = monotonic()
        self.status = "idle"
        self.config = {"configurable": {"thread_id": MAIL["id"]}}
        builder = StateGraph(State)
        for node in NODES:
            builder.add_node(node["id"], getattr(self, node["id"]))
        builder.add_edge(START, "monitor")
        builder.add_conditional_edges(
            "monitor", lambda s: "classify" if s["received"] else END, {"classify": "classify", END: END}
        )
        builder.add_edge("classify", "route")
        builder.add_edge("route", "content")
        builder.add_edge("content", "accounting")
        builder.add_edge("accounting", "approval")
        builder.add_conditional_edges(
            "approval", lambda s: "apply" if s["approved"] else "done", {"apply": "apply", "done": "done"}
        )
        builder.add_edge("apply", "done")
        builder.add_edge("done", END)
        self.graph = builder.compile(checkpointer=InMemorySaver())

    def emit(self, node: str, status: str, text: str, **data: Any) -> None:
        self.store.event(
            {"node": node, "status": status, "text": text, "at_ms": round((monotonic() - self.started) * 1000), **data}
        )

    def start(self, mail: dict[str, Any]) -> None:
        if self.status != "idle":
            raise ValueError("Запуск уже начат")
        self.status = "running"
        if not self.store.events():
            self.started = monotonic()
        self.invoke({"mail": mail, "drafts": {}})

    def decide(self, approve: bool) -> None:
        if self.status != "waiting_approval" or type(approve) is not bool:
            raise ValueError("Нет ожидающего подтверждения")
        self.status = "running"
        self.emit(
            "approval", "completed", "Подтверждаю изменения." if approve else "Отклоняю изменения.", actor="human"
        )
        self.invoke(Command(resume=approve))

    def invoke(self, value: Any) -> None:
        try:
            result = self.graph.invoke(value, self.config)
            if result.get("__interrupt__"):
                self.status = "waiting_approval"
                self.emit(
                    "approval",
                    "waiting",
                    "Всё подготовлено. Обновить учебный сайт и сохранить черновик счёта?",
                    drafts=result["drafts"],
                    tool="human.approve",
                )
            else:
                self.status = "completed" if result.get("approved") else "rejected"
        except Exception:
            self.status = "failed"
            self.emit(
                "done", "failed", "Не удалось завершить запуск. Изменения не подтверждены; проверьте локальный журнал."
            )
            raise

    def monitor(self, state: State) -> dict[str, Any]:
        result = self.gateway.call("dispatcher", "mailbox.poll", state["mail"])
        self.emit(
            "monitor",
            "completed",
            "Получено новое письмо: «Обновление сайта и счёт».",
            tool="mailbox.poll",
            mail=state["mail"],
        )
        return {"received": result["received"]}

    def classify(self, state: State) -> dict[str, Any]:
        self.emit("classify", "running", "Читаю письмо и определяю, какие отделы нужны.", tool="model.classify")
        result, usage = self.model.classify(state["mail"])
        self.emit("classify", "completed", result["summary"], tool="model.classify", usage=usage)
        return {"classification": result}

    def route(self, state: State) -> dict[str, Any]:
        result = self.gateway.call("dispatcher", "department.assign", state["classification"])
        names = {"content": "ИИ-отдел", "accounting": "Бухгалтерия", "support": "Поддержка"}
        self.emit(
            "route",
            "completed",
            "Передаю поручения: " + ", ".join(names[d] for d in result["departments"]) + ".",
            tool="department.assign",
            departments=result["departments"],
        )
        return {}

    def content(self, state: State) -> dict[str, Any]:
        if "content" not in state["classification"]["departments"]:
            self.emit("content", "skipped", "Изменение сайта не требуется.")
            return {}
        self.emit("content", "running", "ИИ-отдел готовит текст страницы по письму.", tool="website.prepare")
        page, usage = self.model.prepare_page(state["mail"])
        draft = self.gateway.call("content", "website.prepare", page)
        self.emit(
            "content",
            "completed",
            "Подготовлена страница «" + draft["title"] + "». Публикация ждёт вашего решения.",
            tool="website.prepare",
            usage=usage,
        )
        return {"drafts": {**state["drafts"], "website": draft}}

    def accounting(self, state: State) -> dict[str, Any]:
        if "accounting" not in state["classification"]["departments"]:
            self.emit("accounting", "skipped", "Счёт не требуется.")
            return {}
        draft = self.gateway.call("accounting", "invoice.prepare", state["mail"])
        self.emit(
            "accounting",
            "completed",
            "Подготовлен счёт на 45 000 ₽. Сумма взята из заказа DEMO-104.",
            tool="invoice.prepare",
        )
        return {"drafts": {**state["drafts"], "invoice": draft}}

    def approval(self, state: State) -> dict[str, Any]:
        decision = interrupt({"drafts": state["drafts"]})
        if type(decision) is not bool:
            raise ValueError("Неверное решение")
        return {"approved": decision}

    def apply(self, state: State) -> dict[str, Any]:
        result = self.gateway.call(
            "controller",
            "office.apply",
            {
                "mail_id": state["mail"]["id"],
                "artifacts": state["drafts"],
                "approved": state["approved"],
            },
        )
        self.emit("apply", "completed", "Подтверждение принято. Изменения сохранены.", tool="office.apply")
        return {"result": result}

    def done(self, state: State) -> dict[str, Any]:
        result = self.gateway.call("controller", "receipt.read", {"mail_id": state["mail"]["id"]})
        outcomes = []
        if "website" in result:
            outcomes.append("учебный сайт обновлён")
        if "invoice" in result:
            outcomes.append("черновик счёта сохранён")
        text = (
            "Готово: " + ", ".join(outcomes) + ". В банк ничего не отправлялось."
            if outcomes
            else "Обращение передано в поддержку."
        )
        if not state.get("approved"):
            self.store.apply(state["mail"]["id"], {}, False)
            text = "Отклонено. Сайт и счета не изменены."
        self.emit("done", "completed", text, tool="receipt.read", results=result)
        return {"result": result}
