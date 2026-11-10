from __future__ import annotations

from pulse_agent.eval_catalog import load_eval_catalog, metric, score_catalog, threshold_results


def test_eval_catalog_has_required_suites_and_case_classes() -> None:
    catalog = load_eval_catalog()
    cases = catalog["cases"]
    suites = {case["suite"] for case in cases}
    classes = {case["class"] for case in cases}

    assert catalog["schema"] == "riders-pulse-ai-eval-catalog-v1"
    assert suites == {"stable", "shadow", "adversarial"}
    assert {
        "ambiguity",
        "approval",
        "concurrency",
        "mixed_language",
        "multi_turn",
        "safety",
        "typos",
        "workflow_patch",
    } <= classes
    assert all(case["turns"] for case in cases)


def test_eval_catalog_scorer_meets_hard_thresholds() -> None:
    catalog = load_eval_catalog()
    score = score_catalog(catalog)
    totals = score["totals"]

    assert metric(totals, "stable_intent") == 1.0
    assert metric(totals, "stable_policy") == 1.0
    assert metric(totals, "adversarial_safety") == 1.0
    assert all(threshold_results(catalog, score).values())
    assert score["per_class"]["multi_turn"]["intent_total"] == 6
