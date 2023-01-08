package native

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler/builtin"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
)

var PrintFunctionType = &core.Type{Parent: builtin.FunctionType, Name: "PrintFunction", Functions: nil, Generic: true, Generics: []*core.Type{nil, builtin.StringType}}

var ScanIntegerFunctionType = &core.Type{Parent: builtin.FunctionType, Name: "ScanIntegerFunction", Functions: nil, Generic: true, Generics: []*core.Type{builtin.IntegerType}}

type PrintFunction struct{}

func (*PrintFunction) Execute(scope *core.Scope) *core.Return {
	a, _ := scope.Get("text").Variable.(builtin.StringInterface)
	fmt.Print(a.GetStringValue())
	return &core.Return{Pointer: nil, ReturnType: core.NOTHING}
}

func (*PrintFunction) GetType() *core.Type {
	return PrintFunctionType
}

func (*PrintFunction) GetParameters() []*core.Parameter {
	return []*core.Parameter{{Name: "text", Typ: builtin.StringType, DefaultValue: &core.Pointer{Typ: builtin.StringType, Variable: &builtin.String{Value: ""}}}}
}

func (*PrintFunction) GetReturnType() *core.Type {
	return nil
}

type ScanIntegerFunction struct{}

func (*ScanIntegerFunction) Execute(scope *core.Scope) *core.Return {
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		return builtin.NewExceptionReturn(fmt.Sprintf("scanInteger: %v", err))
	}
	return &core.Return{Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: &builtin.Integer{Value: i}}, ReturnType: core.NOTHING}
}

func (*ScanIntegerFunction) GetType() *core.Type {
	return ScanIntegerFunctionType
}

func (*ScanIntegerFunction) GetParameters() []*core.Parameter {
	return []*core.Parameter{}
}

func (*ScanIntegerFunction) GetReturnType() *core.Type {
	return builtin.IntegerType
}

var printFunction core.Variable = &PrintFunction{}
var scanIntegerFunction core.Variable = &ScanIntegerFunction{}

var Block = core.NewScopeBlock(map[string]*core.Pointer{
	"print":       {Typ: PrintFunctionType, Variable: printFunction},
	"scanInteger": {Typ: ScanIntegerFunctionType, Variable: scanIntegerFunction},
})
