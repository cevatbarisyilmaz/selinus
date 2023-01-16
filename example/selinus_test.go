package example_test

import (
	_ "embed"
	"github.com/cevatbarisyilmaz/selinus/library/standard/native"
	"github.com/cevatbarisyilmaz/selinus/runner"
	"strings"
	"testing"
)

//go:embed files/helloworld.selinus
var helloWorldTest string

//go:embed files/recursive_fibonacci.selinus
var recursiveFibonacciTest string

var examples = []*struct {
	testFileContent string
	testFilePath    string
	expectedOutput  string
}{
	{
		testFileContent: helloWorldTest,
		testFilePath:    "helloworld.selinus",
		expectedOutput:  "Hello, World!\n",
	},
	{
		testFileContent: recursiveFibonacciTest,
		testFilePath:    "recursive_fibonacci.selinus",
		expectedOutput:  "Fibonacci 0: 0\nFibonacci 1: 1\nFibonacci 2: 1\nFibonacci 3: 2\nFibonacci 4: 3\nFibonacci 5: 5\nFibonacci 6: 8\nFibonacci 7: 13\nFibonacci 8: 21\nFibonacci 9: 34\nFibonacci 10: 55\n",
	},
}

func TestExamples(t *testing.T) {
	builder := &strings.Builder{}
	native.SetDefaultOutputWriter(builder)
	for _, example := range examples {
		code := runner.Run(example.testFilePath, example.testFileContent)
		if code != 0 {
			t.Fatal(example.testFilePath, " output code is ", code)
		}
		output := builder.String()
		if output != example.expectedOutput {
			t.Fatalf("output mismatch, expected: %s, got: %s", example.expectedOutput, output)
		}
		builder.Reset()
	}
}
