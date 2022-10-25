package common

type nodeColor bool

const (
	black nodeColor = false
	red   nodeColor = true
)

type Node[T any] struct {
	color  nodeColor
	value  T
	left   *Node[T]
	right  *Node[T]
	parent *Node[T]
}

func (n *Node[T]) GetValue() T {
	return n.value
}

func (n *Node[T]) Next() *Node[T] {
	switch {
	case n == nil:
		return nil
	case n.getRight() != nil:
		next := n.getRight()
		for left := next.getLeft(); left != nil; left = next.getLeft() {
			next = left
		}
		return next
	case n.getParent() != nil:
		next := n
		for parent := next.getParent(); next == parent.getRight(); parent = next.getParent() {
			next = parent
		}
		return next.getParent()
	default:
		return nil
	}
}

type Tree[T any] struct {
	root *Node[T]
	size int
	less func(*T, *T) bool
}

func NewTree[T any](less func(lhs *T, rhs *T) bool) Tree[T] {
	return Tree[T]{less: less}
}

func (t *Tree[T]) Insert(value T) {
	result := &Node[T]{value: value, parent: t.getInsertionPlace(value)}

	switch {
	case result.parent == nil:
		t.root = result
	case t.less(&result.value, &result.parent.value):
		result.parent.left = result
	default:
		result.parent.right = result
	}

	result.color = red
	t.fixInsert(result)
	t.size++
}

func (t *Tree[T]) Size() int {
	return t.size
}

func (t *Tree[T]) Contains(value T) bool {
	return t.Find(value) != nil
}

func (t *Tree[T]) Find(value T) *Node[T] {
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

func (t *Tree[T]) Begin() *Node[T] {
	begin := t.root
	for left := begin.getLeft(); left != nil; left = begin.getLeft() {
		begin = left
	}

	return begin
}

func (n *Node[T]) getParent() *Node[T] {
	if n == nil {
		return nil
	}
	return n.parent
}

func (n *Node[T]) getLeft() *Node[T] {
	if n == nil {
		return nil
	}
	return n.left
}

func (n *Node[T]) getRight() *Node[T] {
	if n == nil {
		return nil
	}
	return n.right
}

func (n *Node[T]) getColor() nodeColor {
	if n == nil {
		return black
	}
	return n.color
}

func (t *Tree[T]) swapParents(x, y *Node[T]) {
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

func (t *Tree[T]) rotateLeft(x *Node[T]) {
	y := x.right
	x.right, y.left = y.left, x
	if x.right != nil {
		x.right.parent = x
	}
	t.swapParents(x, y)
}

func (t *Tree[T]) rotateRight(x *Node[T]) {
	y := x.left
	x.left, y.right = y.right, x
	if x.left != nil {
		x.left.parent = x
	}
	t.swapParents(x, y)
}

func (t *Tree[T]) fixInsert(node *Node[T]) {
	for node.getParent().getColor() == red {
		//node is not nil and has a parent and grandparent
		if node.parent == node.parent.parent.left {
			y := node.getParent().getParent().getRight() //y migh be nil
			switch {
			case y.getColor() == red:
				//y is not nil
				node.parent.color = black
				y.color = black
				node.parent.parent.color = red
				node = node.getParent().getParent() //node might be nil
			case node == node.parent.right:
				node = node.parent
				t.rotateLeft(node)
				fallthrough
			case node == node.parent.left:
				node.parent.color = black
				node.parent.parent.color = red
				t.rotateRight(node.parent.parent)
			}
		} else {
			y := node.getParent().getParent().getLeft() //y might be nil
			switch {
			case y.getColor() == red:
				//y is not nil
				node.parent.color = black
				y.color = black
				node.parent.parent.color = red
				node = node.getParent().getParent() //node might be nil
			case node == node.parent.left:
				node = node.parent
				t.rotateRight(node)
				fallthrough
			case node == node.parent.right:
				node.parent.color = black
				node.parent.parent.color = red
				t.rotateLeft(node.parent.parent)
			}
		}
	}
	//fixInsert is called only on insertion => tree has at least one node => root is not nil
	t.root.color = black
}

func (t *Tree[T]) getInsertionPlace(value T) *Node[T] {

	var y *Node[T]
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
