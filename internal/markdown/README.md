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
    - `ir.Document` becomes `ast.Document`
    - Performs inline parsing per content span
- Inline Parse (invoked during lowering)
    - Inline Scan: outputs `[]Event` (span-based lexical tokens)
    - Gather: builds working items and delimiter records
    - Resolve: pairs compatible delimiters and constructs inline nodes
    - Finalize: emits `[]ast.Inline`
- Code Generation
    - `ast.Document` becomes `html.Node` tree
- HTML Emission
    - Serializes `html.Tree` to string output or io.Writer
    - `html.Write` writes to a provided io.Writer
    - `html.Render` returns a serialized string directly

### Inline Parsing Model

Inline parsing follows a staged pipeline designed to separate lexical recognition from delimiter resolution.

- **Inline Scan**: The scanner walks a content span and emits a sequence of inline events. Events represent either literal text or delimiter runs (such as `*` sequences). Each event carries a `ByteSpan` into the original source.
- **Gather**: Gather converts events into a working stream while recording delimiter metadata. The gather phase performs no semantic interpretation beyond delimiter eligibility classification.
- **Resolve**: Resolve walks the delimiter stack and pairs compatible delimiter runs. This phase constructs nesting inline nodes such as `Em` and `Strong`.
- **Finalize**: Finalize walks the working item stream and materializes the final `ast.Inline` nodes. The result is the finalized inline AST for the parsed span.

### Representation Boundaries

- `source.Source`: Immutable input buffer with span utilities and line/column mapping.
- `ir.Document`: Block-level intermediate representation. Structural only and span-based.
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

The compiler distinguishes between **Block IR** (structural parsing) and **AST** (semantic representation). Block parsing occurs first and determines structural boundaries, while lowering to AST determines semantic meaning. This separation keeps rule logic local and prevents semantic concerns from leaking into scanning.

### 3. Lowering as a First-Class Stage

Mentioned above, lowering is a structural transformation pass and performs real transformations. Lowering converts block IR into semantic AST nodes, invokes line parsing per content span, and preserves spans across transformations. Lowering is not rendering.

### 4. Code Generation vs. Emission

The compiler distinguishes between **code generation** (AST -> `html.Node` tree) and **emission** (`html.Node` -> serialized output). Text materialization occurs exactly once, during code generation.

### 5. Scanner Discipline

Scanners are mechanical, meaning they do *not* interpret structure or create semantic nodes. Their only responsibility is to segment input into span-referenced units. All interpretation occurs in build or lowering rules.

### 6. Delimiter Resolution Model

Inline emphasis and strong emphasis are implemented using a delimiter stack model similar to the one described in the CommonMark specification.

Delimiter runs are first gathered into a working stream along with metadata describing their eligibility to open or close emphasis. A subsequent resolution phase pairs compatible delimiters and constructs inline nodes while preserving index stability within the working stream.

This approach separates delimiter recognition from pairing logic and enables nested constructs to be resolved without modifying earlier parse stages.

## Markdown Rules (CommonMark-ish)

### Indentation Model

Block-level constructs use visual column indentation.

- Indentation is measured in columns.
- A space (` `) advances indentation by one column.
- A tab (`\t`) advances indentation to the next multiple of 4 columns.
- Only leading spaces and tabs contribute to indentation.
- A block rules that reference "0-3 spaces" are interpreted as "0-3 columns".

Indentation is used only for structural recognition. Tabs are not expanded in content.

### Block Elements

#### ATX Headers (`#`)

A header is a block used to create titles, subtitles, or otherwise structure content.

A line is recognized as a header if and only if the following is true:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Marker Run**: After leading spaces, there is a run of 1-6 `#` characters.
- **Delimiter**: The marker run is followed by at least one delimiter character: space or tab.
- **Content**: Header content is defined as the rest of the line after consuming all consecutive spaces or tabs following the marker run.
- **Normalization**: Trailing whitespace is trimmed from the content.
- **Termination**: The header is a single line. A newline ends it.

The Header IR node stores both the full line span and the content span (excludes marker and trimmed whitespace).

Headers are rendered as `<h1></h1>` ... `<h6></h6>` in HTML.

#### Setext Headers (`===`, `---`)

A Setext header is a two-line construct used to create level 1 (`=`) or level 2 (`-`) headings.

A Setext header is recognized if and only if the following is true:

- **Structure**: A paragraph candidate line (or contiguous paragraph run) is immediately followed by a valid underline line.
- **No Blank Separation**: The underline line must appear directly after the paragraph content with no intervening blank line.
- **Indentation**: The line begins with 0-3 columns of indentation.
- **Marker Character**: The first non-indent character of the underline is either `=` or `-`.
- **Marker Run**: The underline line contains a run of one or more identical marker characters.
- **Line Purity**: Aside from indentation and optional trailing spaces or tabs, the underline line must contain only the chosen marker character. Internal spaces between markers are not permitted.
- **Trailing Whitespace**: Trailing spaces or tabs after the marker run are permitted.

The entire preceding paragraph run becomes the header content. The underline line contributes only the level and is not included in the content span.

Setext headers consume exactly two logical components: the paragraph run and the underline line.

Setext headers are lowered into the same `Header` IR node used for ATX headers and rendered as `<h1>` or `<h2>` in HTML.

#### Thematic Breaks (`---`, `***`, `___`)

A thematic break is a leaf block representing a horizontal rule.

A line is recognized as a thematic break if all of the following are true:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Marker Character**: The first non-indent character is one of `-` `*` or `_`.
- **Marker Count**: The line contains at least three marker characters, and all marker characters must be identical.
- **Separator Rules**: Marker characters may be separated by any number of spaces or tabs, but no other characters are permitted.
- **Line Purity**: Aside from indentation and optional inter-marker whitespace, the line must contain only the chosen marker. Trailing whitespace is permitted.

A thematic break consumes exactly one line, and may interrupt paragraphs.

When a line of dashes (`---`) directly follows a paragraph run and satisfies Setext underline rules, it is interpreted as a Setext level 2 header rather than a thematic break. Otherwise, thematic break rules apply.

Breaks are rendered as `<hr>` in HTML.

#### Block Quotes (`>`)

A block quote is a container block used to quote or otherwise offset content. Block quotes may contain any other block elements supported by the compiler, including paragraphs, headers, thematic breaks, lists, and other (nested) block quotes.

A line is recognized as a part of a block quote if and only if the following is true:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Marker Unit**: After indentation, the line contains at least one quote marker unit. A quote marker unit is:
    - a single `>` character, followed by
    - an optional single delimiter character, space or tab.
- **Marker Run**: The quote marker run is one or more consecutive quote marker units. The nesting depth of the line is the number of `>` characters consumer by the marker run.
- **Content**: Quote line content is defined as the remainder of the line after consuming indentation and the full marker run.
- **Whitespace Preservation**: Only one delimiter character (space or tab) may be consumed after each `>` marker. Any additional spaces or tabs are preserved as content.

A block quote consists of a maximal contiguous sequence of quote-eligible lines. Lazy continuation is not supported; every physical line in a block quote must bear a leading `>` marker.

Blank lines inside a block quote must also include a `>` marker. Such lines are treated as blank lines within the quoted content and may separate paragraphs or other blocks.

Multiple consecutive `>` markers indicate nested block quotes. Each nesting layer is parsed by stripping exactly one leading `>` marker (and optional delimiter) from each line in the block and recursively invoking block parsing on the resulting content. This process yields structurally nested `BlockQuote` nodes in the IR.

Block quotes are rendered as `<blockquote>...</blockquote>` in HTML.

#### Lists

Lists are container blocks composed of one or more list items. Each item begins with a marker followed by a delimiter and item content. Lists may contain any other block elements supported by the compiler.

A list begins when a line satisfies either the unordered list marker rules or the ordered list marker rules described below.

- **Indentation**: The line beginning a list item must start with 0-3 columns of indentation. The list indentation column is defined as the visual column where the marker begins. The item content baseline is the visual column immediately after the marker and delimiter run.
- **Item Body Continuation**: After a marker line is consumed, additional lines may belong to the same list item. A subsequent line is treated as a continuation of the current item if the line is not blank and the line's indentation is greater than or equal to the item content baseline column. Continuation lines are included in the list item body and parsed recursively as block content.
- **Blank Lines**: Blank lines inside list items are permitted. Blank lines are tentatively consumed. If the following non-blank line does not satisfy the continuation rule, the blank lines are discarded and parsing resumes outside the list item. Retained blank lines determine whether the list is tight or loose.
- **Item Termination**: A list item ends when a subsequent non-blank line has indentation less than the item content baseline or begins a sibling list item at the same list indentation level.
- **Sibling Items**: After completing an item, the parser attempts to recognize another list item. A sibling item begins when the next line has indentation equal to the list indentation column and satisfies a valid list marker rule.
- **List Termination**: The list ends when the next line has indentation less than the list indentation column or does not form a valid sibling list item.
- **Structure**: Each list item is parsed as a separate block scope. The marker and delimiter are removed and the item body is parsed recursively using the item content baseline as the indentation baseline.
- **Tight vs. Loose Lists**: Lists may be tight or loose depending on whether retained blank lines appear between block elements inside items. Tight lists do not render `<p>` tags around list item text content, while loose lists do.

#### Unordered List Markers (`-`, `*`, `+`)

A line is recognized as the beginning of an unordered list item if:

- after indentation, the line begins with one of the marker characters `-`, `*`, or `+`
- the marker is followed by at least one space or tab delimiter.

#### Ordered List Markers (`1.`, `1)`)

A line is recognized as the beginning of an ordered list item if:

- after indentation, the line begins with a run of one or more digits (`0-9`)
- the digits are followed by either `.` or `)`
- the punctuation is followed by at least one space or tab delimiter

The numeric value does not need to be sequential. The punctuation must remain consistent within a single ordered list. If the first marker value is not `1`, the resulting `<ol>` element is rendered with a `start` attribute.

#### List Nesting

Lists may be nested within other lists or container blocks. A nested list begins when a line inside a list item satisfies a list marker rule and its indentation is greater than or equal to the current item content baseline. Ordered and unordered lists may nest within each other without restriction.

#### Code Blocks

Code blocks represent literal content and are not subject to inline parsing. All Markdown syntax inside a code block is treated as plain text.

Two forms of code blocks are supported: indented code blocks and fenced code blocks. Both forms are lowered into a unified `CodeBlock` AST node and rendered as `<pre><code>...</code></pre>` in HTML.

#### Indented Code Blocks

An indented code block is a leaf block representing literal code content introduced by indentation.

A line is recognized as part of an indented code block if and only if the following is true:

- **Indentation**: The line begins with at least 4 columns of indentation.
- **Leading Whitespace**: Only spaces and tabs may contribute to indentation. Indentation is measure in visual columns according to the indentation model described above.

An indented code block consists of a maximal contiguous sequence of such lines, with the following rules:

- **Blank Lines**: Blank lines may appear inside the block. Blank lines are tentatively consumed and retained only if followed by another indented code block line.
- **Termination**: The block ends when a non-blank line is encountered that has fewer than 4 columns of indentation.
- **Content Preservation**: The original line content is preserved except for the indentation normalization described below.

During lowering, the leading indentation of each payload line is normalized as follows:

- Up to 4 visual columns of leading whitespace are removed from each line.
- Only spaces and tabs are consumed during this process.
- Stripping stops if the next character is not whitespace, even if fewer than 4 columns have been removed.
- Any additional indentation beyond the first 4 columns is preserved as literal code indentation.

Line boundaries are preserved exactly. Each line of code block content is separated by a literal newline character (`\n`) in the resulting payload.

The `IndentedCodeBlock` IR node stores:

- The span covering the entire block
- The spans of each payload line

Lowering converts these spans into normalized inline payload content.

#### Fenced Code Blocks ("```", `~~~`)

A fenced code block is a leaf block introduced by a run of fence markers.

A line is recognized as the opening fence of a fenced code block if and only if the following is true:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Fence Marker**: The first non-indent character is either ``` or `~`.
- **Marker Run**: The marker is repeated at least three times without interruption.
- **Delimiter Whitespace**: Optional spaces or tabs may follow the marker run.
- **Info String**: The remainder of the line is treated as an optional info string.

The closing fence must satisfy the following rules:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Marker Type**: The marker character must match the opening fence marker.
- **Marker Run**: The closing run must contain at least as many markers as the opening fence (but may contain more than the opening).
- **Line Purity**: Aside from optional trailing whitespace, the closing line must contain only the marker run.

A fenced code block consists of all lines between the opening and closing fence. If no closing fence is encountered, the block extends to the end of the document.

The optional info string may follow the opening fence after delimiter whitespace. The first whitespace-delimited token of the info string is extracted as the language token. Only this token is preserved during lowering and is used to generate a language class in the rendered HTML.

Payload lines are normalized relative to the indentation of the opening fence:

- Up to the opening fence indentation column count may be removed from each payload line.
- Only spaces and tabs are consumed during this stripping process.
- Stripping stops when a non-whitespace character is encountered.
- Additional indentation beyond this amount is preserved as literal content.

Line boundaries are preserved exactly and represented by literal newline characters in the payload.

The `FencedCodeBlock` IR node stores:

- The span covering the entire block
- The spans of each payload line
- The indentation column of the opening fence
- The span of the info string

Lowering extracts the language token and normalizes the payload indentation.

#### Rendering Code Blocks

Both indented and fenced code blocks are lowered into a unified `CodeBlock` AST node containing:

- The block span
- The normalized payload content
- An optional language token

The code block payload is rendered as literal text, with line boundaries preserved. Markdown syntax within code blocks is not interpreted as inline elements.

#### HTML Blocks

An HTML block is a raw passthrough block. Its contents are preserved exactly and are not interpreted as markdown.

HTML blocks allow authors to embed raw HTML within Markdown documents when Markdown syntax alone is insufficient. When a supported HTML opener appears at the beginning of a line, the compiler suspends Markdown parsing and treats the block as literal HTML until the appropriate termination condition is met.

HTML blocks may interrupt paragraphs.

A line begins an HTML block if and only if the following conditions are met:

- **Indentation**: The line begins with 0-3 columns of indentation.
- **Position**: The HTML opener begins at the first non-indent byte of the line.
- **Supported Opener**: The line matches one of the supported HTML block opener forms described below.

If none of supported forms match, the line is not treated as an HTML block and normal Markdown parsing continues.

HTML block contents are preserved exactly as written in the source. Within an HTML block, inline parsing does not occur, Markdown constructs are not interpreted, whitespace is not normalized, and HTML is not escaped. During HTML generation, HTML blocks are emitted as raw HTML nodes so that embedded markup is rendered verbatim.

#### Delimiter-Terminated HTML Blocks

Some HTML block forms are terminated by a specific closing delimiter. These blocks continue until a line containing the matching delimiter is encountered.

- **HTML Comment**: Opener `<!--`, Terminator `-->`.
- **Processing Instruction**: Opener `<?`, Terminator `?>`.
- **Declaration**: Opener `<!`, Terminator `>`.
- **CDATA Section**: Opener `<![CDATA[`, Terminator `]]>`.

The closing delimiter may appear on the same line as the opener or any later line. Blank lines are permitted inside the block. If the closing delimiter is nver encountered, the block continues through end of file. The entire line containing the closing delimiter is included in the block.

#### Named Block Tag HTML Blocks

A line begins a named-tag HTML block if it starts with an HTML opening or closing tag whose name belongs to the supported block-level tag set.

The tag name must being with an ASCII letter, contain only ACII letters or digits, and mtch one of the supported block tag names.

Supported tag names:

```html
address article aside blockquote body details dialog div dl fieldset figcaption figure footer form h1 h2 h3 h4 h5 h6 header hr html main menu nav ol p pre section table tbody td tfoot th thead tr ul
```

Inline HTML tags such as `<span>` or `<em>` are not recognized as HTML block openers.

Named-tag HTML blocks continue through subsequent non-blank lines and terminate immediately before the first blank line or at end of file. Matching closing tags do not terminate the block.


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

#### Emphasis (`*`) and Strong Emphasis (`**`)

Emphasis and strong emphasis are inline constructs used to mark text with semantic stress. Emphasis corresponds to `<em>` in HTML and strong emphasis corresponds to `<strong>`.

These constructs are formed using runs of asterisk (`*`) delimiter characters surrounding inline content.

A delimiter run is defined as a maximal sequence of consecutive `*` characters appearing within a paragraph or other inline content.

Delimiter runs participate in emphasis parsing according to the following rules:

- **Delimiter Runs**: A delimiter run consists of one or more consecutive `*` characters. The run length determines how many delimiter characters are available for pairing.
- **Eligibility**: A delimiter run may open emphasis, close emphasis, or both depending on the surrounding characters. Eligibility is determined by inspecting the characters immediately before and after the run.
- **Left-Flanking Runs**: A delimiter run is considered left-flanking (eligible to open emphasis) if the character immediately following the run is not whitespace and either the following character is not punctuation, or the preceding character is whitespace or punctuation.
- **Right-Flanking Runs**: A delimiter run is considered right-flanking (eligible to close emphasis) if the character immediately preceding the run is not whitespace and either the preceding character is not punctuation, or the following character is whitespace or punctuation.
- **Pairing**: When a delimiter run that can close emphasis is encountered, the parser searches backward for the nearest earlier delimiter run that can open emphasis and has remaining delimiter characters available.
- **Consumption**: When a pair is resolved, delimiter characters are consumed from both runs. If both runs contain at least two delimiter characters, two characters are consumed from each run to produce a strong emphasis node. Otherwise, one delimiter character is consumed from each run to produce an emphasis node.
- **Nested Constructs**: Because delimiter runs may contain multiple characters nested emphasis may be produced by resolving multiple pairs from the same runs.

Delimiter characters that cannot participate in a valid pairing are emitted as literial text.

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

Block constructs are parsed according to clear structural rules. Surface syntax is normalized early: distinct syntactic forms that represent the same semantic construct (e.g., ATX headers and Setext headers) are lowered into a single `Header` IR node. Downstream stages operate only on semantic structure, not original delimiter forms.

The system remains mechanically predictable and extensible while preserving precise coordinate semantics throughout.
