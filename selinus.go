package main

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler"
	"github.com/cevatbarisyilmaz/selinus/compiler/builtin"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"github.com/cevatbarisyilmaz/selinus/executer"
	"github.com/cevatbarisyilmaz/selinus/executer/module"
	"github.com/cevatbarisyilmaz/selinus/lexer"
	"github.com/cevatbarisyilmaz/selinus/library/standard"
	"github.com/cevatbarisyilmaz/selinus/parser"
	"github.com/cevatbarisyilmaz/selinus/reader"
	"os"
)

func main() {
	var fileName = "./examples/sum.selinus"
	stream, err := reader.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Reading error: %v", err)
		os.Exit(1)
	}
	lexTokens, err := lexer.Lex(stream, fileName)
	if err != nil {
		fmt.Printf("Scanning error: %v", err)
		os.Exit(1)
	}
	rootParseNode, err := parser.Parse(lexTokens)
	//printToken(rootParseNode, 0)
	//os.Exit(-1)
	if err != nil {
		fmt.Printf("Parsing error: %v", err)
		os.Exit(1)
	}
	scope := getInitialScope()
	rootCompileNode, err := compiler.Compile(rootParseNode, scope)
	if err != nil {
		fmt.Printf("Compile error: %v", err)
		os.Exit(1)
	}
	res := executer.Execute(rootCompileNode, scope)
	if res.ReturnType == core.EXCEPTION {
		fmt.Println(res.Pointer.Variable.(*builtin.String).GetStringValue())
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func printToken(token *parser.ParseNode, level int) {
	for i := level; i > 0; i-- {
		fmt.Print("\t")
	}
	fmt.Println("Token:")
	for i := level; i > 0; i-- {
		fmt.Print("\t")
	}
	fmt.Printf("%v", token)
	for i := level; i > 0; i-- {
		fmt.Print("\t")
	}
	fmt.Println()
	for i := level; i > 0; i-- {
		fmt.Print("\t")
	}
	fmt.Println("Children:")
	for _, e := range token.GetChildren() {
		printToken(e, level+1)
	}
	for t := token.Next(); t != nil; t = t.Next() {
		printToken(t, level)
	}
}

func getInitialScope() *core.Scope {
	scope := core.NewScope()
	scope.AddBlock(builtin.Block)
	importModule(standard.Module, scope)
	return scope
}

func importModule(module *module.Module, scope *core.Scope) *core.Return {
	scope.AddBlock(module.NativeBlock)
	stream := reader.ReadString(module.RootFile)
	lexTokens, err := lexer.Lex(stream, module.Name)
	if err != nil {
		fmt.Printf("Lexing error: %v", err)
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
