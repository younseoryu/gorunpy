# Docling Example

Convert PDFs and documents to Markdown/text using [Docling](https://github.com/DS4SD/docling).

## Setup

```bash
cd examples/docling

# Create venv and install dependencies
python -m venv .venv  # or python3 -m venv .venv
source .venv/bin/activate
pip install "gorunpy[build]" docling

# Build and generate Go client
gorunpy
```

## Run

```bash
go run . sample.pdf
```

## Exported Functions

| Function | Description |
|----------|-------------|
| `PdfToMarkdown(path)` | Convert PDF to Markdown |
| `PdfToText(path)` | Convert PDF to plain text |
| `ExtractTables(path)` | Extract tables as structured data |

## Notes

- First run downloads Docling models (~1GB)
- Supports PDF, DOCX, PPTX, HTML, and more
- Heavy dependency - binary will be large (~500MB+)

