#!/usr/bin/env python3
"""Check SDK route coverage, local references and generated schema provenance."""

import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
schema = json.loads((ROOT / "openapi.json").read_text())
inventory = {
    tuple(line.split("\t")[:2])
    for line in (ROOT.parent / "contracts/openapi/endpoints.inventory.tsv")
    .read_text()
    .splitlines()
}
operations = {
    (method.upper(), path)
    for path, methods in schema["paths"].items()
    for method in methods
}
assert operations <= inventory, (
    "SDK contains routes absent from the application inventory"
)


def references(node):
    if isinstance(node, dict):
        if "$ref" in node:
            reference = node["$ref"]
            assert reference.startswith("#/"), "SDK must not resolve remote references"
            target = schema
            for key in reference[2:].split("/"):
                target = target[key.replace("~1", "/").replace("~0", "~")]
        for value in node.values():
            references(value)
    elif isinstance(node, list):
        for value in node:
            references(value)


references(schema)
digest = hashlib.sha256((ROOT / "openapi.json").read_bytes()).hexdigest()
manifests = list((ROOT / "generated").glob("*/GENERATION.json"))
assert len(manifests) == 7, "Expected all seven generated clients"
for manifest in manifests:
    assert json.loads(manifest.read_text())["schema_sha256"] == digest, (
        f"Stale client: {manifest.parent.name}"
    )
print(
    f"SDK: {len(operations)} application routes, local references, {len(manifests)} matching client manifests"
)
