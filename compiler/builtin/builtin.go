package builtin

import (
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"strconv"
)

//String Sub-Types

var BooleanType = &core.Type{Name: "Boolean", Parent: core.StringType, Functions: nil}

var FunctionType = &core.Type{Name: "Function", Parent: core.StringType, Functions: nil}

var IntegerType = &core.Type{Name: "Integer", Parent: core.StringType, Functions: nil}

type StringInterface interface {
	GetStringValue() string
}

type Integer struct {
	Value int
}

func (*Integer) GetType() *core.Type {
	return IntegerType
}

func (i *Integer) GetStringValue() string {
	return strconv.Itoa(i.Value)
}

func NewIntegerPointer(value int) *core.Pointer {
	return &core.Pointer{
		Typ:      IntegerType,
		Variable: &Integer{Value: value},
	}
}

type Boolean struct {
	Value bool
}

func (Boolean) GetType() *core.Type {
	return BooleanType
}

func (b Boolean) GetStringValue() string {
	if b.Value {
		return "true"
	}
	return "false"
}

//Built-in Functions

var Block = core.NewScopeBlock(map[string]*core.Pointer{
	"int": {Typ: core.TypeType, Variable: core.TypeToVariable(IntegerType)},
})
