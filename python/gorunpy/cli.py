#!/usr/bin/env python3
"""GoRunPy CLI - Build Python modules for Go consumption."""

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import List, Optional, Dict, Any, Tuple

# Directories to skip during auto-detection
SKIP_DIRS = {
    "venv", ".venv", "env", ".env",
    "__pycache__", ".git", ".hg", ".svn",
    "node_modules", "dist", "build", ".build",
    ".gorunpy", ".tox", ".pytest_cache",
    "site-packages", ".eggs", "*.egg-info",
}

MAX_DEPTH_UP = 3
MAX_DEPTH_DOWN = 3


def find_gorunpy_module(start_dir: Path, max_up: int = MAX_DEPTH_UP, max_down: int = MAX_DEPTH_DOWN) -> Optional[Path]:
    """Find nearest Python module containing @gorunpy.export decorators."""
    start = start_dir.resolve()
    
    # Search starting from current dir, expanding outward
    # Level 0: current dir only
    # Level 1: current dir children + parent
    # Level 2: parent's children + grandparent
    # etc.
    
    dirs_to_check = [start]
    checked = set()
    
    # First check current directory itself
    if _is_gorunpy_module(start):
        return start
    checked.add(start)
    
    # Check immediate children of start dir (depth 1 down)
    module = _search_children(start, checked, max_down)
    if module:
        return module
    
    # Now expand upward, checking each parent and its children
    current = start
    for level in range(max_up):
        parent = current.parent
        if parent == current:  # reached root
            break
        
        # Check parent itself
        if parent not in checked:
            if _is_gorunpy_module(parent):
                return parent
            checked.add(parent)
        
        # Check parent's children (siblings of current + their descendants)
        module = _search_children(parent, checked, max_down)
        if module:
            return module
        
        current = parent
    
    return None


def _search_children(directory: Path, checked: set, max_depth: int) -> Optional[Path]:
    """Search children of directory for gorunpy modules."""
    if max_depth < 0:
        return None
    
    try:
        for child in sorted(directory.iterdir()):  # sorted for deterministic order
            if child in checked or not child.is_dir() or _should_skip(child):
                continue
            
            checked.add(child)
            
            if _is_gorunpy_module(child):
                return child
            
            # Search deeper
            result = _search_children(child, checked, max_depth - 1)
            if result:
                return result
    except PermissionError:
        pass
    
    return None


def _should_skip(path: Path) -> bool:
    """Check if path should be skipped."""
    name = path.name
    if name.startswith("."):
        return True
    for skip in SKIP_DIRS:
        if skip.endswith("*"):
            if name.startswith(skip[:-1]):
                return True
        elif name == skip:
            return True
    return False




def _is_gorunpy_module(directory: Path) -> bool:
    """Check if directory is a Python module with @gorunpy.export."""
    init_file = directory / "__init__.py"
    if not init_file.exists():
        return False
    
    # Look for @gorunpy.export in any .py file
    try:
        for py_file in directory.glob("*.py"):
            content = py_file.read_text(encoding="utf-8", errors="ignore")
            if "@gorunpy.export" in content or "gorunpy.export" in content:
                return True
    except Exception:
        pass
    
    return False


def _find_exported_files(module_dir: Path) -> List[Path]:
    """Find all .py files in module that have @gorunpy.export."""
    exported = []
    for py_file in module_dir.glob("*.py"):
        if py_file.name.startswith("_"):
            continue
        try:
            content = py_file.read_text(encoding="utf-8", errors="ignore")
            if "@gorunpy.export" in content:
                exported.append(py_file)
        except Exception:
            pass
    return sorted(exported)


def generate_main_py(module_dir: Path) -> str:
    """Generate __main__.py content for a module."""
    module_name = module_dir.name
    exported_files = _find_exported_files(module_dir)
    
    imports = []
    for f in exported_files:
        imports.append(f"import {module_name}.{f.stem}")
    
    lines = imports + [
        "import gorunpy",
        "",
        'if __name__ == "__main__":',
        "    gorunpy.main()",
        "",
    ]
    return "\n".join(lines)


def ensure_main_py(module_dir: Path) -> Path:
    """Ensure __main__.py exists, create if missing."""
    main_py = module_dir / "__main__.py"
    if not main_py.exists():
        content = generate_main_py(module_dir)
        main_py.write_text(content)
        print(f"Generated: {main_py}")
    return main_py


def build_binary(module_dir: Path, output_dir: Path, name: Optional[str] = None) -> Path:
    """Build Python module into executable binary."""
    module_name = name or module_dir.name
    entry_point = ensure_main_py(module_dir)
    
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Find all exported files for hidden imports
    exported_files = _find_exported_files(module_dir)
    hidden_imports = [
        "--hidden-import", "gorunpy",
        "--hidden-import", module_name,
    ]
    for f in exported_files:
        hidden_imports.extend(["--hidden-import", f"{module_name}.{f.stem}"])

    print(f"Building {module_name}...")

    cmd = [
        sys.executable, "-m", "PyInstaller",
        "--onefile", "--noconfirm",
        "--name", module_name,
        "--paths", str(module_dir.parent),
        "--distpath", str(output_dir),
        "--workpath", str(output_dir / ".build"),
        "--specpath", str(output_dir / ".build"),
        *hidden_imports,
        "--log-level", "WARN",
        str(entry_point),
    ]

    result = subprocess.run(cmd)
    if result.returncode != 0:
        print("Error: Build failed", file=sys.stderr)
        sys.exit(1)
    
    binary_path = output_dir / module_name
    print(f"Built: {binary_path}")
    return binary_path


def introspect_binary(binary_path: Path) -> List[Dict[str, Any]]:
    """Get function signatures from binary via introspection."""
    proc = subprocess.run(
        [str(binary_path)],
        input=json.dumps({"function": "__introspect__", "args": {}}),
        capture_output=True,
        text=True,
    )
    
    if proc.returncode != 0:
        print(f"Error: Introspection failed: {proc.stderr}", file=sys.stderr)
        sys.exit(1)

    data = json.loads(proc.stdout)
    functions = data.get("result", {}).get("value", {}).get("functions", [])
    
    # Filter out private functions
    return [f for f in functions if not f["name"].startswith("_")]


# Go code generation
def py_type_to_go(t: str) -> str:
    """Convert Python type to Go type."""
    type_map = {
        "int": "int",
        "float": "float64",
        "str": "string",
        "bool": "bool",
        "None": "",
        "NoneType": "",
        "Any": "any",
    }
    
    if t in type_map:
        return type_map[t]
    
    if t.startswith("List["):
        inner = t[5:-1]
        return "[]" + py_type_to_go(inner)
    
    if t.startswith("Dict["):
        inner = t[5:-1]
        parts = inner.split(", ", 1)
        if len(parts) == 2:
            return "map[string]" + py_type_to_go(parts[1])
        return "map[string]any"
    
    if t.startswith("Optional["):
        inner = t[9:-1]
        return "*" + py_type_to_go(inner)
    
    return "any"


def go_zero_value(t: str) -> str:
    """Get Go zero value for a type."""
    go_type = py_type_to_go(t)
    if go_type == "int":
        return "0"
    if go_type == "float64":
        return "0"
    if go_type == "string":
        return '""'
    if go_type == "bool":
        return "false"
    if go_type == "":
        return ""
    return "nil"


def to_go_name(name: str) -> str:
    """Convert snake_case to PascalCase."""
    return "".join(word.capitalize() for word in name.split("_"))


def to_go_param_name(name: str) -> str:
    """Convert to Go parameter name (camelCase)."""
    if not name:
        return name
    return name[0].lower() + name[1:]


def generate_go_client(
    functions: List[Dict[str, Any]],
    package_name: str,
    binary_path: str,
    module_name: str,
    module_path: str = "github.com/younseoryu/gorunpy/gorunpy",
) -> str:
    """Generate Go client code."""
    # Convert module name to PascalCase for Go naming
    client_name = to_go_name(module_name) + "Client"
    
    lines = [f"package {package_name}", ""]
    
    # Imports
    lines.append("import (")
    lines.append('\t"context"')
    lines.append('\t_ "embed"')
    lines.append('\t"os"')
    lines.append('\t"path/filepath"')
    lines.append(f'\t"{module_path}"')
    lines.append(")")
    lines.append("")
    
    # Embed directive and extraction logic
    lines.append(f"//go:embed {binary_path}")
    lines.append("var embeddedBinary []byte")
    lines.append("")
    lines.append("var extractedBinaryPath string")
    lines.append("")
    lines.append("func init() {")
    lines.append('\tdir, err := os.MkdirTemp("", "gorunpy-*")')
    lines.append("\tif err != nil {")
    lines.append('\t\tpanic("gorunpy: failed to create temp dir: " + err.Error())')
    lines.append("\t}")
    lines.append(f'\textractedBinaryPath = filepath.Join(dir, "{Path(binary_path).name}")')
    lines.append("\tif err := os.WriteFile(extractedBinaryPath, embeddedBinary, 0755); err != nil {")
    lines.append('\t\tpanic("gorunpy: failed to write binary: " + err.Error())')
    lines.append("\t}")
    lines.append("}")
    lines.append("")
    
    # Client struct
    lines.append(f"type {client_name} struct {{")
    lines.append("\t*gorunpy.Client")
    lines.append("}")
    lines.append("")
    
    # NewClient function
    lines.append(f"func New{client_name}() *{client_name} {{")
    lines.append(f"\treturn &{client_name}{{Client: gorunpy.NewClient(extractedBinaryPath)}}")
    lines.append("}")
    lines.append("")
    
    # Generate method for each function
    for func in functions:
        name = func["name"]
        params = func.get("parameters", {})
        return_type = func.get("return_type", "None")
        
        go_name = to_go_name(name)
        go_ret = py_type_to_go(return_type)
        go_zero = go_zero_value(return_type)
        
        # Build parameter list (preserves original Python function order)
        param_items = list(params.items())
        
        param_parts = ["ctx context.Context"]
        for pname, ptype in param_items:
            go_pname = to_go_param_name(pname)
            go_ptype = py_type_to_go(ptype)
            param_parts.append(f"{go_pname} {go_ptype}")
        
        # Build return type
        if go_ret:
            ret_part = f"({go_ret}, error)"
        else:
            ret_part = "error"
        
        # Function signature
        lines.append(f"func (c *{client_name}) {go_name}({', '.join(param_parts)}) {ret_part} {{")
        
        # Args map
        args_parts = []
        for pname, _ in param_items:
            go_pname = to_go_param_name(pname)
            args_parts.append(f'"{pname}": {go_pname}')
        lines.append(f"\targs := map[string]any{{{', '.join(args_parts)}}}")
        
        # Call and return
        if go_ret:
            lines.append(f"\tvar result {go_ret}")
            lines.append(f'\tif err := c.Call(ctx, "{name}", args, &result); err != nil {{')
            lines.append(f"\t\treturn {go_zero}, err")
            lines.append("\t}")
            lines.append("\treturn result, nil")
        else:
            lines.append(f'\treturn c.Call(ctx, "{name}", args, nil)')
        
        lines.append("}")
        lines.append("")
    
    return "\n".join(lines)


def detect_go_package(directory: Path) -> str:
    """Detect Go package name from existing .go files."""
    for go_file in directory.glob("*.go"):
        try:
            content = go_file.read_text()
            match = re.search(r"^package\s+(\w+)", content, re.MULTILINE)
            if match:
                return match.group(1)
        except Exception:
            pass
    return "main"


def gen(
    module_path: Optional[str] = None,
    output_dir: Optional[str] = None,
    client_output: Optional[str] = None,
):
    """Main generation command - auto-detect, build, and generate Go client."""
    cwd = Path.cwd()
    
    # Find or use specified module
    if module_path:
        module_dir = Path(module_path).resolve()
        if not module_dir.exists():
            print(f"Error: {module_dir} does not exist", file=sys.stderr)
            sys.exit(1)
        if not _is_gorunpy_module(module_dir):
            print(f"Error: {module_dir} is not a valid gorunpy module", file=sys.stderr)
            print("Hint: Make sure it has __init__.py and files with @gorunpy.export", file=sys.stderr)
            sys.exit(1)
    else:
        # Auto-detect (searches nearest first, up to 3 levels up/down)
        print("Searching for gorunpy module...")
        module_dir = find_gorunpy_module(cwd)
        
        if not module_dir:
            print("Error: No gorunpy module found", file=sys.stderr)
            print("Hint: Create a Python package with @gorunpy.export decorated functions", file=sys.stderr)
            sys.exit(1)
        
        print(f"Found module: {module_dir}")
    
    module_name = module_dir.name
    
    # Determine output paths
    gorunpy_dir = cwd / ".gorunpy"
    binary_output_dir = Path(output_dir) if output_dir else gorunpy_dir
    binary_path = binary_output_dir / module_name
    
    # Build binary
    build_binary(module_dir, binary_output_dir, module_name)
    
    # Introspect
    print("Introspecting functions...")
    functions = introspect_binary(binary_path)
    print(f"Found {len(functions)} exported function(s)")
    
    # Detect package name
    go_package = detect_go_package(cwd)
    
    # Generate Go client
    relative_binary = binary_path.relative_to(cwd)
    client_code = generate_go_client(
        functions=functions,
        package_name=go_package,
        binary_path=str(relative_binary),
        module_name=module_name,
    )
    
    # Write client
    client_file = Path(client_output) if client_output else cwd / "gorunpy_client.go"
    client_file.write_text(client_code)
    print(f"Generated: {client_file}")
    
    client_type = to_go_name(module_name) + "Client"
    print("\nDone! Usage:")
    print(f"  client := New{client_type}()")
    print("  result, err := client.YourFunction(ctx, args...)")


def main():
    parser = argparse.ArgumentParser(
        prog="gorunpy",
        description="Build Python modules for Go consumption",
    )
    parser.add_argument("module", nargs="?", help="Path to Python module (default: pylib/)")
    parser.add_argument("-o", "--output", help="Output directory for binary (default: .gorunpy/)")
    parser.add_argument("--client", help="Output path for Go client (default: gorunpy_client.go)")

    args = parser.parse_args()

    gen(
        module_path=args.module,
        output_dir=args.output,
        client_output=args.client,
    )


if __name__ == "__main__":
    main()
