package builtin

import (
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"strconv"
)

//Built-in Types

//Base Type

var VariableType = &core.Type{Name: "Variable", Parent: nil, Functions: nil}

//Variable Sub-Types

var StringType = &core.Type{Name: "String", Parent: VariableType, Functions: nil}

//String Sub-Types

var BooleanType = &core.Type{Name: "Boolean", Parent: StringType, Functions: nil}

var FunctionType = &core.Type{Name: "Function", Parent: StringType, Functions: nil}

var IntegerType = &core.Type{Name: "Integer", Parent: StringType, Functions: nil}

type StringInterface interface {
	GetStringValue() string
}

func NewExceptionReturn(exceptionMessage string) *core.Return {
	return &core.Return{
		ReturnType: core.EXCEPTION,
		Pointer:    NewStringPointer(exceptionMessage),
	}
}

func NewStringPointer(value string) *core.Pointer {
	return &core.Pointer{
		Typ:      StringType,
		Variable: &String{Value: value},
	}
}

type String struct {
	Value string
}

func (*String) GetType() *core.Type {
	return StringType
}

func (s *String) GetStringValue() string {
	return s.Value
}

type Integer struct {
	Value int
}

func (Integer) GetType() *core.Type {
	return IntegerType
}

func (i Integer) GetStringValue() string {
	return strconv.Itoa(i.Value)
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
