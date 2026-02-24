package block

import (
	"strings"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ir"
)

type BuildRule interface {
	Apply(c *Cursor) (ir.Block, bool, error)
}

type HeaderRule struct{}

func (r HeaderRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine() {
		return nil, false, nil
	}

	// count the leading spaces, reject if more than 3
	offset, ok := line.BlockIndent()
	if !ok {
		return nil, false, nil
	}

	s := line.Text
	pos := offset
	level := 0

	// validate the marker
	if pos >= len(s) || s[pos] != '#' {
		return nil, false, nil
	}

	// count the marker run, reject if more than 6
	for pos < len(s) && s[pos] == '#' {
		pos++
		level++

		if level == 7 {
			return nil, false, nil
		}
	}

	// validate the delimiter
	if pos >= len(s) || (s[pos] != ' ' && s[pos] != '\t') {
		return nil, false, nil
	}

	// consume the delimiter
	for pos < len(s) && (s[pos] == ' ' || s[pos] == '\t') {
		pos++
	}

	// trim trailing spaces and tabs
	content := strings.TrimRight(s[pos:], " \t")

	start := c.Index
	line, _ = c.Next()
	end := c.Index

	span := ir.LineSpan{
		Start: start,
		End:   end,
	}

	applied := ir.Header{
		Level: level,
		Text:  content,
		Span:  span,
	}

	return applied, true, nil
}

type ParagraphRuleMarker interface {
	isParagraphRule()
}

type ParagraphRule struct{}

func (ParagraphRule) isParagraphRule() {}

func (r ParagraphRule) Apply(c *Cursor) (ir.Block, bool, error) {
	line, ok := c.Peek()
	if !ok {
		return nil, false, nil
	}
	if line.IsBlankLine() {
		return nil, false, nil
	}

	var s []string
	start := c.Index

	line, _ = c.Next()
	s = append(s, line.Text)

	for {
		line, ok := c.Peek()
		if !ok {
			break
		}
		if line.IsBlankLine() {
			break
		}

		startsBlock, err := c.StartsNonParagraphBlock()
		if err != nil {
			return nil, false, err
		}
		if startsBlock {
			break
		}

		line, _ = c.Next()
		s = append(s, line.Text)
	}

	end := c.Index

	span := ir.LineSpan{
		Start: start,
		End:   end,
	}

	content := strings.Join(s, "\n")

	applied := ir.Paragraph{
		Text: content,
		Span: span,
	}

	return applied, true, nil
}
