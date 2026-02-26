package diagnostic

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"
)

type Severity int

const (
	_ Severity = iota
	SeverityError
	SeverityWarning
)

type Diagnostic struct {
	Message  string
	Span     source.ByteSpan
	Severity Severity
}

func (d Diagnostic) Format(src *source.Source) string {
	line, col := src.LineColumn(d.Span.Start)
	lineSpan := src.LineSpan(line)
	lineText := src.Slice(lineSpan)

	dl, dc := line+1, col+1

	ln := strconv.Itoa(dl)
	w := len(ln)

	blankPad := strings.Repeat(" ", w)
	caretPad := strings.Repeat(" ", col)

	headerLine := fmt.Sprintf("%s at %d:%d\n", d.Message, dl, dc)
	gutterLine := fmt.Sprintf("%s |\n", blankPad)
	sourceLine := fmt.Sprintf("%s | %s\n", ln, lineText)
	caretLine := fmt.Sprintf("%s | %s^\n", blankPad, caretPad)

	return headerLine + gutterLine + sourceLine + caretLine
}

type DiagnosticError struct {
	Diagnostic Diagnostic
}

func (e DiagnosticError) Error() string {
	return e.Diagnostic.Message
}
