package common

type nodeColor bool

const (
	black nodeColor = false
	red   nodeColor = true
)

type node[T any] struct {
	color  nodeColor
	value  T
	left   *node[T]
	right  *node[T]
	parent *node[T]
}

type Node[T any] interface {
	GetValue() T
}

func (n *node[T]) GetValue() T {
	return n.value
}

func (n *node[T]) getParent() *node[T] {
	if n == nil {
		return nil
	}
	return n.parent
}

func (n *node[T]) getLeft() *node[T] {
	if n == nil {
		return nil
	}
	return n.left
}

func (n *node[T]) getRight() *node[T] {
	if n == nil {
		return nil
	}
	return n.right
}

func (n *node[T]) getColor() nodeColor {
	if n == nil {
		return black
	}
	return n.color
}

type tree[T any] struct {
	root *node[T]
	size int
	less func(*T, *T) bool
}

func (t *tree[T]) swapParents(x, y *node[T]) {
	y.parent = x.parent
	switch {
	case y.parent == nil:
		t.root = y
	case x == x.parent.left:
		x.parent.left = y
	default:
		x.parent.right = y
	}
	x.parent = y
}

func (t *tree[T]) rotateLeft(x *node[T]) {
	y := x.right
	x.right, y.left = y.left, x
	if x.right != nil {
		x.right.parent = x
	}
	t.swapParents(x, y)
}

func (t *tree[T]) rotateRight(x *node[T]) {
	y := x.left
	x.left, y.right = y.right, x
	if x.left != nil {
		x.left.parent = x
	}
	t.swapParents(x, y)
}

func (t *tree[T]) fixInsert(node *node[T]) {
	for node.getParent().getColor() == red {
		if node.getParent() == node.getParent().getParent().getLeft() {
			y := node.getParent().getParent().getRight()
			switch {
			case y.getColor() == red:
				node.getParent().color = black
				y.color = black
				node.getParent().getParent().color = red
				node = node.getParent().getParent()
			case node == node.getParent().getRight():
				node = node.getParent()
				t.rotateLeft(node)
				fallthrough
			case node == node.getParent().getLeft():
				node.getParent().color = black
				node.getParent().getParent().color = red
				t.rotateRight(node.getParent().getParent())
			}
		} else {
			y := node.getParent().getParent().getLeft()
			switch {
			case y.getColor() == red:
				node.getParent().color = black
				y.color = black
				node.getParent().getParent().color = red
				node = node.getParent().getParent()
			case node == node.getParent().getLeft():
				node = node.getParent()
				t.rotateRight(node)
				fallthrough
			case node == node.getParent().getRight():
				node.getParent().color = black
				node.getParent().getParent().color = red
				t.rotateLeft(node.getParent().getParent())
			}
		}
	}
	t.root.color = black
}

func (t *tree[T]) getInsertionPlace(value T) *node[T] {

	var y *node[T]
	for x := t.root; x != nil; {
		y = x
		if t.less(&value, &x.value) {
			x = x.left
		} else {
			x = x.right
		}
	}

	return y
}

func NewTree[T any](less func(lhs *T, rhs *T) bool) tree[T] {
	return tree[T]{less: less}
}

func (t *tree[T]) Insert(value T) {
	result := &node[T]{value: value, parent: t.getInsertionPlace(value)}

	if result.parent == nil {
		t.root = result
		//return
	} else if t.less(&result.value, &result.parent.value) {
		result.parent.left = result
	} else {
		result.parent.right = result
	}
	result.color = red
	t.fixInsert(result)
	t.size++
}

func (t *tree[T]) Size() int {
	return t.size
}

func (t *tree[T]) Contains(value T) bool {
	return t.Find(value) != nil
}

func (t *tree[T]) Find(value T) Node[T] {
	for x := t.root; x != nil; {
		switch {
		case t.less(&value, &x.value):
			x = x.left
		case t.less(&x.value, &value):
			x = x.right
		default:
			return x
		}
	}
	return nil
}
