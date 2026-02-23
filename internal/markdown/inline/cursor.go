package inline

type Cursor struct {
	Events []Event
	Index  int
}

func NewCursor(events []Event) *Cursor {
	return &Cursor{
		Events: events,
		Index:  0,
	}
}

func (c *Cursor) Peek() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	return c.Events[c.Index], true
}

func (c *Cursor) Next() (Event, bool) {
	if c.EOF() {
		return Event{}, false
	}

	out := c.Events[c.Index]
	c.Index++
	return out, true
}

func (c *Cursor) EOF() bool {
	return c.Index >= len(c.Events)
}
