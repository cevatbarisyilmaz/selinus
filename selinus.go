package main

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/parser"
	"github.com/cevatbarisyilmaz/selinus/runner"
	"os"
)

func main() {
	var filePath = "./example/files/recursive.selinus"
	os.Exit(runner.Run(filePath, ""))
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
