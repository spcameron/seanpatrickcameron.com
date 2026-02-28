# Markdown Compiler

## Design Overview

This package implements a small, staged compiler that transforms Markdown source into HTML.

The compiler operates on a single immutable `Source` buffer. All structural elements (lines, blocks, inline nodes) are represented as `ByteSpan` coordinates into that source. Text is not copied or passed between stages, and strings are derived from spans only at the final HTML emission boundary.

Each stage performs one transformation and hands off a well-defined representation to the next:

- **Scanning** segments input into structural units.
- **Parsing** assembles those units into a block-level intermediate representation (IR).
- **Lowering** transforms block IR into a semantic abstract syntax tree (AST).
- **Code generation** converts the AST into an HTML node tree.
- **Emission** serializes the HTML tree into bytes or strings.

No stage reinterprets raw input that belongs to another layer.

This guarantees a single source of truth for content, stable byte-coordinate semantics across all stages, precise span-based diagnostics, and no string drift or duplicated normalization logic.

*The shape of the compiler is stable. New Markdown features expand rule vocabularies, not the architecture itself.*

## Compilation Pipeline

### Pipeline Overview

- Markdown (string)
- Source (immutable buffer + line index)
- Block Parse
    - Block Scan: outputs `[]Line` (spans)
    - Block Build: outputs `ir.Document` (block IR)
- Lowering
    - ir.Document becomes ast.Document
    - Performs inline parsing per content span
- Inline Parse (invoked during lowering)
    - Inline Scan: outputs `[]Event` (spans)
    - Inline Build: outputs `[]ast.Inline`
- Code Generation
    - ast.Document becomes html.Node tree
- HTML Emission
    - Serializes `html.Tree` to string output or io.Writer
    - `html.Write` writes to a provided io.Writer
    - `html.Render` returns a serialized string directly

### Representation Boundaries

- `source.Source`: Immutable input buffer with span utilities and line/column mapping.
- `ir.Document`: Block-level intermediate representation. Structual only and span-based.
- `ast.Document`: Semantic tree suitable for code generation. Span-based.
- `html.Node`: Target-language representation (HTML tree). String-backed.

Only the HTML layer operates on concrete strings.

## Entry Points

- `HTML(md string) (string, error)`: Executes the full pipeline and returns serialized HTML. 
- `Tree(md string) (html.Node, error)`: Returns the HTML node tree (useful for templ integration or further processing).

## Architectural Decisions

### 1. Immutable Source

All parsing and lowering operate on a single `Source`. Structural elements carry spans into the source rather than copying substrings. This enables zero string passing across seams, accurate diagnostics, and consistent normalization rules.

### 2. IR vs AST Separation

The compiler distinguishes between **Block IR** (structural parsing) and **AST** (semantic representaiton). Block parsing occurs first and determines structural boundaries, while lowering to AST determines semeantic meaning. This separation keeps rule logic local and prevents semantic concerns from leaking into scanning.

### 3. Lowering as a First-Class Stage

Mentioned above, lowering is a structural transformation pass and performs real transformations. Lowering converts block IR into semantic AST nodes, invokes line parsing per content span, and preserves spans across transformations. Lowering is not rendering.

### 4. Code Generation vs. Emission

The compiler distinguishes between **code generation** (AST -> `html.Node` tree) and **emission** (`html.Node` -> serialized output). Text materialization occurs exactly once, during code generation.

### 5. Scanner Discipline

Scanners are mechanical, meaning they do *not* interpret structure or create semantic nodes. Their only responsibility is to segment input into span-referenced units. All interpretation occurs in build or lowering rules.

## Markdown Rules (CommonMark-ish)

### Block Elements

#### ATX Headers (`#`)

A header is a block used to create titles, subtitles, or otherwise structure content.

A line is recognized as a header if and only if the following is true:

- **Indentation**: The line begins with 0-3 spaces. Tabs do not count as indentation.
- **Marker Run**: After leading spaces, there is a run of 1-6 `#` characters.
- **Delimiter**: The marker run is followed by at least one delimiter character: space or tab.
- **Content**: Header content is defined as the rest of the line after consuming all consecutive spaces or tabs following the marker run.
- **Normalization**: Trailing whitespace is trimmed from the content.
- **Termination**: The header is a single line. A newline ends it.

The Header IR node stores both the full line span and the content span (excludes marker and trimmed whitespace).

Headers are rendered as `<h1></h1>` ... `<h6></h6>` in HTML.

#### Thematic Breaks (`---`, `***`, `___`)

A thematic break is a leaf block representing a horizontal rule.

A line is recognized as a thematic break if all of the following are true:

- **Indentation**: The line begins with 0-3 spaces. Tabs do not count as indentation.
- **Marker Character**: The first non-indent character is one of `-` `*` or `_`.
- **Marker Count**: The line contains at least three marker characters, and all marker characters must be identical.
- **Separator Rules**: Marker characters may be separated by any number of spaces or tabs, but no other characters are permitted.
- **Line Purity**: Aside from indentation and optional inter-marker whitespace, the line must contain only the chosen marker. Trailing whitespace is permitted.

A thematic break consumes exactly one line, and may interrupt paragraphs. Setext heading underlines are not supported, so `---` is always interpreted as a thematic break when it satisfies the rules above.

Breaks are rendered as `<hr>` in HTML.

#### Block Quotes (`>`)

A block quote is a container block used to quote or otherwise offset content. Block quotes may contain any other block elements supported by the compiler, including paragraphs, headers, thematic breaks, lists, and other (nested) block quotes.

A line is recognized as a part of a block quote if and only if the following is true:

- **Indentation**: The line begins with 0-3 spaces. Tabs do not count as indentation.
- **Marker Unit**: After indentation, the line contains at least one quote marker unit. A quote marker unit is:
    - a single `>` character, followed by
    - an optional single delimiter character, space or tab.
- **Marker Run**: The quote marker run is one or more consecutive quote marker units. The nesting depth of the line is the number of `>` characters consumer by the marker run.
- **Content**: Quote line content is defined as the remainder of the line after consuming indentation and the full marker run.
- **Whitespace Preservation**: Only one delimiter character (space or tab) may be consumed after each `>` marker. Any additional spaces or tabs are preserved as contnet. Tabs are not expanded.

A block quote consists of a maximal contiguous sequence of quote-eligible lines. Lazy continuation is not supported; every physical line in a block quote must bear a leading `>` marker.

Blank lines inside a block quote must also include a `>` marker. Such lines are treated as blank lines within the quoted content and may separate paragraphs or other blocks.

Multiple consecutive `>` markers indicate nested block quotes. Each nesting layer is parsed by stripping exactly one leading `>` marker (and optional delimiter) from each line in the block and recursively invoking block parsing on the resulting content. This process yields structurally nested `BlockQuote` nodes in the IR.

Block quotes are rendered as `<blockquote>...</blockquote>` in HTML.


### Inline Elements

#### Paragraphs

A paragraph consists of one or more consecutive non-blank lines that do not begin another block construct.

Paragraph IR stores:
- A span covering all lines
- Individual line spans (used during lowering)

Lowering inserts break semantics:
- A line ending with two spaces or `\\` produces a `HardBreak`
- Otherwise, inter-line boundaries produce `SoftBreak`

Breaks are represented explicitly in the AST.

A paragraph is rendered as `<p></p>`, and a break is rendered as `<br>` in HTML.

## Diagnostics

Because all nodes carry spans into a single `Source`, the compiler can produce precise, location-aware diagnostics.

`Source` provides:
- `LineCol(BytePos) (line, column)`
- Span slicing with bounds validation

Diagnostics may be emitted during:
- Block parsing
- Inline parsing
- Lowering
- Code generation

Example diagnostic output:

```
invalid header delimiter at 3:7
  |
3 | ###Header
  |       ^
```

## Extending the Compiler

New Markdown features are added by expanding rule sets within existing layers:

1. Determine whether the feature is block-level or inline-level.
2. Add scanner vocabulary only if new delimiters are required.
3. Introduce a build rule in the appropriate package.
4. Lower into new AST node types as needed.
5. Extend code generation to produce HTML.

*The shape of the compiler is stable.*

## Philosophy

This project treats Markdown as a small language and HTML as its target language.

The design mirrors conventional compiler structure:
- Immutable source buffer
- Span-based structural nodes
- Staged transformations
- Explicit lowering
- Target-language code generation

The system remains mechanically predictable and extensible while preserving precise coordinate semantics throughout.
