package main

import (
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/runner"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a target file")
		return
	}
	os.Exit(runner.Run(os.Args[1], ""))
}
