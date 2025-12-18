"""Data analysis using pandas.

Common data processing tasks exposed to Go.
"""

import gorunpy
from typing import Optional

# Ensure PyInstaller bundles jinja2 (used by pandas for some operations)
import jinja2  # noqa: F401


@gorunpy.export
def read_csv_stats(file_path: str) -> dict[str, dict]:
    """Read CSV and return descriptive statistics for each column."""
    import pandas as pd

    df = pd.read_csv(file_path)
    stats = df.describe(include="all").to_dict()
    
    # Convert NaN to None for JSON serialization
    for col in stats:
        for key in stats[col]:
            if pd.isna(stats[col][key]):
                stats[col][key] = None
    
    return stats


@gorunpy.export
def csv_to_json(file_path: str) -> list[dict]:
    """Convert CSV file to list of JSON objects."""
    import pandas as pd

    df = pd.read_csv(file_path)
    return df.to_dict(orient="records")


@gorunpy.export
def filter_csv(
    file_path: str,
    column: str,
    operator: str,
    value: float,
) -> list[dict]:
    """Filter CSV rows by condition. Operators: gt, lt, eq, gte, lte."""
    import pandas as pd

    df = pd.read_csv(file_path)
    
    ops = {
        "gt": lambda x: x > value,
        "lt": lambda x: x < value,
        "eq": lambda x: x == value,
        "gte": lambda x: x >= value,
        "lte": lambda x: x <= value,
    }
    
    if operator not in ops:
        raise gorunpy.ValidationError(f"Invalid operator: {operator}")
    
    filtered = df[ops[operator](df[column])]
    return filtered.to_dict(orient="records")


@gorunpy.export
def aggregate_csv(
    file_path: str,
    group_by: str,
    agg_column: str,
    agg_func: str,
) -> dict[str, float]:
    """Aggregate CSV data. Functions: sum, mean, count, min, max."""
    import pandas as pd

    df = pd.read_csv(file_path)
    
    if agg_func not in ["sum", "mean", "count", "min", "max"]:
        raise gorunpy.ValidationError(f"Invalid function: {agg_func}")
    
    result = df.groupby(group_by)[agg_column].agg(agg_func)
    return result.to_dict()


@gorunpy.export
def analyze_with_aggregation(
    file_path: str,
    group_by: str,
    agg_column: str,
) -> dict[str, float]:
    """Get average of agg_column grouped by group_by column."""
    import pandas as pd

    df = pd.read_csv(file_path)
    return df.groupby(group_by)[agg_column].mean().to_dict()

