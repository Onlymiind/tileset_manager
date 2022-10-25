package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {
	tree := NewTree(func(lhs, rhs *TileRef) bool { return lhs.Less(rhs) })
	values := []TileRef{
		{Range: IndexRange{1, 3}},
		{Range: IndexRange{4, 6}},
		{Range: IndexRange{7, 10}},
		{Range: IndexRange{23, 40}},
		{Range: IndexRange{44, 64}},
		{Range: IndexRange{70, 90}},
		{Range: IndexRange{91, 91}},
		{Range: IndexRange{98, 100}},
		{Range: IndexRange{128, 129}},
		{Range: IndexRange{200, 255}},
	}

	t.Run("insertion", func(t *testing.T) {
		for _, val := range values {
			tree.Insert(val)
		}
		for _, val := range values {
			assert.True(t, tree.Contains(val), "inserted value is not present")
		}

		assert.Equal(t, len(values), tree.Size(), "wrong size")
	})

	t.Run("find", func(t *testing.T) {
		n := tree.Find(TileRef{Range: IndexRange{2, 2}})
		n2 := tree.Find(TileRef{Range: IndexRange{8, 8}})
		n3 := tree.Find(TileRef{Range: IndexRange{199, 199}})
		n4 := tree.Find(TileRef{Range: IndexRange{91, 91}})
		n5 := tree.Find(TileRef{Range: IndexRange{200, 220}})
		var nilNode *Node[TileRef]

		assert.Equal(t, IndexRange{1, 3}, n.GetValue().Range)
		assert.Equal(t, IndexRange{7, 10}, n2.GetValue().Range)
		assert.Equal(t, nilNode, n3)
		assert.Equal(t, IndexRange{91, 91}, n4.GetValue().Range)
		assert.Equal(t, IndexRange{200, 255}, n5.GetValue().Range)
	})
	t.Run("iteration", func(t *testing.T) {
		intTree := NewTree(func(lhs, rhs *int) bool { return *lhs < *rhs })
		data := map[int]struct{}{
			1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {},
			7: {}, 8: {}, 9: {}, 10: {}, 11: {}, 12: {},
			13: {}, 14: {}, 15: {}, 16: {},
		}
		for val := range data {
			intTree.Insert(val)
		}

		found := map[int]struct{}{}

		for it := intTree.Begin(); it != nil; it = it.Next() {
			found[it.GetValue()] = struct{}{}
		}

		for val := range found {
			_, ok := data[val]
			assert.True(t, ok, "excessive data")
		}

		for val := range data {
			_, ok := found[val]
			assert.True(t, ok, "value not present")
		}
	})
}

func makeValues[T any](b *testing.B, getter func(int) T) []T {
	b.StopTimer()
	b.ResetTimer()
	s := make([]T, 0, b.N)
	for cnt := 0; cnt < b.N; cnt++ {
		s = append(s, getter(cnt))
	}
	return s
}

func benchmarkTree[T any](b *testing.B, less func(*T, *T) bool, getter func(int) T) {
	tree := NewTree(less)
	b.Run("insetrion", func(b *testing.B) {
		s := makeValues(b, getter)
		b.StartTimer()
		for cnt := 0; cnt < b.N; cnt++ {
			tree.Insert(s[cnt])
		}
	})
	b.Run("search", func(b *testing.B) {
		s := makeValues(b, getter)
		b.StartTimer()
		for cnt := 0; cnt < b.N; cnt++ {
			_ = tree.Find(s[cnt])
		}
	})
}

func BenchmarkInts(b *testing.B) {
	benchmarkTree(b, func(lhs *int, rhs *int) bool { return *lhs < *rhs }, func(i int) int { return i })
}
