package builtin

import "github.com/cevatbarisyilmaz/selinus/compiler/core"

var BooleanType = &core.Type{Name: "Boolean", Parent: core.VariableType, Methods: map[string]core.Function{}, Converters: map[*core.Type]core.Function{}, Scope: core.NewScope()}

type Boolean struct {
	Value bool
}

func NewBooleanPointer(value bool) *core.Pointer {
	return &core.Pointer{
		Typ:      BooleanType,
		Variable: core.NewVariable(&Boolean{Value: value}),
	}
}

func (Boolean) GetType() *core.Type {
	return BooleanType
}
