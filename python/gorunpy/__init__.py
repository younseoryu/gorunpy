from __future__ import annotations

import json
import sys
import traceback
from dataclasses import dataclass
from enum import Enum
from functools import wraps
from typing import Any, Callable, Dict, List, Optional, TypeVar, Union, get_args, get_origin, get_type_hints

__all__ = ["export", "main", "GoRunPyError", "ValidationError", "registry"]

T = TypeVar("T")


class ExitCode(Enum):
    SUCCESS = 0
    HANDLED_ERROR = 1
    CRASH = 2


@dataclass
class FunctionInfo:
    name: str
    func: Callable
    type_hints: Dict[str, Any]
    return_type: Any


class Registry:
    def __init__(self):
        self._functions: Dict[str, FunctionInfo] = {}

    def register(self, name: str, func: Callable, type_hints: Dict[str, Any], return_type: Any):
        self._functions[name] = FunctionInfo(name, func, type_hints, return_type)

    def get(self, name: str) -> Optional[FunctionInfo]:
        return self._functions.get(name)

    def list_functions(self) -> List[str]:
        return list(self._functions.keys())


registry = Registry()


class GoRunPyError(Exception):
    def __init__(self, message: str, kind: str = "Error", field: Optional[str] = None):
        super().__init__(message)
        self.message = message
        self.kind = kind
        self.field = field

    def to_dict(self) -> Dict[str, Any]:
        result = {"kind": self.kind, "message": self.message}
        if self.field:
            result["field"] = self.field
        return result


class ValidationError(GoRunPyError):
    def __init__(self, message: str, field: Optional[str] = None):
        super().__init__(message, kind="ValidationError", field=field)


class TypeError_(GoRunPyError):
    def __init__(self, message: str, field: Optional[str] = None):
        super().__init__(message, kind="TypeError", field=field)


class FunctionNotFoundError(GoRunPyError):
    def __init__(self, function_name: str):
        super().__init__(f"function '{function_name}' not found", kind="FunctionNotFoundError")


def _validate_value(value: Any, type_hint: Any, field_name: str) -> Any:
    origin = get_origin(type_hint)

    if value is None:
        if type_hint is type(None):
            return None
        if origin is Union and type(None) in get_args(type_hint):
            return None
        raise TypeError_(f"expected {type_hint}, got None", field=field_name)

    if type_hint is Any:
        return value

    if origin is Union:
        for arg in get_args(type_hint):
            if arg is type(None):
                continue
            try:
                return _validate_value(value, arg, field_name)
            except (TypeError_, ValidationError):
                pass
        raise TypeError_(f"type mismatch", field=field_name)

    if origin is list:
        if not isinstance(value, list):
            raise TypeError_(f"expected list, got {type(value).__name__}", field=field_name)
        args = get_args(type_hint)
        if args:
            return [_validate_value(item, args[0], f"{field_name}[{i}]") for i, item in enumerate(value)]
        return value

    if origin is dict:
        if not isinstance(value, dict):
            raise TypeError_(f"expected dict, got {type(value).__name__}", field=field_name)
        args = get_args(type_hint)
        if args:
            return {k: _validate_value(v, args[1], f"{field_name}.{k}") for k, v in value.items()}
        return value

    if type_hint is int:
        if isinstance(value, bool) or not isinstance(value, int):
            raise TypeError_(f"expected int, got {type(value).__name__}", field=field_name)
        return value

    if type_hint is float:
        if isinstance(value, bool):
            raise TypeError_(f"expected float, got bool", field=field_name)
        if isinstance(value, int):
            return float(value)
        if not isinstance(value, float):
            raise TypeError_(f"expected float, got {type(value).__name__}", field=field_name)
        return value

    if type_hint is str:
        if not isinstance(value, str):
            raise TypeError_(f"expected str, got {type(value).__name__}", field=field_name)
        return value

    if type_hint is bool:
        if not isinstance(value, bool):
            raise TypeError_(f"expected bool, got {type(value).__name__}", field=field_name)
        return value

    return value


def export(func: Callable[..., T]) -> Callable[..., T]:
    type_hints = get_type_hints(func)
    return_type = type_hints.pop("return", Any)
    registry.register(func.__name__, func, type_hints, return_type)

    @wraps(func)
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)

    return wrapper


def _dispatch(request: Dict[str, Any]) -> Any:
    if "function" not in request:
        raise ValidationError("missing 'function' field")

    function_name = request["function"]
    func_info = registry.get(function_name)
    if func_info is None:
        raise FunctionNotFoundError(function_name)

    args = request.get("args", {})
    if not isinstance(args, dict):
        raise ValidationError("'args' must be an object")

    validated_args = {}
    for param_name, param_type in func_info.type_hints.items():
        if param_name not in args:
            origin = get_origin(param_type)
            if origin is Union and type(None) in get_args(param_type):
                validated_args[param_name] = None
            else:
                raise ValidationError(f"missing required argument '{param_name}'", field=param_name)
        else:
            validated_args[param_name] = _validate_value(args[param_name], param_type, param_name)

    unexpected = set(args.keys()) - set(func_info.type_hints.keys())
    if unexpected:
        raise ValidationError(f"unexpected argument(s): {', '.join(sorted(unexpected))}")

    return func_info.func(**validated_args)


def _type_to_string(type_hint: Any) -> str:
    if type_hint is type(None):
        return "None"
    if type_hint is Any:
        return "Any"
    if type_hint in (int, float, str, bool):
        return type_hint.__name__

    origin = get_origin(type_hint)
    args = get_args(type_hint)

    if origin is list:
        return f"List[{_type_to_string(args[0])}]" if args else "List"
    if origin is dict:
        return f"Dict[{_type_to_string(args[0])}, {_type_to_string(args[1])}]" if args else "Dict"
    if origin is Union:
        non_none = [a for a in args if a is not type(None)]
        if len(non_none) == 1 and type(None) in args:
            return f"Optional[{_type_to_string(non_none[0])}]"
        return f"Union[{', '.join(_type_to_string(a) for a in args)}]"
    return str(type_hint)


def _introspect() -> Dict[str, Any]:
    functions = []
    for name in registry.list_functions():
        func_info = registry.get(name)
        if func_info:
            functions.append({
                "name": func_info.name,
                "parameters": {k: _type_to_string(v) for k, v in func_info.type_hints.items()},
                "return_type": _type_to_string(func_info.return_type),
            })
    return {"functions": functions}


def main():
    registry.register("__introspect__", _introspect, {}, Dict[str, Any])

    exit_code = ExitCode.SUCCESS
    try:
        raw_input = sys.stdin.read()
        if not raw_input.strip():
            raise ValidationError("empty input")

        request = json.loads(raw_input)
        result = _dispatch(request)
        sys.stdout.write(json.dumps({"ok": True, "result": {"value": result}}))
        sys.stdout.flush()

    except GoRunPyError as e:
        exit_code = ExitCode.HANDLED_ERROR
        sys.stderr.write(json.dumps({"ok": False, "error": e.to_dict()}))
        sys.stderr.flush()

    except Exception as e:
        exit_code = ExitCode.CRASH
        error = GoRunPyError(f"{type(e).__name__}: {e}\n{traceback.format_exc()}", kind="RuntimeError")
        sys.stderr.write(json.dumps({"ok": False, "error": error.to_dict()}))
        sys.stderr.flush()

    sys.exit(exit_code.value)
