package inline

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type ItemKind int

const (
	_ ItemKind = iota
	ItemText
)

type ItemRecord struct {
	next, prev *ItemRecord
	list       *ItemList

	OriginalSpan source.ByteSpan
	LiveSpan     source.ByteSpan
	Kind         ItemKind
}

// Next returns the next list ItemRecord or nil.
func (r *ItemRecord) Next() *ItemRecord {
	if p := r.next; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

// PRev returns the previous list ItemRecord or nil.
func (r *ItemRecord) Prev() *ItemRecord {
	if p := r.prev; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

type ItemList struct {
	root ItemRecord
	len  int
}

// Init initializes or clears list l.
func (l *ItemList) Init() *ItemList {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

// New returns an initialized list.
func NewItemList() *ItemList {
	return new(ItemList).Init()
}

// Front returns the first element of list l or nil if the list is empty.
func (l *ItemList) Front() *ItemRecord {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *ItemList) Back() *ItemRecord {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// PushBack inserts an ItemRecord at the back of list l and returns the ItemRecord.
func (l *ItemList) PushBack(item *ItemRecord) *ItemRecord {
	last := l.root.prev
	item.prev = last
	item.next = last.next
	item.prev.next = item
	item.next.prev = item
	item.list = l
	l.len++
	return item
}
