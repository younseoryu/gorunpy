#!/usr/bin/env python3
"""
GoRunPy CLI - Simple build tool for Go-callable Python.

Usage:
    gorunpy build ./mymodule --output ./dist
    gorunpy run ./dist/mymodule sum --args '{"a": 1, "b": 2}'
"""

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path


def find_module_entry(module_path: Path) -> Path:
    """Find the __main__.py entry point."""
    main_file = module_path / "__main__.py"
    if main_file.exists():
        return main_file
    raise FileNotFoundError(f"No __main__.py found in {module_path}")


def build(module_path: str, output_dir: str, name: str = None):
    """Build a Python module into a standalone executable."""
    module = Path(module_path).resolve()
    
    if not module.exists():
        print(f"Error: {module} does not exist", file=sys.stderr)
        sys.exit(1)
    
    # Determine module name
    module_name = name or module.name
    
    # Find entry point
    entry_point = find_module_entry(module)
    
    # Find gorunpy package location
    import gorunpy
    gorunpy_path = Path(gorunpy.__file__).parent.parent
    
    # Output directory
    out = Path(output_dir).resolve()
    out.mkdir(parents=True, exist_ok=True)
    
    print(f"Building {module_name} from {module}...")
    
    # Build PyInstaller command
    cmd = [
        sys.executable, "-m", "PyInstaller",
        "--onefile",
        "--name", module_name,
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
        print("Build failed!", file=sys.stderr)
        sys.exit(1)
    
    executable = out / module_name
    print(f"\n✓ Built: {executable}")
    print(f"\nTest it:")
    print(f"  echo '{{\"function\":\"__introspect__\",\"args\":{{}}}}' | {executable}")
    print(f"\nGenerate Go client:")
    print(f"  gorunpy-gen -binary {executable} -package {module_name} -output {module_name}/client.go")


def run(binary: str, function: str, args: str = "{}"):
    """Run a function in the executable."""
    request = {
        "function": function,
        "args": json.loads(args),
    }
    
    proc = subprocess.run(
        [binary],
        input=json.dumps(request),
        capture_output=True,
        text=True,
    )
    
    if proc.returncode == 0:
        response = json.loads(proc.stdout)
        print(json.dumps(response.get("result", {}).get("value"), indent=2))
    else:
        print(proc.stderr, file=sys.stderr)
        sys.exit(proc.returncode)


def list_functions(binary: str):
    """List all exported functions in an executable."""
    request = {"function": "__introspect__", "args": {}}
    
    proc = subprocess.run(
        [binary],
        input=json.dumps(request),
        capture_output=True,
        text=True,
    )
    
    if proc.returncode != 0:
        print(proc.stderr, file=sys.stderr)
        sys.exit(1)
    
    response = json.loads(proc.stdout)
    functions = response.get("result", {}).get("value", {}).get("functions", [])
    
    print("Exported functions:\n")
    for func in functions:
        if func["name"].startswith("_"):
            continue
        params = ", ".join(f"{k}: {v}" for k, v in func["parameters"].items())
        print(f"  {func['name']}({params}) -> {func['return_type']}")


def main():
    parser = argparse.ArgumentParser(
        description="GoRunPy - Build Go-callable Python executables",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  gorunpy build ./mymodule --output ./dist
  gorunpy list ./dist/mymodule
  gorunpy run ./dist/mymodule sum '{"a": 1, "b": 2}'
        """,
    )
    
    subparsers = parser.add_subparsers(dest="command", required=True)
    
    # build command
    build_parser = subparsers.add_parser("build", help="Build module into executable")
    build_parser.add_argument("module", help="Path to Python module directory")
    build_parser.add_argument("--output", "-o", default="./dist", help="Output directory")
    build_parser.add_argument("--name", "-n", help="Executable name (default: module name)")
    
    # list command
    list_parser = subparsers.add_parser("list", help="List exported functions")
    list_parser.add_argument("binary", help="Path to executable")
    
    # run command
    run_parser = subparsers.add_parser("run", help="Run a function")
    run_parser.add_argument("binary", help="Path to executable")
    run_parser.add_argument("function", help="Function name")
    run_parser.add_argument("args", nargs="?", default="{}", help="JSON arguments")
    
    args = parser.parse_args()
    
    if args.command == "build":
        build(args.module, args.output, args.name)
    elif args.command == "list":
        list_functions(args.binary)
    elif args.command == "run":
        run(args.binary, args.function, args.args)


if __name__ == "__main__":
    main()

