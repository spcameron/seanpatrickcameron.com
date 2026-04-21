# scribe

## Quick Start

`scribe` is a Markdown compiler that produces HTML.

It exposes a minimal API: 

* compile Markdown into a renderable document
* render that document to HTML

### Installation

```sh
go get github.com/spcameron/scribe
```

### Basic Usage

```go
html, err := scribe.HTML(md)
if err != nil {
  // handle error
}
```

### Rendering to an `io.Writer`

```go
doc, err := scribe.Compile(md)
if err != nil {
  // handle error
}

if err := doc.Write(w); err != nil {
  // handle error
}
```

### Using with templ

```go
func MarkdownHTML(doc scribe.Document) templ.Component {
  return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
    if doc == nil {
      return nil
    }
    return doc.Write(w)
  })
}
```

### Notes

* Output is HTML.
* The compiler follows CommonMark-style rules with some intentional departures and simplifications.
* The API is small by design; internal structure is not exposed.

---

## Design Overview

This package implements a staged compiler that transforms Markdown source into HTML.

The compiler operates over a single immutable `Source` buffer. All structural elements—lines, blocks, and inline nodes—are represented as `ByteSpan` coordinates into that source. Text is never copied or passed between stages; it is materialized only at the HTML boundary.

Each stage performs a single transformation and hands off a representation with a narrower and more semantic shape than the one before it:

* scanning segments input into structural units
* parsing assembles those units into block-level intermediate representation (IR)
* lowering converts block IR into semantic AST and invokes inline parsing over content spans
* code generation produces an HTML node tree
* emission serializes that tree to an `io.Writer` or string

No stage reinterprets raw input that belongs to another layer.

As a consequence:

* the source buffer is the single authority for all text
* all intermediate structures are span-based projections of that buffer
* coordinate semantics remain stable across all stages
* normalization and interpretation occur exactly once, at the appropriate layer

The shape of the compiler is stable. New Markdown features expand rule vocabularies, not the architecture itself.

---

## Compilation Pipeline

### Pipeline Overview

* Markdown (`string`)

* `source.Source`

  * immutable buffer
  * line index

* Block Parse

  * block scan → `[]Line`
  * block build → `ir.Document`

* Lowering

  * `ir.Document` → `ast.Document`
  * invokes inline parsing for content-bearing spans

* Inline Parse (during lowering)

  * inline scan → `[]Token`
  * build → mutable item list + delimiter stack
  * lower → `[]ast.Inline`

* Code Generation

  * `ast.Document` → renderable document

* HTML Emission

  * document writes to `io.Writer`
  * helper functions may render to string

Inline parsing is not a separate compilation stage; it is a transformation applied during lowering to spans that carry inline content.

---

## Representation Boundaries

The compiler is organized around four representation layers:

* `source.Source`: immutable input buffer with span utilities and line/column mapping
* `ir.Document`: block-level intermediate representation; structural only, span-based
* `ast.Document`: semantic representation used for code generation; still span-based
* `markdown.Document`: target-language representation used for HTML serialization

Only the HTML layer materializes concrete output text. All earlier layers operate by preserving and transforming coordinates into the original source.

---

## Entry Points

* `Compile(md string) (Document, error)`: executes the full pipeline and returns a renderable document
* `HTML(md string) (string, error)`: executes the full pipeline and returns serialized HTML

The returned `Document` writes HTML directly to an `io.Writer`.

---

## Architectural Decisions

### 1. Immutable Source

All stages operate on a single `Source`. Structural elements carry spans into that source rather than copying substrings.

This eliminates string drift, centralizes normalization, and ensures that every node can be traced back to an exact byte range in the original input.

### 2. IR vs AST Separation

The compiler distinguishes between:

* **Block IR**: structural parsing of block boundaries and hierarchy
* **AST**: semantic representation suitable for code generation

Block parsing determines *where* structure exists. Lowering determines *what it means*. This separation keeps rule logic local and prevents semantic concerns from leaking into scanning or block construction.

### 3. Lowering as a First-Class Stage

Lowering is a semantic transformation stage.

It converts block IR into AST nodes, invokes inline parsing for content-bearing spans, and normalizes distinct surface forms into unified semantic constructs. Lowering preserves span identity across these transformations and does not perform rendering.

### 4. Code Generation vs. Emission

The compiler distinguishes between:

* **code generation**: AST → `html.Node` tree
* **emission**: `html.Node` → serialized output

Output structure is materialized during code generation. Output text is produced during emission.

### 5. Scanner Discipline

Scanners are mechanical. They do not construct semantic nodes or partially interpret structure.

Their responsibility is limited to segmenting input into span-referenced units. Interpretation is deferred to build and lowering stages, where sufficient context exists to make correct decisions.

### 6. Inline Working Model

Inline parsing is built around two mutable structures: an item list and a delimiter stack.

The parser consumes tokens in a single pass, initially treating recognized syntax as provisional text. As closing conditions are encountered, regions of the item list are rewritten in place into structured nodes (emphasis, links, images, code spans).

This approach avoids premature interpretation while allowing nested and overlapping constructs to be resolved incrementally. Unmatched candidates remain literal text.

--- 

## Block Parsing Model

Block parsing in this compiler is intentionally conservative and structurally driven. Rather than re-specifying CommonMark in full, this section describes the subset of constructs supported and the principles used to recognize them.

The parser operates over line spans derived from a normalized source buffer and applies a fixed set of block rules in precedence order. Each rule consumes the maximal sequence of lines that form a valid construct. Container blocks transform their input (e.g., by stripping markers or adjusting indentation) and recursively invoke the same parser, ensuring that all structure is derived through the same mechanism.

The goal is not full CommonMark compliance, but a predictable and internally consistent system that aligns with CommonMark where practical and diverges where simplicity or clarity is preferred.

### Indentation

Indentation is measured in visual columns, with tabs advancing to the next multiple of four. Only leading whitespace contributes to indentation; content is not rewritten or expanded.

This model is used strictly for structural recognition. It determines whether a line participates in a construct but does not alter the underlying source text.

### Headers

Both ATX (`#`) and Setext (`===`, `---`) headers are supported and normalized into a single header representation.

ATX headers are recognized when a line begins (after up to three columns of indentation) with a run of one to six unescaped `#` characters. The opening run must be followed by either whitespace (space or tab) or end-of-line.

The remainder of the line forms the heading field. Content is derived from this field by:

* removing an optional closing marker run: a trailing sequence of one or more unescaped `#` characters that is preceded by whitespace and followed only by optional whitespace
* trimming leading and trailing spaces or tabs

As a result, ATX headings may be empty (e.g., `#,` `##`), and closing marker runs are not required to match the length of the opening run.

Setext headers are recognized as a paragraph immediately followed by an underline line consisting entirely of `=` or `-` characters (aside from indentation and trailing whitespace). The underline determines the level, and the paragraph provides the content.

In both cases, the original syntactic form is discarded during lowering; downstream stages operate only on the semantic header node.

### Thematic Breaks

Thematic breaks are recognized as lines consisting of at least three identical marker characters (`-`, `*`, or `_`), optionally separated by spaces or tabs. Aside from indentation and inter-marker whitespace, no other characters are permitted.

When ambiguity arises between a thematic break and a Setext underline (notably with `---`), the Setext interpretation takes precedence if a valid paragraph precedes the line.

### Block Quotes

Block quotes are formed from lines beginning with one or more `>` markers (after indentation), each optionally followed by a single space or tab. The number of markers determines the nesting depth.

Each level of quoting is constructed by stripping one marker layer and recursively parsing the resulting content. This produces structurally nested block quote nodes rather than a flat representation.

This implementation does not support lazy continuation. Every line within a block quote must carry an explicit `>` marker, including blank lines. This constraint simplifies parsing and preserves a direct correspondence between source lines and structure.

### Lists

Ordered and unordered lists are supported as container blocks whose structure is determined by marker recognition and indentation.

Unordered lists use `-`, `*`, or `+` markers. Ordered lists use a sequence of digits followed by `.` or `)`. In both cases, the marker must be followed by whitespace, and the indentation of the marker establishes the list’s structural baseline.

A list item consists of the marker line and any subsequent lines whose indentation meets or exceeds the item’s content baseline. These continuation lines are parsed recursively as block content.

Blank lines within items are permitted and influence whether the list is rendered as tight or loose. Nested lists emerge naturally when a continuation line itself satisfies a list marker rule at a deeper indentation level.

The parser does not enforce sequential numbering for ordered lists. If the first item does not begin at `1`, the resulting HTML includes a `start` attribute.

### Code Blocks

Code blocks are treated as literal regions and are never subject to inline parsing.

Two forms are supported:

Indented code blocks arise from lines with at least four columns of indentation. The first four columns are removed during normalization, and any additional indentation is preserved as content.

Fenced code blocks are introduced by runs of backticks or tildes (at least three). The closing fence must use the same marker and meet or exceed the opening length. An optional info string may follow the opening fence; its first token is interpreted as a language identifier during rendering.

In both forms, line boundaries are preserved exactly, and the resulting content is emitted as literal text within `<pre><code>`.

### HTML Blocks

HTML blocks provide a passthrough mechanism for raw HTML. When a recognized HTML opener appears at the start of a line (after indentation), the parser suspends Markdown interpretation and treats the content as literal until a corresponding termination condition is met.

Supported forms include comments, declarations, CDATA sections, processing instructions, and a restricted set of block-level tags.

Delimiter-terminated forms (e.g., `<!-- ... -->`) continue until their closing sequence is found. Named-tag blocks continue until a blank line is encountered.

Within an HTML block, no inline parsing or normalization occurs. The content is emitted verbatim.

### Paragraphs

A paragraph consists of one or more consecutive non-blank lines that do not form another block construct.

Paragraphs serve as the default block and the primary carrier of inline content. During lowering, line boundaries are interpreted to produce either soft breaks (rendered as spaces) or hard breaks (rendered as `<br>`), depending on trailing whitespace or escape markers.

### Deviations from CommonMark

This implementation intentionally diverges from CommonMark in a small number of areas:

* **No lazy continuation**: Block quotes require explicit markers on every line. This avoids implicit structure and simplifies parsing.
* **Restricted HTML block recognition**: Only a subset of block-level tags is recognized to prevent accidental capture of inline HTML.
* **Inline newline handling**: Inline parsing does not admit arbitrary newlines; line structure is resolved at the block level.
* **Simplified ambiguity resolution**: In edge cases, precedence rules favor structural clarity over exhaustive spec compliance.

These deviations are chosen to preserve a clear separation between structural parsing and inline semantics, and to keep the parser mechanically predictable.

--- 

## Inline Parsing Model

Inline parsing is implemented as a single forward pass over a token stream, backed by two mutable structures: an item list and a delimiter stack.

Unlike the block layer, which is purely structural, inline parsing must resolve overlapping and nested constructs whose interpretation depends on surrounding context. The parser therefore operates incrementally, treating all input as provisional text and selectively upgrading regions into semantic nodes as patterns become valid.

### Overview

Inline parsing proceeds in three conceptual steps:

1. **Scan**: Convert a content span into a stream of lexical tokens.
2. **Build**: Walk the token stream once, constructing a working representation.
3. **Lower**: Convert the working representation into `ast.Inline` nodes.

Only the final step produces semantic nodes. All prior stages operate on span-referenced structures.

### Scanning

The scanner performs a linear pass over the source slice and emits tokens representing:

* delimiter runs (`*`, `_`, `` ` ``)
* structural markers (`[`, `]`, `(`, `)`, `<`, `>`)
* escape markers (`\`)
* composite forms (`![`)
* plain text

Tokens are span-based and do not interpret meaning. In particular:

* delimiter runs are emitted as single tokens with width
* no attempt is made to classify tokens as “opening” or “closing”
* no structure is constructed during scanning

The scanner is mechanical. It segments input but does not participate in parsing decisions.

### Working Representation

The Build phase maintains two coordinated structures:

* **ItemList**: a doubly-linked list of `ItemRecord`s representing inline content
* **DelimiterList**: a stack of delimiter records referencing items within the list

Each item initially represents literal text (via its span). As parsing progresses, items may be transformed in place into structured nodes (e.g., emphasis, links, code spans) while preserving their original span boundaries.

This design keeps all intermediate state anchored to the original source while allowing localized structural rewrites.

### Single-Pass Construction

The token stream is consumed exactly once.

For each token, the parser performs one of the following actions:

* append a literal text item
* append a provisional delimiter (and record it in the delimiter stack)
* attempt to resolve a construct immediately (e.g., code spans, autolinks, inline HTML)
* attempt to close a previously opened construct (e.g., links or images)

Crucially, delimiter-based constructs (emphasis, links, images) are not resolved eagerly in all cases. Instead, the parser records sufficient metadata to allow later resolution when a closing condition is encountered.

### Delimiter Handling

Delimiter runs (`*`, `_`) are inserted into the item list as plain text and recorded in the delimiter stack with metadata describing:

* delimiter kind (asterisk or underscore)
* run length
* whether the run may open and/or close (derived from flanking conditions)
* a reference to the corresponding item in the item list

At this point, delimiter runs carry no structure. They are indistinguishable from literal text except for the presence of a corresponding delimiter record.

Resolution is triggered only when a delimiter capable of closing is encountered.

The parser then walks backward through the delimiter stack to locate a compatible opener. Compatibility is determined by:

* matching delimiter kind
* opener `CanOpen` and closer `CanClose` flags
* modulo-3 constraints on run lengths
* additional restrictions for underscores (intraword behavior)

If no matching opener is found, the delimiter remains literal. In some cases it is removed from the stack if it can no longer participate in future matches.

When a matching opener is found, the parser performs a localized rewrite:

1. **Determine strength**
   One or two delimiter characters are consumed from each side depending on run lengths.

2. **Adjust delimiter runs**
   The opener and closer item spans are shortened to reflect consumed characters. If a run is fully consumed, its item and delimiter record are removed.

3. **Extract children**
   All items strictly between the opener and closer are detached from the item list as a contiguous range.

4. **Construct new item**
   A new item is created (`Emphasis` or `Strong`) with:

   * an original span covering both delimiters
   * a live span covering only the enclosed content
   * the detached items as its children

5. **Reinsert structure**
   The new item is inserted at the opener position, preserving list order.

6. **Clean delimiter state**
   All delimiter records between the opener and closer are removed. The parser then resumes from a stable position in the delimiter stack.

Because resolution operates directly on the item list, no index-based rewriting is required. The structure evolves through local mutations rather than global passes.

Unmatched delimiter runs remain as text. No backtracking or re-scanning is performed.

### Code Spans

Backtick runs are resolved immediately.

Upon encountering a backtick token, the parser scans forward for a matching run of equal length. If found:

* the enclosed span is extracted
* leading/trailing space normalization is applied
* a code span item is created

If no matching closer is found, the run is treated as literal text.

No delimiter stack interaction is required for code spans.

### Links and Images

Bracket delimiters (`[` and `![`) are pushed onto the delimiter stack as provisional openers.

When a closing `]` is encountered, the parser searches backward for a matching active opener. If found, it attempts to parse an inline link tail beginning at the next position.

If a valid tail is parsed:

* emphasis resolution is performed within the bracketed region
* the enclosed items are detached and assigned as children
* the opener item is transformed into a link or image node
* prior link delimiters are deactivated (to prevent nested links)

If inline tail parsing fails, the parser attempts to validate and lookup a full reference, collapsed reference, or shortcut reference form.

If all reference parse attempts fail, the closing bracket is emitted as literal text and the opener remains inactive.

### Autolinks and Inline HTML

Angle-bracket sequences are handled opportunistically.

When `<` is encountered, the parser first checks for:

* URI autolinks
* email autolinks

If those fail, it attempts to match inline HTML constructs using a byte-level scan. Valid constructs are emitted as raw HTML items. Otherwise, the `<` is treated as literal text.

These constructs are resolved immediately and do not interact with the delimiter stack.

### Escapes

Backslash escapes are resolved contextually based on the following token:

* some tokens are **literalized** (treated as plain text)
* some are **decomposed** (e.g., `\![` becomes `!` + `[`)
* others leave the backslash intact

Escape handling occurs during the Build pass and affects how subsequent tokens are interpreted.

### Final Emphasis Resolution

After the full token stream has been consumed, any remaining delimiter runs are processed to resolve emphasis across the entire item list.

Unresolved delimiters are discarded, leaving their corresponding items as literal text.

### Lowering

The final step walks the item list and converts each item into an `ast.Inline` node.

This includes:

* text nodes (span-backed)
* code spans
* emphasis and strong nodes (with recursively lowered children)
* links and images
* autolinks
* raw HTML segments

Lowering is purely structural. It does not reinterpret spans or perform additional parsing.

### Design Rationale

This model avoids premature interpretation and keeps parsing decisions local:

* Scanning is purely lexical.
* Structure is introduced only when sufficient context exists.
* All intermediate state remains span-based and mutable.
* Resolution operates directly on a stable item list, avoiding index invalidation.

The result is a parser that can handle nested and overlapping inline constructs while preserving a direct correspondence to the original source.

--- 

## Reference Definitions and Reference Links

Reference-style links and images are supported through a two-phase mechanism: definitions are collected during block parsing, and references are resolved during inline parsing.

This design preserves the separation between structural parsing and semantic resolution while allowing reference definitions to be declared independently of their use.

### Reference Definitions

A reference definitions has the form:

```
[label]: destination "optional title"
```

and is recognized as a block-level construct.

A valid definition:

* begins with a bracketed label (`[label]`)
* is followed immediately by a colon (`:`)
* includes a link destination
* may include an optional title, separated from the destination by whitespace
* must occupy a single line (multi-line forms are not supported)

If a definition is valid, it is not emitted as a block. Instead, it is recorded in a document-level map keyed by a normalized form of the label.

Normalization:

* is case-insensitive
* collapses consecutive whitespace into a single space
* ignores leading and trailing whitespace
* resolves escaped sequences before comparison

If multiple definitions normalize to the same key, the first definition wins and subsequent definitions are ignored.

Reference definitions do not interrupt paragraphs. If a line resembles a definition but appears within paragraph content, it is treated as literal text.

### Reference Resolution

Reference links and images are resolved during inline parsing when a closing bracket (`]`) is encountered.

The parser attempts to interpret the bracketed construct in the following order:

1. inline link/image
2. full reference
3. collapsed reference
4. shortcut reference

The first successful interpretation is accepted. If no interpretation succeeds, the brackets are treated as literal text.

### Reference Link Forms

Three reference forms are supported:

* Full Reference: `[label][ref]`
* Collapsed Reference: `[label][]`
* Shortcut Reference: `[label]`

Image references follow the same forms, prefixed by `!`.

### Resolution Semantics

When resolving a reference:

* the lookup label is validated and normalized using the same rules as definitions
* the normalized key is used to query the document's definition map
* if a matching definition is found, its destination and title are applied
* if no matching definition exists, the construct is treated as literal text

The visible label content is parsed as inline content and becomes the children of the resulting link or image node.

---

## Diagnostics

Because all nodes carry spans into a single `Source`, the compiler can produce precise, location-aware diagnostics. That being said, the program does not currently take advantage of this capability.

`Source` provides:
* `LineCol(BytePos) (line, column)`
* Span slicing with bounds validation

Diagnostics may be emitted during:
* Block parsing
* Inline parsing
* Lowering
* Code generation

Example diagnostic output:

```
invalid header delimiter at 3:7
  |
3 | ###Header
  |       ^
```

---

## Extending the Compiler

New Markdown features are added by expanding rule sets within existing layers:

1. Determine whether the feature is block-level or inline-level.
2. Add scanner vocabulary only if new delimiters are required.
3. Introduce a build rule in the appropriate package.
4. Lower into new AST node types as needed.
5. Extend code generation to produce HTML.

*The shape of the compiler is stable.*

---

## Philosophy

This project treats Markdown as a small language and HTML as its target output format.

The design mirrors conventional compiler structure:

* Immutable source buffer
* Span-based structural nodes
* Staged transformations
* Explicit lowering
* Target-language code generation

Block constructs are parsed according to clear structural rules. Surface syntax is normalized early: distinct syntactic forms that represent the same semantic construct (e.g., ATX headers and Setext headers) are lowered into a single `Header` IR node. Downstream stages operate only on semantic structure, not original delimiter forms.

The system remains mechanically predictable and extensible while preserving precise coordinate semantics throughout.
