from builtins import sum as builtin_sum
from typing import Dict, List, Optional
from gorunpy import ValidationError, export

@export
def sum(a: int, b: int) -> int:
    return a + b

@export
def multiply(a: float, b: float) -> float:
    return a * b

@export
def divide(a: float, b: float) -> float:
    if b == 0:
        raise ValidationError("division by zero", field="b")
    return a / b

@export
def greet(name: str, greeting: Optional[str] = None) -> str:
    return f"{greeting or 'Hello'}, {name}!"

@export
def get_stats(numbers: List[float]) -> Dict[str, float]:
    if not numbers:
        raise ValidationError("empty list", field="numbers")
    total = builtin_sum(numbers)
    return {"sum": total, "count": float(len(numbers)), "mean": total / len(numbers)}

