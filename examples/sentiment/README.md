# Sentiment Analysis Example

Text sentiment analysis using [HuggingFace Transformers](https://huggingface.co/docs/transformers).

## Setup

```bash
cd examples/sentiment

python -m venv .venv  # or python3 -m venv .venv
source .venv/bin/activate
pip install "gorunpy[build]" transformers torch

gorunpy
```

## Run

```bash
go run .
```

## Exported Functions

| Function | Description |
|----------|-------------|
| `Analyze(text)` | Get sentiment (POSITIVE/NEGATIVE) with confidence |
| `AnalyzeBatch(texts)` | Analyze multiple texts at once |
| `Classify(text, labels)` | Zero-shot classification with custom labels |

## Notes

- First run downloads models (~300MB for DistilBERT)
- Uses `distilbert-base-uncased-finetuned-sst-2-english` (lightweight)
- For GPU: `pip install torch --index-url https://download.pytorch.org/whl/cu118`

## Example Output

```
Text: "I love this product! It's amazing!"
  → POSITIVE (99.87%)

Text: "This is terrible, worst purchase ever."
  → NEGATIVE (99.91%)
```

