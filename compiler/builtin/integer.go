package builtin

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
)

var IntegerType = &core.Type{Name: "Integer", Parent: core.VariableType, Methods: map[string]core.Function{}, Converters: map[*core.Type]core.Function{
	StringType: &IntegerToStringConverterFunction{},
}, Scope: core.NewScope()}

type Integer struct {
	Value int64
}

func (*Integer) GetType() *core.Type {
	return IntegerType
}

func NewIntegerPointer(value int64) *core.Pointer {
	return &core.Pointer{
		Typ:      IntegerType,
		Variable: core.NewVariable(&Integer{Value: value}),
	}
}

var IntegerToStringConverterFunctionType = &core.Type{Parent: FunctionType, Name: "convertIntegerToString", Generic: true, Generics: []*core.Type{StringType}}

type IntegerToStringConverterFunction struct{}

func (integerToStringConverterFunction *IntegerToStringConverterFunction) Execute(scope *core.Scope) *core.Return {
	getResult := scope.Get(core.Self)
	if getResult.ReturnType != core.NOTHING {
		return getResult
	}
	result := fmt.Sprint(getResult.Pointer.Variable.VariableInterface.(*Integer).Value)
	return &core.Return{
		ReturnType: core.NOTHING,
		Pointer:    NewStringPointer(result),
	}
}

func (integerToStringConverterFunction *IntegerToStringConverterFunction) GetType() *core.Type {
	return IntegerToStringConverterFunctionType
}

func (integerToStringConverterFunction *IntegerToStringConverterFunction) GetParameters() []*core.Parameter {
	return nil
}

func (integerToStringConverterFunction *IntegerToStringConverterFunction) GetReturnType() *core.Type {
	return StringType
}

func (integerToStringConverterFunction *IntegerToStringConverterFunction) GetScope() *core.Scope {
	return scope
}
