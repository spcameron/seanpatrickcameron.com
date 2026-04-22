// Package source defines the canonical coordinate system for Markdown parsing.
//
// Input is normalized to a single newline representation, and all byte
// positions and spans refer to this normalized buffer.
//
// Conventions:
//
//   - ByteSpan is half-open: [Start, End)
//   - Start and End are byte offsets into Source.Raw
//   - Source.LineStarts records the byte offset of each line start
//   - If the input ends with '\n', a final empty line start is included
//
// IR and AST nodes carry ByteSpan values. Line and column information is
// derived on demand from the Source.
package source
