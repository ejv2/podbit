package components

// A Selecter represents any object which has a visibly selected
// element and which is capable of changing said element
type Selecter interface {
	GetSelection() (int, interface{})
	ChangeSelection(index int)
	MoveSelection(offset int)
}

// List represents a generic component which contains elements
// of an arbitrary type. The list can be scrolled and has an
// associated selected element
type List[Item any] struct {
	W, H  int
	Items []Item

	scroll int
	sel    int
}

// GetSelection returns the currently selected menu element, or
// nil if none is selected
func (l *List[Item]) GetSelection() (int, Item) {
	if len(l.Items) < 1 {
		return 0, *new(Item)
	}

	return l.sel, l.Items[l.sel]
}

// ChangeSelection changes the selection to the index specified.
// If index is out of range, no action is taken.
func (l *List[Item]) ChangeSelection(index int) {
	if index >= len(l.Items) || index < 0 {
		return
	}

	l.sel = index

	scrollAt := l.H + l.scroll + 1
	underscrollAt := l.scroll - 1
	if l.sel >= scrollAt {
		l.scroll += (l.sel - scrollAt) + 1
	} else if l.sel <= underscrollAt {
		l.scroll = l.sel
	}
}

// MoveSelection changes the selected item relative to the current
// position. If the new selection would be out of range, no action
// is taken.
func (l *List[Item]) MoveSelection(offset int) {
	off := l.sel + offset
	l.ChangeSelection(off)
}
