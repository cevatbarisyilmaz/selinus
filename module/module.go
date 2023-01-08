package module

import "github.com/cevatbarisyilmaz/selinus/compiler/core"

type Module struct {
	NativeBlock *core.ScopeBlock
	RootFile    string
	Name        string
}
