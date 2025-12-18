#!/usr/bin/env python3
import argparse
import json
import os
import subprocess
import sys
from pathlib import Path


def build(module_path: str, output_dir: str, name: str = None):
    module = Path(module_path).resolve()
    if not module.exists():
        print(f"Error: {module} does not exist", file=sys.stderr)
        sys.exit(1)

    module_name = name or module.name
    entry_point = module / "__main__.py"
    if not entry_point.exists():
        print(f"Error: {entry_point} not found", file=sys.stderr)
        sys.exit(1)

    import gorunpy
    gorunpy_path = Path(gorunpy.__file__).parent.parent

    out = Path(output_dir).resolve()
    out.mkdir(parents=True, exist_ok=True)

    print(f"Building {module_name}...")

    cmd = [
        sys.executable, "-m", "PyInstaller",
        "--onefile", "--name", module_name,
        "--paths", str(module.parent),
        "--paths", str(gorunpy_path),
        "--distpath", str(out),
        "--workpath", str(out / ".build"),
        "--specpath", str(out / ".build"),
        "--collect-all", "gorunpy",
        "--hidden-import", "gorunpy",
        "--hidden-import", module_name,
        "--hidden-import", f"{module_name}.functions",
        "--log-level", "WARN",
        str(entry_point),
    ]

    result = subprocess.run(cmd)
    if result.returncode != 0:
        sys.exit(1)

    print(f"Built: {out / module_name}")


def run(binary: str, function: str, args: str = "{}"):
    proc = subprocess.run(
        [binary],
        input=json.dumps({"function": function, "args": json.loads(args)}),
        capture_output=True, text=True,
    )
    if proc.returncode == 0:
        print(json.dumps(json.loads(proc.stdout).get("result", {}).get("value"), indent=2))
    else:
        print(proc.stderr, file=sys.stderr)
        sys.exit(proc.returncode)


def list_functions(binary: str):
    proc = subprocess.run(
        [binary],
        input=json.dumps({"function": "__introspect__", "args": {}}),
        capture_output=True, text=True,
    )
    if proc.returncode != 0:
        print(proc.stderr, file=sys.stderr)
        sys.exit(1)

    functions = json.loads(proc.stdout).get("result", {}).get("value", {}).get("functions", [])
    for f in functions:
        if not f["name"].startswith("_"):
            params = ", ".join(f"{k}: {v}" for k, v in f["parameters"].items())
            print(f"{f['name']}({params}) -> {f['return_type']}")


def main():
    parser = argparse.ArgumentParser(prog="gorunpy")
    sub = parser.add_subparsers(dest="cmd", required=True)

    b = sub.add_parser("build")
    b.add_argument("module")
    b.add_argument("-o", "--output", default="./dist")
    b.add_argument("-n", "--name")

    l = sub.add_parser("list")
    l.add_argument("binary")

    r = sub.add_parser("run")
    r.add_argument("binary")
    r.add_argument("function")
    r.add_argument("args", nargs="?", default="{}")

    args = parser.parse_args()

    if args.cmd == "build":
        build(args.module, args.output, args.name)
    elif args.cmd == "list":
        list_functions(args.binary)
    elif args.cmd == "run":
        run(args.binary, args.function, args.args)


if __name__ == "__main__":
    main()
