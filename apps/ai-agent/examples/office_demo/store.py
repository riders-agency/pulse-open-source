from __future__ import annotations

import json
import sqlite3
import threading
from pathlib import Path
from typing import Any


class Store:
    """Local example storage. Unique mail IDs and atomic approval prevent duplicate effects."""

    def __init__(self, path: Path):
        self.lock = threading.RLock()
        self.db = sqlite3.connect(path, check_same_thread=False)
        self.db.executescript("""
            CREATE TABLE IF NOT EXISTS messages (id TEXT PRIMARY KEY, payload TEXT NOT NULL);
            CREATE TABLE IF NOT EXISTS events (seq INTEGER PRIMARY KEY, payload TEXT NOT NULL);
            CREATE TABLE IF NOT EXISTS artifacts (
                mail_id TEXT, kind TEXT, payload TEXT NOT NULL, PRIMARY KEY (mail_id, kind));
            CREATE TABLE IF NOT EXISTS decisions (mail_id TEXT PRIMARY KEY, approved INTEGER NOT NULL);
        """)

    def receive(self, mail: dict[str, Any]) -> bool:
        with self.lock, self.db:
            return (
                self.db.execute(
                    "INSERT OR IGNORE INTO messages VALUES (?, ?)", (mail["id"], json.dumps(mail, ensure_ascii=False))
                ).rowcount
                == 1
            )

    def event(self, event: dict[str, Any]) -> None:
        with self.lock, self.db:
            self.db.execute("INSERT INTO events(payload) VALUES (?)", (json.dumps(event, ensure_ascii=False),))

    def events(self) -> list[dict[str, Any]]:
        with self.lock:
            return [json.loads(row[0]) for row in self.db.execute("SELECT payload FROM events ORDER BY seq")]

    def apply(self, mail_id: str, artifacts: dict[str, Any], approved: bool) -> dict[str, Any]:
        with self.lock, self.db:
            prior = self.db.execute("SELECT approved FROM decisions WHERE mail_id = ?", (mail_id,)).fetchone()
            if prior:
                if bool(prior[0]) != approved:
                    raise ValueError("Решение уже принято")
                return self.results(mail_id)
            self.db.execute("INSERT INTO decisions VALUES (?, ?)", (mail_id, int(approved)))
            if approved:
                for kind, value in artifacts.items():
                    self.db.execute(
                        "INSERT INTO artifacts VALUES (?, ?, ?)", (mail_id, kind, json.dumps(value, ensure_ascii=False))
                    )
            return self.results(mail_id)

    def results(self, mail_id: str) -> dict[str, Any]:
        with self.lock:
            return {
                kind: json.loads(payload)
                for kind, payload in self.db.execute(
                    "SELECT kind, payload FROM artifacts WHERE mail_id = ?", (mail_id,)
                )
            }

    def close(self) -> None:
        self.db.close()
