# Markdown Compiler

## Design Overview

This package implements a small compiler capable of transforming Markdown source into HTML through a series of explicit, layered stages. Each stage performs one transformation and hands off a well-defined intermediate representation to the next. Scanners segment input into units; builders assemble those unites into structure; lowering converts structure into a semantic HTML tree; rendering serializes that tree. No stage reinterprets raw input that belongs to another layer.

The goal of this architecture is predicatability, extensibility, and testability. New Markdown features are added by expanding the scanner vocabulary and introducing new build rules, not by modifying the overall pipeline. The shape of the compiler remains stable while individual slices grow in capability. This keeps the system mechanically sound and testable, locally comprehensible, and resistant to the fragility of regex-driven or ad-hoc parsing approaches.

## Compilation Pipeline

- Markdown (string)
- Block Parse
    - Block Scan: outputs `[]Line`
    - Block Build: outputs `ir.Document`
- Inline Parse
    - Inline Scan: outputs `[]Event`
    - Inline Build: outputs `[]ast.Inline`
- IR lowering
    - Converts `ir.Document` into `ast.Document`
    - Calls Inline Parse during lowering
- AST Render
    - Converts `ast.Document` to `html.Node` tree
- HTML Render
    - Serializes `html.Tree` to string output
    - `html.Write` writes to the provide `io.Writer`
    - `html.Render` returns a serialized string directly

## Entry Points

- `CompileAndRender(md string) (string, error)`
    - Full pipeline to HTML string
- `Compile(md string) (html.Node, error)`
    - Returns the HTML node tree for templ integration

## Markdown Rules: CommonMark-ish

### Headers

A line is recognized as a header if and only if the following is true:
- **Indentation**: The line may begin with 0-3 spaces. Tabs do not count as indentation.
- **Marker run**: After leading spaces, there is a a run of 1-6 `#` characters.
- **Delimiter**: The marker run is followed by at least one delimiter character: space or tab.
- **Content**: Header content is the rest of the line after consuming all consecutive spaces or tabs following the marker run.
- **Normalization**: Trailing whitespace is trimmed from the content.
- **Termination**: The header is a single line. A newline ends it.

