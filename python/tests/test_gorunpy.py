"""
Tests for the GoRunPy Python SDK.

These tests verify:
- The @export decorator works correctly
- Input/output validation
- Error handling
- JSON serialization
"""

import io
import json
import sys
from typing import Any, Dict, List, Optional
from unittest import mock

import pytest

from gorunpy import (
    GoRunPyError,
    ValidationError,
    export,
    main,
    registry,
)


class TestExportDecorator:
    """Tests for the @export decorator."""

    def setup_method(self):
        """Clear registry before each test."""
        registry._functions.clear()

    def test_export_simple_function(self):
        """Test that simple functions can be exported."""
        @export
        def test_func(a: int, b: int) -> int:
            return a + b

        assert "test_func" in registry.list_functions()
        func_info = registry.get("test_func")
        assert func_info is not None
        assert func_info.name == "test_func"

    def test_export_preserves_function(self):
        """Test that export preserves the original function."""
        @export
        def add_numbers(x: int, y: int) -> int:
            return x + y

        result = add_numbers(3, 4)
        assert result == 7

    def test_export_with_optional_params(self):
        """Test export with Optional parameters."""
        @export
        def greet_opt(name: str, prefix: Optional[str] = None) -> str:
            return f"{prefix or 'Hello'}, {name}!"

        func_info = registry.get("greet_opt")
        assert func_info is not None

    def test_export_with_complex_types(self):
        """Test export with complex JSON-serializable types."""
        @export
        def process_list(items: List[str]) -> Dict[str, int]:
            return {"count": len(items)}

        func_info = registry.get("process_list")
        assert func_info is not None


class TestInputValidation:
    """Tests for input validation."""

    def setup_method(self):
        """Clear registry before each test."""
        registry._functions.clear()

    def test_valid_input(self):
        """Test that valid input passes validation."""
        from gorunpy import _validate_value

        assert _validate_value(42, int, "test") == 42
        assert _validate_value("hello", str, "test") == "hello"
        assert _validate_value(True, bool, "test") is True

    def test_invalid_int(self):
        """Test that string for int raises TypeError."""
        from gorunpy import _validate_value, TypeError_

        with pytest.raises(TypeError_) as exc_info:
            _validate_value("not an int", int, "a")

        assert exc_info.value.field == "a"
        assert "expected int" in exc_info.value.message

    def test_invalid_str(self):
        """Test that int for str raises TypeError."""
        from gorunpy import _validate_value, TypeError_

        with pytest.raises(TypeError_) as exc_info:
            _validate_value(123, str, "b")

        assert exc_info.value.field == "b"

    def test_bool_not_int(self):
        """Test that bool is not accepted as int."""
        from gorunpy import _validate_value, TypeError_

        with pytest.raises(TypeError_) as exc_info:
            _validate_value(True, int, "a")

        assert "expected int, got bool" in exc_info.value.message

    def test_list_validation(self):
        """Test list type validation."""
        from gorunpy import _validate_value

        result = _validate_value([1, 2, 3], List[int], "items")
        assert result == [1, 2, 3]

    def test_list_item_validation(self):
        """Test that list item types are validated."""
        from gorunpy import _validate_value, TypeError_

        with pytest.raises(TypeError_) as exc_info:
            _validate_value([1, "two", 3], List[int], "items")

        assert "items[1]" in exc_info.value.field

    def test_dict_validation(self):
        """Test dict type validation."""
        from gorunpy import _validate_value

        result = _validate_value({"a": 1, "b": 2}, Dict[str, int], "data")
        assert result == {"a": 1, "b": 2}

    def test_optional_none(self):
        """Test that Optional accepts None."""
        from gorunpy import _validate_value

        result = _validate_value(None, Optional[str], "maybe")
        assert result is None

    def test_optional_value(self):
        """Test that Optional accepts the wrapped type."""
        from gorunpy import _validate_value

        result = _validate_value("hello", Optional[str], "maybe")
        assert result == "hello"

    def test_int_to_float_coercion(self):
        """Test that int is accepted for float."""
        from gorunpy import _validate_value

        result = _validate_value(42, float, "num")
        assert result == 42.0
        assert isinstance(result, float)


class TestDispatch:
    """Tests for request dispatch."""

    def setup_method(self):
        """Set up test fixtures."""
        registry._functions.clear()

        @export
        def dispatch_test(x: int) -> int:
            return x * 2

    def test_dispatch_success(self):
        """Test successful dispatch."""
        from gorunpy import _dispatch

        result = _dispatch({
            "function": "dispatch_test",
            "args": {"x": 5}
        })
        assert result == 10

    def test_dispatch_missing_function(self):
        """Test dispatch with missing function name."""
        from gorunpy import _dispatch, FunctionNotFoundError

        with pytest.raises(FunctionNotFoundError):
            _dispatch({
                "function": "nonexistent",
                "args": {}
            })

    def test_dispatch_missing_args(self):
        """Test dispatch with missing required argument."""
        from gorunpy import _dispatch

        with pytest.raises(ValidationError):
            _dispatch({
                "function": "dispatch_test",
                "args": {}  # missing 'x'
            })

    def test_dispatch_extra_args(self):
        """Test dispatch with unexpected arguments."""
        from gorunpy import _dispatch

        with pytest.raises(ValidationError) as exc_info:
            _dispatch({
                "function": "dispatch_test",
                "args": {"x": 1, "unexpected": 2}
            })

        assert "unexpected" in str(exc_info.value)


class TestMain:
    """Tests for the main entry point."""

    def setup_method(self):
        """Set up test fixtures."""
        registry._functions.clear()

        @export
        def main_test(a: int, b: int) -> int:
            return a + b

        @export
        def raises_error() -> None:
            raise ValidationError("intentional error", field="test")

        @export
        def raises_exception() -> None:
            raise RuntimeError("unexpected crash")

    def test_main_success(self):
        """Test main with valid input."""
        request = json.dumps({
            "function": "main_test",
            "args": {"a": 1, "b": 2}
        })

        with mock.patch.object(sys, 'stdin', io.StringIO(request)):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 0
        response = json.loads(stdout.getvalue())
        assert response["ok"] is True
        assert response["result"]["value"] == 3

    def test_main_validation_error(self):
        """Test main with validation error."""
        request = json.dumps({
            "function": "raises_error",
            "args": {}
        })

        with mock.patch.object(sys, 'stdin', io.StringIO(request)):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1
        response = json.loads(stderr.getvalue())
        assert response["ok"] is False
        assert response["error"]["kind"] == "ValidationError"

    def test_main_crash(self):
        """Test main with unhandled exception."""
        request = json.dumps({
            "function": "raises_exception",
            "args": {}
        })

        with mock.patch.object(sys, 'stdin', io.StringIO(request)):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 2
        response = json.loads(stderr.getvalue())
        assert response["ok"] is False
        assert response["error"]["kind"] == "RuntimeError"

    def test_main_invalid_json(self):
        """Test main with invalid JSON input."""
        with mock.patch.object(sys, 'stdin', io.StringIO("not json")):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1
        response = json.loads(stderr.getvalue())
        assert response["ok"] is False
        assert "JSON" in response["error"]["message"]

    def test_main_empty_input(self):
        """Test main with empty input."""
        with mock.patch.object(sys, 'stdin', io.StringIO("")):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1

    def test_main_type_error(self):
        """Test main with type mismatch."""
        request = json.dumps({
            "function": "main_test",
            "args": {"a": "not an int", "b": 2}
        })

        with mock.patch.object(sys, 'stdin', io.StringIO(request)):
            stdout = io.StringIO()
            stderr = io.StringIO()
            with mock.patch.object(sys, 'stdout', stdout):
                with mock.patch.object(sys, 'stderr', stderr):
                    with pytest.raises(SystemExit) as exc_info:
                        main()

        assert exc_info.value.code == 1
        response = json.loads(stderr.getvalue())
        assert response["error"]["kind"] == "TypeError"
        assert response["error"]["field"] == "a"


class TestIntrospection:
    """Tests for function introspection."""

    def setup_method(self):
        """Set up test fixtures."""
        registry._functions.clear()

        @export
        def introspect_test(a: int, b: str) -> bool:
            return True

    def test_introspect(self):
        """Test introspection returns function metadata."""
        from gorunpy import _introspect

        result = _introspect()
        assert "functions" in result

        func_names = [f["name"] for f in result["functions"]]
        assert "introspect_test" in func_names

        for func in result["functions"]:
            if func["name"] == "introspect_test":
                assert "a" in func["parameters"]
                assert "b" in func["parameters"]
                assert func["parameters"]["a"] == "int"
                assert func["parameters"]["b"] == "str"
                assert func["return_type"] == "bool"
