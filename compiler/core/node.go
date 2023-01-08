package core

type Node interface {
	Execute(*Scope) *Return
	Next() Node
	SetNext(Node)
}
