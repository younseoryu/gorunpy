from builtins import sum as builtin_sum
from typing import Dict, List, Optional

import gorunpy


@gorunpy.export
def sum(a: int, b: int) -> int:
    return a + b


@gorunpy.export
def multiply(a: float, b: float) -> float:
    return a * b


@gorunpy.export
def divide(a: float, b: float) -> float:
    if b == 0:
        raise gorunpy.ValidationError("division by zero", field="b")
    return a / b


@gorunpy.export
def greet(name: str, greeting: Optional[str] = None) -> str:
    return f"{greeting or 'Hello'}, {name}!"


@gorunpy.export
def get_stats(numbers: List[float]) -> Dict[str, float]:
    if not numbers:
        raise gorunpy.ValidationError("empty list", field="numbers")
    total = builtin_sum(numbers)
    count = len(numbers)
    return {"sum": total, "count": float(count), "mean": total / count}

