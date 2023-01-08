package executer

import (
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
)

func Execute(root core.Node, scope *core.Scope) *core.Return {
	return execute(root, scope)
}

func execute(root core.Node, scope *core.Scope) *core.Return {
	for root != nil {
		res := root.Execute(scope)
		if res.ReturnType == core.EXCEPTION {
			return res
		}
		root = root.Next()
	}
	return &core.Return{
		ReturnType: core.NOTHING,
		Pointer:    nil,
	}
}
