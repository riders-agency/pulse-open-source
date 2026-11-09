from __future__ import annotations

import json
import os
import threading
from html import escape
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from time import monotonic
from typing import Any

from .fixtures import MAIL, NODES
from .graph import Office
from .model import Model
from .store import Store


def serve(directory: Path, port: int = 19862, origin: str = "http://127.0.0.1:19861") -> None:
    directory.mkdir(mode=0o700, parents=True, exist_ok=True)
    inbox = directory / "inbox"
    inbox.mkdir(exist_ok=True)
    store = Store(directory / "office.sqlite")
    office = Office(store, Model())
    lock = threading.Lock()
    stop = threading.Event()

    def monitor() -> None:
        while not stop.wait(0.5):
            with lock:
                if office.status == "watching":
                    letter = inbox / "demo-letter-001.json"
                    if letter.exists():
                        office.status = "idle"
                        try:
                            office.start(json.loads(letter.read_text()))
                        except Exception as error:
                            print(type(error).__name__ + ": запуск остановлен", flush=True)

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, format: str, *args: Any) -> None:
            pass

        def allowed(self) -> bool:
            return self.headers.get("Host") == f"127.0.0.1:{port}" and self.headers.get("Origin") == origin

        def respond(self, code: int, data: Any) -> None:
            self.send_response(code)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Cache-Control", "no-store")
            self.send_header("Access-Control-Allow-Origin", origin)
            self.send_header("Access-Control-Allow-Headers", "Content-Type")
            self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            self.end_headers()
            self.wfile.write(json.dumps(data, ensure_ascii=False).encode())

        def do_OPTIONS(self) -> None:
            self.respond(200 if self.allowed() else 403, {})

        def do_GET(self) -> None:
            if self.path == "/site" and self.headers.get("Host") == f"127.0.0.1:{port}":
                page = store.results(MAIL["id"]).get("website")
                title = escape(page["title"] if page else "Наши услуги")
                body = escape(page["body"] if page else "Новая услуга ещё не опубликована.")
                self.send_response(200)
                self.send_header("Content-Type", "text/html; charset=utf-8")
                self.send_header(
                    "Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'"
                )
                self.send_header("Cache-Control", "no-store")
                self.end_headers()
                self.wfile.write(
                    (
                        '<!doctype html><html lang="ru"><meta charset="utf-8"><title>Учебный сайт — Pulse</title>'
                        "<style>body{font:18px/1.7 system-ui;max-width:800px;margin:80px auto;padding:24px}"
                        "small{color:#777}h1{font-size:52px;line-height:1.1}</style>"
                        f"<small>Учебный сайт · ООО «Демо»</small><h1>{title}</h1><p>{body}</p></html>"
                    ).encode()
                )
                return
            if not self.allowed() or self.path != "/state":
                self.respond(403, {"error": "Недоступно"})
                return
            self.respond(
                200,
                {
                    "status": office.status,
                    "events": store.events(),
                    "nodes": NODES,
                    "edges": [[e.source, e.target] for e in office.graph.get_graph().edges],
                    "results": store.results(MAIL["id"]),
                },
            )

        def do_POST(self) -> None:
            if not self.allowed() or self.headers.get("Content-Type") != "application/json":
                self.respond(403, {"error": "Недоступно"})
                return
            try:
                size = int(self.headers.get("Content-Length", "0"))
                if not 0 < size <= 1024:
                    raise ValueError("Неверный размер запроса")
                data = json.loads(self.rfile.read(size))
                with lock:
                    if self.path == "/start" and office.status == "idle":
                        office.status = "watching"
                        office.started = monotonic()
                        office.emit(
                            "monitor",
                            "running",
                            "Проверяю учебный почтовый ящик каждые полсекунды.",
                            tool="mailbox.poll",
                        )
                    elif self.path == "/deliver" and office.status == "watching":
                        (inbox / "demo-letter-001.json").write_text(json.dumps(MAIL, ensure_ascii=False))
                    elif self.path == "/decision" and type(data.get("approve")) is bool:
                        office.decide(data["approve"])
                    else:
                        raise ValueError("Действие недоступно")
                self.respond(200, {"ok": True})
            except (ValueError, KeyError):
                self.respond(409, {"error": "Действие недоступно в текущем состоянии"})
            except Exception:
                self.respond(500, {"error": "Запуск остановлен. Проверьте локальный журнал."})

    threading.Thread(target=monitor, daemon=True).start()
    server = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    print(f"Office demo listening on 127.0.0.1:{port}; synthetic data only", flush=True)
    try:
        server.serve_forever()
    finally:
        stop.set()
        server.server_close()
        store.close()


if __name__ == "__main__":
    serve(Path(os.environ["OFFICE_DEMO_DIRECTORY"]))
