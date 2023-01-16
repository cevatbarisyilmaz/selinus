package builtin

import (
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
)

var Block = core.NewScopeBlock(map[string]*core.Pointer{
	"int":    {Typ: core.TypeType, Variable: core.TypeToVariable(IntegerType)},
	"bool":   {Typ: core.TypeType, Variable: core.TypeToVariable(BooleanType)},
	"string": {Typ: core.TypeType, Variable: core.TypeToVariable(StringType)},
	"func":   {Typ: core.TypeType, Variable: core.TypeToVariable(FunctionType)},
})

var scope = core.NewScope()

func init() {
	scope.AddBlock(Block)
}
