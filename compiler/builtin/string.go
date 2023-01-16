package builtin

import "github.com/cevatbarisyilmaz/selinus/compiler/core"

var StringType = &core.Type{Name: "String", Parent: core.VariableType, Methods: map[string]core.Function{}, Converters: map[*core.Type]core.Function{}, Scope: core.NewScope()}

type String struct {
	Value string
}

func (*String) GetType() *core.Type {
	return StringType
}

func (s *String) GetStringValue() string {
	return s.Value
}

func NewStringPointer(value string) *core.Pointer {
	return &core.Pointer{
		Typ:      StringType,
		Variable: core.NewVariable(&String{Value: value}),
	}
}
