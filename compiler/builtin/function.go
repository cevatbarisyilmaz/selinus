package builtin

import "github.com/cevatbarisyilmaz/selinus/compiler/core"

var FunctionType = &core.Type{Name: "Function", Parent: core.VariableType, Methods: map[string]core.Function{}, Converters: map[*core.Type]core.Function{}, Scope: core.NewScope()}
