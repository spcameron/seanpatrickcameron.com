package inline

// DelimiterKind identifies the kind of delimiter tracked during inline parsing.
type DelimiterKind int

const (
	_ DelimiterKind = iota
	DelimOpenBracket
	DelimImageOpenBracket
	DelimAsterisk
	DelimUnderscore
)

// DelimiterRecord represents a delimiter in the inline parse, participating
// in a doubly-linked delimiter stack used for emphasis and link resolution.
type DelimiterRecord struct {
	next, prev *DelimiterRecord
	list       *DelimiterList

	Item     *ItemRecord
	Kind     DelimiterKind
	Count    int
	Active   bool
	CanOpen  bool
	CanClose bool
}

// Next returns the next list DelimiterRecord or nil.
func (r *DelimiterRecord) Next() *DelimiterRecord {
	if p := r.next; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

// Prev returns the prev list DelimiterRecord or nil.
func (r *DelimiterRecord) Prev() *DelimiterRecord {
	if p := r.prev; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

// DelimiterList is a doubly-linked list of delimiter records used during
// inline parsing.
type DelimiterList struct {
	root DelimiterRecord
	len  int
}

// Init initializes or clears list l.
func (l *DelimiterList) Init() *DelimiterList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// NewDelimiterList return an initialized list.
func NewDelimiterList() *DelimiterList {
	return new(DelimiterList).Init()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *DelimiterList) Front() *DelimiterRecord {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *DelimiterList) Back() *DelimiterRecord {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// PushBack inserts a DelimiterRecord at the back of list l and returns
// the DelimiterRecord.
func (l *DelimiterList) PushBack(delim *DelimiterRecord) *DelimiterRecord {
	last := l.root.prev
	delim.prev = last
	delim.next = last.next
	delim.prev.next = delim
	delim.next.prev = delim
	delim.list = l
	l.len++
	return delim
}

// Remove removes delim from l if delim is an element of list l.
func (l *DelimiterList) Remove(delim *DelimiterRecord) {
	if delim.list == l {
		delim.prev.next = delim.next
		delim.next.prev = delim.prev
		delim.next = nil
		delim.prev = nil
		delim.list = nil
		l.len--
	}
}

// RemoveRange removes the contiguous range [first, last] from list l.
func (l *DelimiterList) RemoveRange(first, last *DelimiterRecord) {
	if first == nil || last == nil {
		panic("RemoveRange: first and last must be non-nil")
	}

	if first.list != l {
		panic("RemoveRange: first does not belong to receiver list")
	}
	if last.list != l {
		panic("RemoveRange: last does not belong to receiver list")
	}

	found := false
	count := 0
	for item := first; item != nil; item = item.Next() {
		count++
		if item == last {
			found = true
			break
		}
	}
	if !found {
		panic("RemoveRange: last is not reachable from first in receiver list")
	}

	before := first.prev
	after := last.next

	before.next = after
	after.prev = before

	item := first
	for item != after {
		next := item.next
		item.prev = nil
		item.next = nil
		item.list = nil
		item = next
	}

	l.len -= count
}
