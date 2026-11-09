# Riders Pulse Desktop (stub)

CMake + Qt 6 skeleton: polls `GET ${RP_API_BASE:-http://127.0.0.1:8080}/healthz`.

```bash
sudo apt install qt6-base-dev cmake build-essential
cmake -S . -B build -DCMAKE_PREFIX_PATH="$(qtpaths6 --qt-version 2>/dev/null || echo /usr)"
cmake --build build
RP_API_BASE=https://pulse.riders.agency ./build/riders-pulse-desktop
```

Pairing and message polling endpoints from the Pulse plan must be wired in follow-up commits.
