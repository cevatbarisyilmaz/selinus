package standard

import (
	_ "embed"
	"github.com/cevatbarisyilmaz/selinus/executer/module"
	"github.com/cevatbarisyilmaz/selinus/library/standard/native"
)

//go:embed self/standard.selinus
var rootFile string

var Module = &module.Module{
	NativeBlock: native.Block,
	RootFile:    rootFile,
	Name:        "standard.selinus",
}
