# Pandas Example

Data analysis and CSV processing using [pandas](https://pandas.pydata.org/).

## Setup

```bash
cd examples/pandas

python -m venv .venv  # or python3 -m venv .venv
source .venv/bin/activate
pip install "gorunpy[build]" pandas jinja2

gorunpy
```

## Run

```bash
go run .              # uses included data.csv
go run . mydata.csv   # or specify your own
```

## Exported Functions

| Function | Description |
|----------|-------------|
| `ReadCsvStats(path)` | Get descriptive statistics for all columns |
| `CsvToJson(path)` | Convert CSV to JSON array |
| `FilterCsv(path, col, op, val)` | Filter rows (gt, lt, eq, gte, lte) |
| `AggregateCsv(path, groupBy, col, func)` | Group and aggregate (sum, mean, count, min, max) |

