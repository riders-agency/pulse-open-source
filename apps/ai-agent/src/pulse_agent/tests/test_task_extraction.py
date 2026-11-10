import pytest

from pulse_agent.graph.task_extraction import extract_client_query, parse_task_draft


@pytest.mark.parametrize("noun", ["презентацию проекта", "запуск сайта", "договор"])
def test_task_command_is_not_a_client_preposition(noun: str) -> None:
    draft = parse_task_draft(f"Создай задачу подготовить {noun} на завтра", local_date="2026-11-08")
    assert draft.client_query == ""
    assert draft.title == f"Подготовить {noun}"
    assert draft.due_date == "2026-11-09"
    assert draft.missing_fields == []


@pytest.mark.parametrize("text", ["Задача для клиента Орион завтра", "Проверь оплату у Орион завтра"])
def test_explicit_client_reference_is_preserved(text: str) -> None:
    assert extract_client_query(text) == "Орион"


@pytest.mark.parametrize("command", ["Запланируй", "Заведи", "Поставь", "Добавьте", "Создайте"])
def test_task_title_does_not_include_the_request_verb(command: str) -> None:
    draft = parse_task_draft(f"{command} задачу подготовить презентацию Pulse на завтра.", local_date="2026-11-09")
    assert draft.title == "Подготовить презентацию Pulse"
    assert draft.due_date == "2026-11-10"
    assert draft.missing_fields == []
