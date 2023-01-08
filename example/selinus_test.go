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

var examples = []*struct {
	testFileContent string
	testFilePath    string
	expectedOutput  string
}{
	{
		testFileContent: helloWorldTest,
		testFilePath:    "helloworld.selinus",
		expectedOutput:  "Hello World!\n",
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
