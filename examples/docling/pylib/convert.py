"""Document conversion using Docling.

Docling converts PDFs, DOCX, PPTX, and other documents to structured formats.
https://github.com/DS4SD/docling
"""

import gorunpy
from typing import Optional
from pathlib import Path


@gorunpy.export
def pdf_to_markdown(file_path: str) -> str:
    """Convert a PDF file to Markdown format."""
    from docling.document_converter import DocumentConverter

    converter = DocumentConverter()
    result = converter.convert(file_path)
    return result.document.export_to_markdown()


@gorunpy.export
def pdf_to_text(file_path: str) -> str:
    """Convert a PDF file to plain text."""
    from docling.document_converter import DocumentConverter

    converter = DocumentConverter()
    result = converter.convert(file_path)
    return result.document.export_to_text()


@gorunpy.export
def extract_tables(file_path: str) -> list[dict]:
    """Extract tables from a document as list of dicts."""
    from docling.document_converter import DocumentConverter

    converter = DocumentConverter()
    result = converter.convert(file_path)
    
    tables = []
    for table in result.document.tables:
        tables.append({
            "rows": table.num_rows,
            "cols": table.num_cols,
            "data": table.export_to_dataframe().to_dict(orient="records"),
        })
    return tables

