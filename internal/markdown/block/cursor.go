package block

type Cursor struct {
	Lines []Line
	Index int
}

func NewCursor(lines []Line) *Cursor {
	return &Cursor{
		Lines: lines,
		Index: 0,
	}
}

func (c *Cursor) Peek() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	return c.Lines[c.Index], true
}

func (c *Cursor) Next() (Line, bool) {
	if c.EOF() {
		return Line{}, false
	}

	out := c.Lines[c.Index]
	c.Index++
	return out, true
}

func (c *Cursor) EOF() bool {
	return c.Index >= len(c.Lines)
}

func (c *Cursor) SkipBlankLines() {
	for {
		line, ok := c.Peek()
		if !ok {
			return
		}
		if !line.IsBlankLine() {
			return
		}

		c.Next()
	}
}
