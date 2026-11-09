#!/usr/bin/env bash
set -euo pipefail
# Match gen-proto.sh and the descriptor baseline; no unpinned latest lookup.
: "${RUNNER_TEMP:?GitHub Actions temporary directory is required}"
: "${GITHUB_PATH:?GitHub Actions path file is required}"
[[ "$(uname -s)-$(uname -m)" == "Linux-x86_64" ]] || { echo "Expected Linux x86_64" >&2; exit 1; }
proto_tools_dir="$RUNNER_TEMP/pulse-proto-tools"
mkdir -p "$proto_tools_dir"
curl -fL --retry 3 --retry-delay 2 https://github.com/protocolbuffers/protobuf/releases/download/v3.12.4/protoc-3.12.4-linux-x86_64.zip -o "$proto_tools_dir/protoc.zip"
printf '%s  %s\n' d0d4c7a3c08d3ea9a20f94eaface12f5d46d7b023fe2057e834a4181c9e93ff3 "$proto_tools_dir/protoc.zip" | sha256sum -c -
unzip -oq "$proto_tools_dir/protoc.zip" -d "$proto_tools_dir/protoc"
curl -fL --retry 3 --retry-delay 2 https://github.com/bufbuild/buf/releases/download/v1.71.0/buf-Linux-x86_64 -o "$proto_tools_dir/buf"
printf '%s  %s\n' d3de2838c68a5759ca276884254bc70df4e4ad185d6ed5f65f327b6ce6363eab "$proto_tools_dir/buf" | sha256sum -c -
chmod +x "$proto_tools_dir/buf"
printf '%s\n' "$proto_tools_dir/protoc/bin" "$proto_tools_dir" >> "$GITHUB_PATH"
"$proto_tools_dir/protoc/bin/protoc" --version
"$proto_tools_dir/buf" --version
