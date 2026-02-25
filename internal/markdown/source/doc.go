// Package source defines the canonical coordinate system for Markdown parsing.
//
// Source.Raw is normalized at ingest:
//   - '/r/n' becomes '/n'
//   - remaining 'r' becomes '/n'
//
// All byte positions and spans refer to this normalized buffer.
//
// Conventions:
//   - ByteSpan is half-open: [Start, End)
//   - Start and End are byte offsets into Source.Raw
//   - Source.LineStarts contains byte offsets of each line start
//   - If the input ends with '\n', the final empty line start (len(Source.Raw)) is included in LineStarts
//
// IR and AST nodes should store ByteSpan. Line/Column is derived on demand using Source.LineColumn.
package source
