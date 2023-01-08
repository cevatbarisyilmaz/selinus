package core

import (
	"fmt"
)

type ScopeBlock struct {
	References map[string]*Pointer
}

func (block *ScopeBlock) Get(name string) *Pointer {
	return block.References[name]
}

func NewEmptyScopeBlock() *ScopeBlock {
	return &ScopeBlock{References: make(map[string]*Pointer)}
}

func NewScopeBlock(references map[string]*Pointer) *ScopeBlock {
	return &ScopeBlock{References: references}
}

type Scope struct {
	Blocks []*ScopeBlock
	Name   string
}

func (scope *Scope) Clone() *Scope {
	return &Scope{Blocks: append([]*ScopeBlock(nil), scope.Blocks...), Name: scope.Name + "-Copy"}
}

func (scope *Scope) CloneWithName(name string) *Scope {
	return &Scope{Blocks: append([]*ScopeBlock(nil), scope.Blocks...), Name: scope.Name + "-" + name + "Copy"}
}

func NewScope() *Scope {
	return &Scope{}
}

func NewScopeWithName(name string) *Scope {
	return &Scope{Name: name}
}

func (scope *Scope) getCurrentBlock() *ScopeBlock {
	return scope.Blocks[len(scope.Blocks)-1]
}

func (scope *Scope) AddBlock(block *ScopeBlock) {
	scope.Blocks = append(scope.Blocks, block)
}

func (scope *Scope) CreateBlock() {
	scope.Blocks = append(scope.Blocks, NewEmptyScopeBlock())
}

func (scope *Scope) ReleaseBlock() {
	scope.Blocks = scope.Blocks[:len(scope.Blocks)-1]
}

func (scope *Scope) Print() {
	fmt.Println("Scope: ", scope.GetName())
	for i, block := range scope.Blocks {
		fmt.Println("Block ", i)
		for key, value := range block.References {
			fmt.Println(key, ":", value.Typ.Name)
		}
	}
}

func (scope *Scope) Get(name string) *Return {
	for i := len(scope.Blocks) - 1; i >= 0; i-- {
		t := scope.Blocks[i].Get(name)
		if t != nil {
			return &Return{
				ReturnType: NOTHING,
				Pointer:    t,
			}
		}
	}
	return NewExceptionReturn(name + " is not declared")
}

func (scope *Scope) MustGet(name string) *Pointer {
	for i := len(scope.Blocks) - 1; i >= 0; i-- {
		t := scope.Blocks[i].Get(name)
		if t != nil {
			return t
		}
	}
	return nil
}

func (scope *Scope) DeclareAndSet(name string, value *Pointer) {
	scope.getCurrentBlock().References[name] = value
}

func (scope *Scope) Set(name string, value *Pointer) {
	scope.getCurrentBlock().References[name].Variable = value.Variable
}

func (scope *Scope) Declare(name string, typ *Type) {
	scope.getCurrentBlock().References[name] = &Pointer{Typ: typ, Variable: nil}
}

func (scope *Scope) SetName(name string) {
	scope.Name = name
}

func (scope *Scope) GetName() string {
	return scope.Name
}
