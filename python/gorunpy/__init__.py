"""
GoRunPy - Python SDK for Go-callable typed executables.

This module provides the infrastructure to export Python functions
that can be called from Go with type safety and proper error handling.
"""

from __future__ import annotations

import json
import sys
import traceback
from dataclasses import dataclass
from enum import Enum
from functools import wraps
from typing import (
    Any,
    Callable,
    Dict,
    Generic,
    List,
    Optional,
    Type,
    TypeVar,
    Union,
    get_args,
    get_origin,
    get_type_hints,
)

__all__ = ["export", "main", "GoRunPyError", "ValidationError", "registry", "__version__"]

__version__ = "0.1.0"

T = TypeVar("T")


class ExitCode(Enum):
    """Exit codes for the executable."""
    SUCCESS = 0
    HANDLED_ERROR = 1
    CRASH = 2


@dataclass
class FunctionInfo:
    """Metadata about an exported function."""
    name: str
    func: Callable
    type_hints: Dict[str, Any]
    return_type: Any


class Registry:
    """Registry of exported functions."""
    
    def __init__(self):
        self._functions: Dict[str, FunctionInfo] = {}
    
    def register(self, name: str, func: Callable, type_hints: Dict[str, Any], return_type: Any):
        """Register a function with its type information."""
        self._functions[name] = FunctionInfo(
            name=name,
            func=func,
            type_hints=type_hints,
            return_type=return_type,
        )
    
    def get(self, name: str) -> Optional[FunctionInfo]:
        """Get a registered function by name."""
        return self._functions.get(name)
    
    def list_functions(self) -> List[str]:
        """List all registered function names."""
        return list(self._functions.keys())


# Global registry instance
registry = Registry()


class GoRunPyError(Exception):
    """Base exception for GoRunPy errors."""
    
    def __init__(self, message: str, kind: str = "Error", field: Optional[str] = None):
        super().__init__(message)
        self.message = message
        self.kind = kind
        self.field = field
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to error response dict."""
        result = {
            "kind": self.kind,
            "message": self.message,
        }
        if self.field is not None:
            result["field"] = self.field
        return result


class ValidationError(GoRunPyError):
    """Raised when input validation fails."""
    
    def __init__(self, message: str, field: Optional[str] = None):
        super().__init__(message, kind="ValidationError", field=field)


class TypeError_(GoRunPyError):
    """Raised when type validation fails."""
    
    def __init__(self, message: str, field: Optional[str] = None):
        super().__init__(message, kind="TypeError", field=field)


class FunctionNotFoundError(GoRunPyError):
    """Raised when requested function is not found."""
    
    def __init__(self, function_name: str):
        super().__init__(
            f"function '{function_name}' not found",
            kind="FunctionNotFoundError",
        )


def _is_json_serializable_type(type_hint: Any) -> bool:
    """Check if a type hint represents a JSON-serializable type."""
    origin = get_origin(type_hint)
    
    # Handle None/NoneType
    if type_hint is type(None):
        return True
    
    # Basic JSON types
    if type_hint in (int, float, str, bool, type(None)):
        return True
    
    # Any is allowed (no validation)
    if type_hint is Any:
        return True
    
    # List[T]
    if origin is list:
        args = get_args(type_hint)
        if not args:
            return True
        return _is_json_serializable_type(args[0])
    
    # Dict[str, T]
    if origin is dict:
        args = get_args(type_hint)
        if not args:
            return True
        if args[0] is not str:
            return False
        return _is_json_serializable_type(args[1])
    
    # Optional[T] / Union[T, None]
    if origin is Union:
        args = get_args(type_hint)
        return all(_is_json_serializable_type(arg) for arg in args)
    
    return False


def _validate_value(value: Any, type_hint: Any, field_name: str) -> Any:
    """Validate and coerce a value against a type hint."""
    origin = get_origin(type_hint)
    
    # Handle None
    if value is None:
        if type_hint is type(None):
            return None
        # Check if Optional
        if origin is Union:
            args = get_args(type_hint)
            if type(None) in args:
                return None
        raise TypeError_(f"expected {type_hint}, got None", field=field_name)
    
    # Handle Any - no validation
    if type_hint is Any:
        return value
    
    # Handle Union/Optional
    if origin is Union:
        args = get_args(type_hint)
        errors = []
        for arg in args:
            if arg is type(None) and value is None:
                return None
            if arg is not type(None):
                try:
                    return _validate_value(value, arg, field_name)
                except (TypeError_, ValidationError) as e:
                    errors.append(str(e))
        # None of the union types matched
        type_names = [str(arg) for arg in args if arg is not type(None)]
        raise TypeError_(
            f"expected one of {type_names}, got {type(value).__name__}",
            field=field_name
        )
    
    # Handle List[T]
    if origin is list:
        if not isinstance(value, list):
            raise TypeError_(f"expected list, got {type(value).__name__}", field=field_name)
        args = get_args(type_hint)
        if args:
            item_type = args[0]
            return [
                _validate_value(item, item_type, f"{field_name}[{i}]")
                for i, item in enumerate(value)
            ]
        return value
    
    # Handle Dict[str, T]
    if origin is dict:
        if not isinstance(value, dict):
            raise TypeError_(f"expected dict, got {type(value).__name__}", field=field_name)
        args = get_args(type_hint)
        if args:
            key_type, value_type = args
            result = {}
            for k, v in value.items():
                if not isinstance(k, str):
                    raise TypeError_(
                        f"expected string key, got {type(k).__name__}",
                        field=f"{field_name}.{k}"
                    )
                result[k] = _validate_value(v, value_type, f"{field_name}.{k}")
            return result
        return value
    
    # Handle basic types
    if type_hint is int:
        if isinstance(value, bool):  # bool is subclass of int
            raise TypeError_(f"expected int, got bool", field=field_name)
        if not isinstance(value, int):
            raise TypeError_(f"expected int, got {type(value).__name__}", field=field_name)
        return value
    
    if type_hint is float:
        if isinstance(value, bool):
            raise TypeError_(f"expected float, got bool", field=field_name)
        if isinstance(value, int):
            return float(value)  # int is valid for float
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
    
    # Unknown type - pass through
    return value


def _validate_return_value(value: Any, type_hint: Any) -> Any:
    """Validate the return value against its type hint."""
    if type_hint is None or type_hint is type(None):
        return value
    return _validate_value(value, type_hint, "return")


def export(func: Callable[..., T]) -> Callable[..., T]:
    """
    Decorator to export a function for Go calls.
    
    The function must have full type annotations for all parameters
    and the return type. All types must be JSON-serializable.
    
    Example:
        @export
        def add(a: int, b: int) -> int:
            return a + b
    """
    try:
        type_hints = get_type_hints(func)
    except Exception as e:
        raise ValueError(f"Cannot get type hints for {func.__name__}: {e}")
    
    # Extract return type
    return_type = type_hints.pop("return", Any)
    
    # Validate all type hints are JSON-serializable
    for param_name, param_type in type_hints.items():
        if not _is_json_serializable_type(param_type):
            raise ValueError(
                f"Parameter '{param_name}' of function '{func.__name__}' "
                f"has non-JSON-serializable type: {param_type}"
            )
    
    if not _is_json_serializable_type(return_type):
        raise ValueError(
            f"Return type of function '{func.__name__}' "
            f"is not JSON-serializable: {return_type}"
        )
    
    # Register the function
    registry.register(func.__name__, func, type_hints, return_type)
    
    @wraps(func)
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)
    
    return wrapper


def _make_success_response(result: Any) -> Dict[str, Any]:
    """Create a success response."""
    return {
        "ok": True,
        "result": {
            "value": result,
        },
    }


def _make_error_response(error: GoRunPyError) -> Dict[str, Any]:
    """Create an error response."""
    return {
        "ok": False,
        "error": error.to_dict(),
    }


def _dispatch(request: Dict[str, Any]) -> Any:
    """Dispatch a request to the appropriate function."""
    # Validate request structure
    if "function" not in request:
        raise ValidationError("missing 'function' field in request")
    
    function_name = request["function"]
    if not isinstance(function_name, str):
        raise ValidationError("'function' must be a string")
    
    # Get the function
    func_info = registry.get(function_name)
    if func_info is None:
        raise FunctionNotFoundError(function_name)
    
    # Get and validate args
    args = request.get("args", {})
    if not isinstance(args, dict):
        raise ValidationError("'args' must be an object")
    
    # Validate each argument
    validated_args = {}
    for param_name, param_type in func_info.type_hints.items():
        if param_name not in args:
            # Check if parameter has a default or is Optional
            origin = get_origin(param_type)
            if origin is Union and type(None) in get_args(param_type):
                validated_args[param_name] = None
            else:
                raise ValidationError(f"missing required argument '{param_name}'", field=param_name)
        else:
            validated_args[param_name] = _validate_value(
                args[param_name], param_type, param_name
            )
    
    # Check for unexpected arguments
    expected_params = set(func_info.type_hints.keys())
    provided_params = set(args.keys())
    unexpected = provided_params - expected_params
    if unexpected:
        raise ValidationError(f"unexpected argument(s): {', '.join(sorted(unexpected))}")
    
    # Call the function
    result = func_info.func(**validated_args)
    
    # Validate return value
    result = _validate_return_value(result, func_info.return_type)
    
    return result


def _type_to_string(type_hint: Any) -> str:
    """Convert a type hint to its string representation."""
    if type_hint is type(None):
        return "None"
    
    if type_hint is Any:
        return "Any"
    
    if type_hint in (int, float, str, bool):
        return type_hint.__name__
    
    origin = get_origin(type_hint)
    args = get_args(type_hint)
    
    if origin is list:
        if args:
            return f"List[{_type_to_string(args[0])}]"
        return "List"
    
    if origin is dict:
        if args:
            return f"Dict[{_type_to_string(args[0])}, {_type_to_string(args[1])}]"
        return "Dict"
    
    if origin is Union:
        # Check for Optional (Union[T, None])
        non_none_args = [a for a in args if a is not type(None)]
        if len(non_none_args) == 1 and type(None) in args:
            return f"Optional[{_type_to_string(non_none_args[0])}]"
        arg_strs = [_type_to_string(a) for a in args]
        return f"Union[{', '.join(arg_strs)}]"
    
    # Fallback
    return str(type_hint)


def _introspect() -> Dict[str, Any]:
    """Return metadata about all registered functions."""
    functions = []
    for name in registry.list_functions():
        func_info = registry.get(name)
        if func_info is not None:
            functions.append({
                "name": func_info.name,
                "parameters": {
                    param_name: _type_to_string(param_type)
                    for param_name, param_type in func_info.type_hints.items()
                },
                "return_type": _type_to_string(func_info.return_type),
            })
    return {"functions": functions}


# Register introspection as a special function
def _register_introspect():
    """Register the __introspect__ function."""
    registry.register(
        "__introspect__",
        _introspect,
        {},  # No parameters
        Dict[str, Any],  # Return type
    )


def main():
    """
    Main entry point for the executable.
    
    Reads a JSON request from stdin, dispatches to the appropriate
    function, and writes the response to stdout (success) or stderr (error).
    """
    # Register introspection function
    _register_introspect()
    
    exit_code = ExitCode.SUCCESS
    
    try:
        # Read request from stdin
        raw_input = sys.stdin.read()
        if not raw_input.strip():
            raise ValidationError("empty input")
        
        try:
            request = json.loads(raw_input)
        except json.JSONDecodeError as e:
            raise ValidationError(f"invalid JSON: {e}")
        
        # Dispatch and get result
        result = _dispatch(request)
        
        # Write success response to stdout
        response = _make_success_response(result)
        sys.stdout.write(json.dumps(response))
        sys.stdout.flush()
        
    except GoRunPyError as e:
        # Handled error - write to stderr with exit code 1
        exit_code = ExitCode.HANDLED_ERROR
        response = _make_error_response(e)
        sys.stderr.write(json.dumps(response))
        sys.stderr.flush()
        
    except Exception as e:
        # Unhandled exception - crash
        exit_code = ExitCode.CRASH
        error = GoRunPyError(
            message=f"{type(e).__name__}: {str(e)}\n{traceback.format_exc()}",
            kind="RuntimeError",
        )
        response = _make_error_response(error)
        sys.stderr.write(json.dumps(response))
        sys.stderr.flush()
    
    sys.exit(exit_code.value)

