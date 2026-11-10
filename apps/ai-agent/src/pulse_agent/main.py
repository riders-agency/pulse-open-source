from __future__ import annotations

from pulse_agent.config import load_settings
from pulse_agent.grpc_server import serve
from pulse_agent.logging import configure_logging


def main() -> None:
    configure_logging()
    serve(load_settings())


if __name__ == "__main__":
    main()
