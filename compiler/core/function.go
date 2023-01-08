package core

type Parameter struct {
	Name         string
	Typ          *Type
	DefaultValue *Pointer
}

type Function interface {
	Execute(*Scope) *Return
	GetType() *Type
	GetParameters() []*Parameter
	GetReturnType() *Type
	GetScope() *Scope
}

type CustomFunction struct {
	EntryNode  Node
	Parameters []*Parameter
	Typ        *Type
	ReturnType *Type
	Scope      *Scope
}

func (c *CustomFunction) GetScope() *Scope {
	return c.Scope
}

func (c *CustomFunction) Execute(scope *Scope) *Return {
	scope.CreateBlock()
	defer scope.ReleaseBlock()
	node := c.EntryNode
	for node != nil {
		internalReturn := node.Execute(scope)
		if internalReturn.ReturnType != NOTHING {
			if internalReturn.ReturnType == RETURN {
				return &Return{ReturnType: NOTHING, Pointer: internalReturn.Pointer}
			}
			return internalReturn
		}
		node = node.Next()
	}
	return &Return{ReturnType: NOTHING, Pointer: nil}
}

func (c *CustomFunction) GetType() *Type {
	return c.Typ
}

func (c *CustomFunction) GetParameters() []*Parameter {
	return c.Parameters
}

func (c *CustomFunction) GetReturnType() *Type {
	return c.ReturnType
}
