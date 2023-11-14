package lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length    int
	firstItem *ListItem
	lastItem  *ListItem
}

func pushedToEmptyList(l *list, i *ListItem) bool {
	if l.length == 0 {
		i.Prev, i.Next = nil, nil
		l.firstItem, l.lastItem = i, i
		l.length++
		return true
	}
	return false
}

func pushFront(l *list, i *ListItem) {
	if pushedToEmptyList(l, i) {
		return
	}

	l.firstItem.Prev = i
	i.Prev, i.Next = nil, l.firstItem

	l.firstItem = i
	l.length++
}

func pushBack(l *list, i *ListItem) {
	if pushedToEmptyList(l, i) {
		return
	}

	l.lastItem.Next = i
	i.Prev, i.Next = l.lastItem, nil

	l.lastItem = i
	l.length++
}

func (l list) Len() int {
	return l.length
}

func (l list) Front() *ListItem {
	return l.firstItem
}

func (l list) Back() *ListItem {
	return l.lastItem
}

func (l *list) PushFront(v interface{}) *ListItem {
	pushFront(l, &ListItem{Value: v})
	return l.firstItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	pushBack(l, &ListItem{Value: v})
	return l.lastItem
}

func (l *list) Remove(i *ListItem) {
	prevItem := i.Prev
	nextItem := i.Next

	if prevItem != nil {
		prevItem.Next = i.Next
	} else {
		l.firstItem = nextItem
	}

	if nextItem != nil {
		nextItem.Prev = i.Prev
	} else {
		l.lastItem = prevItem
	}

	i.Prev, i.Next = nil, nil
	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if l.firstItem == i {
		return
	}

	l.Remove(i)
	pushFront(l, i)
}

func NewList() List {
	var l List = &list{
		length:    0,
		firstItem: nil,
		lastItem:  nil,
	}
	return l
}
