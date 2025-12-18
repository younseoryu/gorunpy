"""
Example exported functions demonstrating GoRunPy usage.

All functions use type annotations and are decorated with @export.
"""

from builtins import sum as builtin_sum
from typing import Dict, List, Optional

from gorunpy import ValidationError, export


@export
def sum(a: int, b: int) -> int:
    """Add two integers."""
    return a + b


@export
def multiply(a: float, b: float) -> float:
    """Multiply two floats."""
    return a * b


@export
def divide(a: float, b: float) -> float:
    """Divide two numbers. Raises ValidationError if b is zero."""
    if b == 0:
        raise ValidationError("division by zero", field="b")
    return a / b


@export
def greet(name: str, greeting: Optional[str] = None) -> str:
    """Generate a greeting message."""
    if greeting is None:
        greeting = "Hello"
    return f"{greeting}, {name}!"


@export
def get_stats(numbers: List[float]) -> Dict[str, float]:
    """Calculate basic statistics for a list of numbers."""
    if not numbers:
        raise ValidationError("numbers list cannot be empty", field="numbers")

    total = builtin_sum(numbers)
    count = len(numbers)
    mean = total / count

    sorted_nums = sorted(numbers)
    mid = count // 2
    if count % 2 == 0:
        median = (sorted_nums[mid - 1] + sorted_nums[mid]) / 2
    else:
        median = sorted_nums[mid]

    return {
        "sum": total,
        "count": float(count),
        "mean": mean,
        "median": median,
        "min": min(numbers),
        "max": max(numbers),
    }


@export
def concat(strings: List[str]) -> str:
    """Concatenate a list of strings."""
    return "".join(strings)


@export
def echo(value: str) -> str:
    """Echo back a string value."""
    return value

