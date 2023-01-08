package core

type NodeRoot interface {
	Execute(*Scope) *Return
}

type Node interface {
	Execute(*Scope) *Return
	SetNext(Node)
	Next() Node
	Root() NodeRoot
	Position() string
}

func NewNode(nodeRoot NodeRoot, position string) Node {
	return &node{
		root:     nodeRoot,
		next:     nil,
		position: position,
	}
}

type node struct {
	root     NodeRoot
	next     Node
	position string
}

func (node *node) Execute(scope *Scope) *Return {
	res := node.root.Execute(scope)
	AddPositionToStackTrace(res, node.Position())
	return res
}

func (node *node) SetNext(next Node) {
	node.next = next
}

func (node *node) Next() Node {
	return node.next
}

func (node *node) Root() NodeRoot {
	return node.root
}

func (node *node) Position() string {
	return node.position
}
