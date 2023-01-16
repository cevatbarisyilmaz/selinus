package native

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler/builtin"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"io"
	"os"
)

var PrintFunctionType = &core.Type{Parent: builtin.FunctionType, Name: "print", Methods: nil, Generic: true, Generics: []*core.Type{nil, core.StringType}}

var ScanIntegerFunctionType = &core.Type{Parent: builtin.FunctionType, Name: "scanInteger", Methods: nil, Generic: true, Generics: []*core.Type{builtin.IntegerType}}

var defaultOutputWriter io.Writer = os.Stdout

func SetDefaultOutputWriter(writer io.Writer) {
	defaultOutputWriter = writer
}

type PrintFunction struct{}

func (*PrintFunction) Execute(scope *core.Scope) *core.Return {
	scopeResult := scope.Get("text")
	if scopeResult.ReturnType != core.NOTHING {
		return scopeResult
	}
	r := scopeResult.Pointer.Variable.ConvertTo(core.StringType)
	if r.ReturnType != core.NOTHING {
		return r
	}
	_, err := fmt.Fprint(defaultOutputWriter, r.Pointer.Variable.VariableInterface.(*core.String).Value)
	if err != nil {
		return core.NewExceptionReturn("print failed: " + err.Error())
	}
	return &core.Return{Pointer: nil, ReturnType: core.NOTHING}
}

func (*PrintFunction) GetType() *core.Type {
	return PrintFunctionType
}

func (*PrintFunction) GetParameters() []*core.Parameter {
	return []*core.Parameter{{Name: "text", Typ: core.StringType, DefaultValue: core.NewStringPointer("")}}
}

func (*PrintFunction) GetReturnType() *core.Type {
	return nil
}

func (*PrintFunction) GetScope() *core.Scope {
	return scope
}

type ScanIntegerFunction struct{}

func (*ScanIntegerFunction) Execute(scope *core.Scope) *core.Return {
	var i int64
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		return core.NewExceptionReturn(fmt.Sprintf("scanInteger: %v", err))
	}
	return &core.Return{Pointer: builtin.NewIntegerPointer(i), ReturnType: core.NOTHING}
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

func (*ScanIntegerFunction) GetScope() *core.Scope {
	return scope
}

var printFunction core.VariableInterface = &PrintFunction{}
var scanIntegerFunction core.VariableInterface = &ScanIntegerFunction{}

var scope = core.NewScopeWithName("native")

func init() {
	scope.AddBlock(Block)
}

var Block = core.NewScopeBlock(map[string]*core.Pointer{
	"print":       {Typ: PrintFunctionType, Variable: core.NewVariable(printFunction)},
	"scanInteger": {Typ: ScanIntegerFunctionType, Variable: core.NewVariable(scanIntegerFunction)},
})
