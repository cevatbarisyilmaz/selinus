package runner

import (
	"bufio"
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler"
	"github.com/cevatbarisyilmaz/selinus/compiler/builtin"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"github.com/cevatbarisyilmaz/selinus/executer"
	"github.com/cevatbarisyilmaz/selinus/lexer"
	"github.com/cevatbarisyilmaz/selinus/library/standard"
	"github.com/cevatbarisyilmaz/selinus/module"
	"github.com/cevatbarisyilmaz/selinus/parser"
	"github.com/cevatbarisyilmaz/selinus/reader"
	"os"
)

func Run(filePath, fileContent string) int {
	var stream *bufio.Reader
	var err error
	if fileContent == "" {
		stream, err = reader.ReadFile(filePath)
	} else {
		stream = reader.ReadString(fileContent)
	}
	if err != nil {
		fmt.Printf("Read error: %v", err)
		return 1
	}
	lexTokens, err := lexer.Lex(stream, filePath)
	if err != nil {
		fmt.Printf("Scanning error: %v", err)
		return 1
	}
	rootParseNode, err := parser.Parse(lexTokens)
	if err != nil {
		fmt.Printf("Parsing error: %v", err)
		return 1
	}
	scope := getInitialScope()
	rootCompileNode, err := compiler.Compile(rootParseNode, scope)
	if err != nil {
		fmt.Printf("Compile error: %v", err)
		return 1
	}
	res := executer.Execute(rootCompileNode, scope)
	if res.ReturnType == core.EXCEPTION {
		fmt.Println(res.Pointer.Variable.(*core.StackTrace).GetStringValue())
		return 1
	} else {
		return 0
	}
}

func getInitialScope() *core.Scope {
	scope := core.NewScopeWithName("main")
	scope.AddBlock(builtin.Block)
	importModule(standard.Module, scope)
	scope.CreateBlock()
	return scope
}

func importModule(module *module.Module, scope *core.Scope) *core.Return {
	scope.AddBlock(module.NativeBlock)
	stream := reader.ReadString(module.RootFile)
	lexTokens, err := lexer.Lex(stream, module.Name)
	if err != nil {
		fmt.Printf("Scanning error: %v", err)
		os.Exit(-1)
	}
	rootParseNode, err := parser.Parse(lexTokens)
	if err != nil {
		fmt.Printf("Parsing error: %v", err)
		os.Exit(-1)
	}
	rootCompileNode, err := compiler.Compile(rootParseNode, scope)
	if err != nil {
		fmt.Printf("Compile error: %v", err)
		os.Exit(-1)
	}
	return executer.Execute(rootCompileNode, scope)
}
