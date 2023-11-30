package lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	checkEmpty := func(l List) {
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	}

	checkListProps := func(l List, length int, firstVal interface{}, lastVal interface{}) {
		require.Equal(t, length, l.Len())

		require.Equal(t, firstVal, l.Front().Value)
		require.Nil(t, l.Front().Prev)
		if l.Len() > 1 {
			require.NotNil(t, l.Front().Next)
		} else {
			require.Nil(t, l.Front().Next)
		}

		require.Equal(t, lastVal, l.Back().Value)
		require.Nil(t, l.Back().Next)
		if l.Len() > 1 {
			require.NotNil(t, l.Back().Prev)
		} else {
			require.Nil(t, l.Back().Prev)
		}
	}

	t.Run(
		"empty list", func(t *testing.T) {
			l := NewList()
			checkEmpty(l)
		},
	)

	t.Run(
		"only one item in list", func(t *testing.T) {
			l := NewList()

			l.PushFront(10) // [10]
			checkListProps(l, 1, 10, 10)
			l.Remove(l.Back()) // []
			checkEmpty(l)

			l.PushBack(20) // [20]
			checkListProps(l, 1, 20, 20)
			l.MoveToFront(l.Back()) // [20]
			checkListProps(l, 1, 20, 20)
			l.Remove(l.Front()) // []
			checkEmpty(l)
		},
	)

	t.Run(
		"complex", func(t *testing.T) {
			l := NewList()

			l.PushFront(10) // [10]
			l.PushBack(20)  // [10, 20]
			l.PushBack(30)  // [10, 20, 30]
			require.Equal(t, 3, l.Len())

			middle := l.Front().Next // 20
			l.Remove(middle)         // [10, 30]
			require.Equal(t, 2, l.Len())

			for i, v := range [...]int{40, 50, 60, 70, 80} {
				if i%2 == 0 {
					l.PushFront(v)
				} else {
					l.PushBack(v)
				}
			} // [80, 60, 40, 10, 30, 50, 70]
			checkListProps(l, 7, 80, 70)

			l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
			checkListProps(l, 7, 80, 70)

			l.MoveToFront(l.Back()) // [70, 80, 60, 40, 10, 30, 50]
			elems := make([]int, 0, l.Len())
			for i := l.Front(); i != nil; i = i.Next {
				elems = append(elems, i.Value.(int))
			}
			require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
		},
	)

	t.Run(
		"remove first/last item", func(t *testing.T) {
			l := NewList()

			l.PushFront(10) // [10]
			l.PushFront(20) // [20, 10]
			l.PushFront(30) // [30, 20, 10]
			l.PushFront(40) // [40, 30, 20, 10]
			checkListProps(l, 4, 40, 10)

			l.Remove(l.Front()) // [30, 20, 10]
			checkListProps(l, 3, 30, 10)

			l.Remove(l.Back()) // [30, 20]
			checkListProps(l, 2, 30, 20)
		},
	)
}
