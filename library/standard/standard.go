package standard

import (
	_ "embed"
	"github.com/cevatbarisyilmaz/selinus/library/standard/native"
	"github.com/cevatbarisyilmaz/selinus/module"
)

//go:embed self/standard.selinus
var rootFile string

var Module = &module.Module{
	NativeBlock: native.Block,
	RootFile:    rootFile,
	Name:        "standard.selinus",
}
