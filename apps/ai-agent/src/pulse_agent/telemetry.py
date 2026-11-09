from __future__ import annotations

from collections.abc import Iterator
from contextlib import contextmanager


@contextmanager
def trace_span(_name: str, **_attrs: object) -> Iterator[None]:
    # Placeholder until OpenTelemetry exporters are configured by deployment.
    yield
