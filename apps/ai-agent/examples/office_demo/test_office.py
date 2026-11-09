from copy import deepcopy

import pytest

from .fixtures import MAIL
from .gateway import Gateway
from .graph import Office
from .model import validate
from .store import Store

CLASSIFICATION = {
    "departments": ["content", "accounting"],
    "title": "Аудит интерфейса",
    "body": "Проверим путь клиента и подготовим рекомендации.",
    "summary": "Нужно обновить сайт и подготовить счёт.",
}


class FakeModel:
    def prepare_page(self, mail):
        return {key: CLASSIFICATION[key] for key in ("title", "body")}, {
            "provider_key": "test",
            "model_id": "fixture",
            "input_tokens": 0,
            "output_tokens": 0,
        }

    def classify(self, mail):
        return validate({key: deepcopy(CLASSIFICATION[key]) for key in ("departments", "summary")}), {
            "provider_key": "test",
            "model_id": "fixture",
            "input_tokens": 0,
            "output_tokens": 0,
        }


def test_graph_pauses_before_writes_and_persists_after_approval(tmp_path):
    path = tmp_path / "office.sqlite"
    store = Store(path)
    office = Office(store, FakeModel())
    office.start(deepcopy(MAIL))
    assert office.status == "waiting_approval"
    assert store.results(MAIL["id"]) == {}
    assert office.graph.get_state(office.config).next == ("approval",)
    office.decide(True)
    assert office.status == "completed"
    result = store.results(MAIL["id"])
    assert result["invoice"]["amount_minor"] == 4500000
    assert result["invoice"]["bank_connected"] is False
    assert result["website"]["title"] == "Аудит интерфейса"
    assert store.apply(MAIL["id"], result, True) == result
    with pytest.raises(ValueError):
        office.decide(True)
    store.close()
    reopened = Store(path)
    assert reopened.results(MAIL["id"]) == result
    assert not reopened.receive(MAIL)
    reopened.close()


def test_rejection_never_updates_site_or_invoice(tmp_path):
    store = Store(tmp_path / "office.sqlite")
    office = Office(store, FakeModel())
    office.start(deepcopy(MAIL))
    office.decide(False)
    assert office.status == "rejected"
    assert store.results(MAIL["id"]) == {}
    assert not any(event["node"] == "apply" for event in store.events())
    with pytest.raises(ValueError):
        store.apply(MAIL["id"], {"website": {}}, True)
    store.close()


def test_department_cannot_apply_changes_or_access_other_department(tmp_path):
    store = Store(tmp_path / "office.sqlite")
    gateway = Gateway(store)
    for actor, tool in [
        ("content", "office.apply"),
        ("accounting", "website.prepare"),
        ("dispatcher", "invoice.prepare"),
        ("unknown", "mailbox.poll"),
    ]:
        with pytest.raises(PermissionError):
            gateway.call(actor, tool, {})
    with pytest.raises(PermissionError):
        gateway.call("controller", "office.apply", {"approved": "true"})
    store.close()


@pytest.mark.parametrize("extra", [{"approved": True}, {"tool": "office.apply"}, {"amount_minor": 1}])
def test_untrusted_model_cannot_supply_privileges_or_amount(extra):
    with pytest.raises(ValueError):
        validate({"departments": ["content"], "summary": "Текст", **extra})


@pytest.mark.parametrize("departments", [[], ["admin"], [False], "accounting"])
def test_invalid_classification_is_rejected(departments):
    with pytest.raises(ValueError):
        validate({"departments": departments, "summary": "Текст"})


def test_model_failure_does_not_claim_success_or_write(tmp_path):
    class BrokenModel:
        def classify(self, mail):
            raise TimeoutError("unavailable")

    store = Store(tmp_path / "office.sqlite")
    office = Office(store, BrokenModel())
    with pytest.raises(TimeoutError):
        office.start(deepcopy(MAIL))
    assert office.status == "failed"
    assert not store.results(MAIL["id"])
    assert store.events()[-1]["status"] == "failed"
    store.close()
