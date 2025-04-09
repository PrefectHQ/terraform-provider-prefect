# Scripts

## `compare-and-output-markdown.py`

A python script to query:

- latest Prefect Cloud OpenAPI schema
- latest Terraform Provider for Prefect

and output a report of yet-to-be-implemented Prefect Cloud resources.

### Requirements

- [Python](https://www.python.org/downloads/) 3.12+
- [uv](https://docs.astral.sh/uv/getting-started/installation/)

### Usage

```bash
➜ uv run ./scripts/compare-and-output-markdown.py
```
