"""Export the compiled LangGraph topology used by the technical showcase."""
from __future__ import annotations

import argparse
import json
from pathlib import Path

from pulse_agent.graph.builder import build_graph


def graph_contract() -> dict[str, object]:
    builder = build_graph()
    if builder is None:
        raise RuntimeError("LangGraph must be installed to export its topology")
    graph = builder.compile().get_graph()
    return {
        "schema": "pulse-agent-graph-v1",
        "source": "apps/ai-agent/src/pulse_agent/graph/builder.py",
        "nodes": sorted(graph.nodes),
        "edges": sorted(
            [{"source": edge.source, "target": edge.target, "conditional": edge.conditional} for edge in graph.edges],
            key=lambda edge: (str(edge["source"]), str(edge["target"])),
        ),
    }


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    content = json.dumps(graph_contract(), ensure_ascii=False, indent=2) + "\n"
    if args.check:
        if not args.output.exists() or args.output.read_text() != content:
            raise SystemExit("Agent graph contract differs from the compiled runtime")
    else:
        args.output.parent.mkdir(parents=True, exist_ok=True)
        args.output.write_text(content)


if __name__ == "__main__":
    main()
