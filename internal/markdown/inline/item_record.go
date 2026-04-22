package inline

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

// ItemKind identifies the kind of inline item tracked during inline parsing.
type ItemKind int

const (
	_ ItemKind = iota
	ItemText
	ItemCodeSpan
	ItemLink
	ItemImage
	ItemAutolinkURI
	ItemAutolinkEmail
	ItemHTML
	ItemEmphasis
	ItemStrong
)

// ItemRecord represents a provisional or resolved inline item in the
// mutable item list used during inline parsing.
type ItemRecord struct {
	next, prev *ItemRecord
	list       *ItemList

	OriginalSpan source.ByteSpan
	LiveSpan     source.ByteSpan
	Kind         ItemKind

	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool

	Children *ItemList
}

// Next returns the next list ItemRecord or nil.
func (r *ItemRecord) Next() *ItemRecord {
	if p := r.next; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list ItemRecord or nil.
func (r *ItemRecord) Prev() *ItemRecord {
	if p := r.prev; r.list != nil && p != &r.list.root {
		return p
	}
	return nil
}

// ItemList is a doubly-linked list of item records used during inline parsing.
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

// NewItemList returns an initialized list.
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

// InsertAfter inserts a new ItemRecord item immediately after mark.
func (l *ItemList) InsertAfter(item, mark *ItemRecord) {
	if mark == nil || mark.list != l {
		return
	}

	item.prev = mark
	item.next = mark.next
	item.prev.next = item
	item.next.prev = item
	item.list = l
	l.len++
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

// Remove removes item from l if item is an element of list l.
func (l *ItemList) Remove(item *ItemRecord) {
	if item.list == l {
		item.prev.next = item.next
		item.next.prev = item.prev
		item.next = nil
		item.prev = nil
		item.list = nil
		l.len--
	}
}

// DetachRange removes the contiguous range [first, last] from list l
// and returns a new ItemList containing that range.
func (l *ItemList) DetachRange(first, last *ItemRecord) *ItemList {
	if first == nil || last == nil {
		panic("DetachRange: first and last must be non-nil")
	}

	if first.list != l {
		panic("DetachRange: first does not belong to receiver list")
	}
	if last.list != l {
		panic("DetachRange: last does not belong to receiver list")
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
		panic("DetachRange: last is not reachable from first in receiver list")
	}

	newList := NewItemList()

	before := first.prev
	after := last.next

	before.next = after
	after.prev = before

	newList.root.next = first
	newList.root.prev = last
	first.prev = &newList.root
	last.next = &newList.root

	for item := first; item != &newList.root; item = item.next {
		item.list = newList
	}

	newList.len = count
	l.len -= count

	return newList
}
