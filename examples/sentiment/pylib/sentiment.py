"""Sentiment analysis using HuggingFace Transformers.

Uses lightweight models suitable for bundling.
"""

import gorunpy
from typing import Optional

# Cache the pipeline to avoid reloading on each call
_pipeline = None


def _get_pipeline():
    global _pipeline
    if _pipeline is None:
        from transformers import pipeline
        _pipeline = pipeline(
            "sentiment-analysis",
            model="distilbert-base-uncased-finetuned-sst-2-english",
        )
    return _pipeline


@gorunpy.export
def analyze(text: str) -> dict:
    """Analyze sentiment of text. Returns label and confidence score."""
    pipe = _get_pipeline()
    result = pipe(text)[0]
    return {
        "label": result["label"],
        "score": result["score"],
    }


@gorunpy.export
def analyze_batch(texts: list[str]) -> list[dict]:
    """Analyze sentiment of multiple texts."""
    pipe = _get_pipeline()
    results = pipe(texts)
    return [{"label": r["label"], "score": r["score"]} for r in results]


@gorunpy.export
def classify(text: str, labels: list[str]) -> dict[str, float]:
    """Zero-shot classification with custom labels."""
    from transformers import pipeline
    
    classifier = pipeline(
        "zero-shot-classification",
        model="facebook/bart-large-mnli",
    )
    result = classifier(text, labels)
    return dict(zip(result["labels"], result["scores"]))

