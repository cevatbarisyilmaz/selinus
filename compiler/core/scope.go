package core

type ScopeBlock struct {
	references map[string]*Pointer
}

func (block *ScopeBlock) Get(name string) *Pointer {
	return block.references[name]
}

func NewEmptyScopeBlock() *ScopeBlock {
	return &ScopeBlock{references: make(map[string]*Pointer)}
}

func NewScopeBlock(references map[string]*Pointer) *ScopeBlock {
	return &ScopeBlock{references: references}
}

type Scope struct {
	blocks []*ScopeBlock
}

func NewScope() *Scope {
	return &Scope{}
}

func (scope *Scope) getCurrentBlock() *ScopeBlock {
	return scope.blocks[len(scope.blocks)-1]
}

func (scope *Scope) AddBlock(block *ScopeBlock) {
	scope.blocks = append(scope.blocks, block)
}

func (scope *Scope) CreateBlock() {
	scope.blocks = append(scope.blocks, NewEmptyScopeBlock())
}

func (scope *Scope) ReleaseBlock() {
	scope.blocks = scope.blocks[:len(scope.blocks)-1]
}

func (scope *Scope) Get(name string) *Pointer {
	for i := len(scope.blocks) - 1; i >= 0; i-- {
		t := scope.blocks[i].Get(name)
		if t != nil {
			return t
		}
	}
	return nil
}

func (scope *Scope) DeclareAndSet(name string, value *Pointer) {
	scope.getCurrentBlock().references[name] = value
}

func (scope *Scope) Set(name string, value *Pointer) {
	scope.getCurrentBlock().references[name].Variable = value.Variable
}

func (scope *Scope) Declare(name string, typ *Type) {
	scope.getCurrentBlock().references[name] = &Pointer{Typ: typ, Variable: nil}
}
