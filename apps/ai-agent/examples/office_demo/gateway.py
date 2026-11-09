from __future__ import annotations

from typing import Any

from .store import Store

PERMISSIONS = {
    "dispatcher": {"mailbox.poll", "department.assign"},
    "content": {"website.prepare"},
    "accounting": {"invoice.prepare"},
    "controller": {"office.apply", "receipt.read"},
}


class Gateway:
    def __init__(self, store: Store):
        self.store = store

    def call(self, actor: str, tool: str, data: dict[str, Any]) -> dict[str, Any]:
        if tool not in PERMISSIONS.get(actor, set()):
            raise PermissionError("Инструмент недоступен этому отделу")
        if tool == "mailbox.poll":
            return {"received": self.store.receive(data), "subject": data["subject"]}
        if tool == "department.assign":
            departments = data["departments"]
            if not departments or not set(departments) <= {"content", "accounting", "support"}:
                raise ValueError("Неизвестный отдел")
            return {"departments": departments}
        if tool == "website.prepare":
            if set(data) != {"title", "body"}:
                raise ValueError("Неверный формат страницы")
            title, body = data["title"], data["body"]
            if (
                not isinstance(title, str)
                or not isinstance(body, str)
                or not 1 <= len(title) <= 120
                or not 1 <= len(body) <= 1200
            ):
                raise ValueError("Неверный текст страницы")
            return {"path": "/services/interface-audit", "title": title, "body": body, "status": "published_in_demo"}
        if tool == "invoice.prepare":
            order = data["order"]
            if type(order["amount_minor"]) is not int or order["amount_minor"] <= 0 or order["currency"] != "RUB":
                raise ValueError("Неверная сумма заказа")
            return {**order, "number": "DEMO-104", "status": "draft", "bank_connected": False}
        if tool == "office.apply":
            # Only the graph controller can supply a human decision; models cannot call this endpoint.
            if type(data.get("approved")) is not bool:
                raise PermissionError("Нужно решение человека")
            return self.store.apply(data["mail_id"], data["artifacts"], data["approved"])
        if tool == "receipt.read":
            return self.store.results(data["mail_id"])
        raise ValueError("Неизвестный инструмент")
