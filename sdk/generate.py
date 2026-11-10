#!/usr/bin/env python3
"""Reproduce seven clients with a pinned container; Python 3.12 and Docker/Podman."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
VERSION = "7.20.0"
IMAGE = "openapitools/openapi-generator-cli@sha256:fa4add01856e44becf70674164df354d61bd37ba0f444d27be949801e013921b"
TARGETS = {
    "go": ("go", {"packageName": "pulse", "gitRepoId": "pulse-sdk-go"}),
    "php": ("php", {"invokerPackage": "Riders\\Pulse", "gitRepoId": "pulse-sdk-php"}),
    "csharp": ("csharp", {"packageName": "Riders.Pulse", "targetFramework": "net8.0"}),
    "rust": ("rust", {"packageName": "pulse-sdk"}),
    "swift": ("swift6", {"projectName": "PulseSDK"}),
    "dart": ("dart", {"pubName": "pulse_sdk", "pubAuthor": "Riders Agency"}),
    "ruby": ("ruby", {"gemName": "pulse_sdk", "moduleName": "PulseSDK"}),
}


def normalize(directory: Path):
    for path in directory.rglob("*"):
        if not path.is_file():
            continue
        try:
            content = path.read_text()
        except UnicodeError:
            continue
        if "\0" in content:
            continue
        if path.name == "api_helper.dart":
            content = "\n".join(
                line
                for line in content.splitlines()
                if not line.startswith("// Ported from the ")
            )
        path.write_text(
            "\n".join(line.rstrip() for line in content.splitlines()).rstrip() + "\n"
        )


def finish(language: str, result: Path):
    for path in list(result.rglob("*")):
        if path.name in {
            ".github",
            ".git",
            ".openapi-generator",
            ".travis.yml",
            "git_push.sh",
            ".gitlab-ci.yml",
            ".circleci",
            "appveyor.yml",
        }:
            if path.is_dir():
                shutil.rmtree(path)
            elif path.exists():
                path.unlink()
    if language == "php":
        manifest = result / "composer.json"
        package = json.loads(manifest.read_text())
        package.pop("require-dev", None)
        package.pop("autoload-dev", None)
        package["extra"] = {
            "laravel": {"providers": ["Riders\\Pulse\\Laravel\\PulseServiceProvider"]}
        }
        package["suggest"] = {
            "laravel/framework": "Laravel request-scoped provider",
            "symfony/framework-bundle": "Symfony request factory and service configuration",
        }
        package.update(
            name="riders-agency/pulse-sdk",
            license="MIT",
            homepage="https://github.com/riders-agency/pulse-sdk-php",
            authors=[{"name": "Riders Agency", "email": "hello@riders.agency"}],
        )
        manifest.write_text(json.dumps(package, indent=2) + "\n")
    overlay = ROOT / "frameworks" / language
    if overlay.exists():
        shutil.copytree(overlay, result, dirs_exist_ok=True)
    license_path = ROOT.parent / "docs/public/LICENSE"
    shutil.copy2(
        license_path if license_path.exists() else ROOT.parent / "LICENSE",
        result / "LICENSE",
    )
    (result / "GENERATION.json").write_text(
        json.dumps(
            {
                "generator": TARGETS[language][0],
                "version": VERSION,
                "reproduction_image": IMAGE,
                "schema_sha256": hashlib.sha256(
                    (ROOT / "openapi.json").read_bytes()
                ).hexdigest(),
            },
            indent=2,
        )
        + "\n"
    )
    normalize(result)


def generate(language: str, engine: str, output: Path):
    name, properties = TARGETS[language]
    target = output / language
    if target.exists():
        raise RuntimeError(
            f"Output already exists: {target}. Choose a fresh --output directory."
        )
    output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="pulse-sdk-") as temp:
        temporary = Path(temp)
        config = temporary / "config.json"
        config.write_text(
            json.dumps(
                {
                    "gitUserId": "riders-agency",
                    "gitRepoId": "pulse-open-source",
                    "hideGenerationTimestamp": True,
                    "disallowAdditionalPropertiesIfNotPresent": False,
                    "packageVersion": "0.1.0",
                    "artifactVersion": "0.1.0",
                    "licenseName": "MIT",
                    "licenseUrl": "https://opensource.org/license/mit",
                    "gemAuthor": "Riders Agency",
                    "gemAuthorEmail": "hello@riders.agency",
                    "gemHomepage": "https://github.com/riders-agency/pulse-open-source",
                    "gemLicense": "MIT",
                    **properties,
                }
            )
        )
        result = temporary / "client"
        result.mkdir()
        subprocess.run(
            [
                engine,
                "run",
                "--rm",
                "--network",
                "none",
                "--user",
                f"{os.getuid()}:{os.getgid()}",
                "-v",
                f"{ROOT}:/source:ro",
                "-v",
                f"{config}:/config.json:ro",
                "-v",
                f"{result}:/out",
                IMAGE,
                "generate",
                "-i",
                "/source/openapi.json",
                "-g",
                name,
                "-c",
                "/config.json",
                "-o",
                "/out",
                "--global-property",
                "apiTests=false,modelTests=false",
            ],
            check=True,
        )
        finish(language, result)
        shutil.move(str(result), str(target))
    print(f"Generated {language}", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("language", choices=["all", *TARGETS])
    parser.add_argument("--output", type=Path, default=ROOT / "generated")
    parser.add_argument("--engine", choices=["docker", "podman"], default="docker")
    args = parser.parse_args()
    if not shutil.which(args.engine):
        parser.error(f"Container engine not found: {args.engine}")
    for language in TARGETS if args.language == "all" else [args.language]:
        generate(language, args.engine, args.output.resolve())
