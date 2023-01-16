package compiler

import (
	"errors"
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/compiler/builtin"
	"github.com/cevatbarisyilmaz/selinus/compiler/core"
	"github.com/cevatbarisyilmaz/selinus/parser"
	"strconv"
)

type VariableNode struct {
	name string
}

func (node *VariableNode) Execute(scope *core.Scope) *core.Return {
	scopeResult := scope.Get(node.name)
	if scopeResult.ReturnType == core.EXCEPTION {
		return scopeResult
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: scopeResult.Pointer}
}

type StringNode struct {
	value string
}

func (node *StringNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: core.NewStringPointer(node.value)}
}

type BooleanNode struct {
	value bool
}

func (node *BooleanNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: builtin.NewBooleanPointer(node.value)}
}

type FunctionNode struct {
	name       string
	lambda     bool
	returnType *core.Type
	parameters []*core.Parameter
	entryNode  core.Node
}

func (node *FunctionNode) Execute(scope *core.Scope) *core.Return {
	generics := []*core.Type{node.returnType}
	for _, parameter := range node.parameters {
		generics = append(generics, parameter.Typ)
	}
	typ := &core.Type{Name: "CustomFunction", Parent: builtin.FunctionType, Generic: true, Generics: generics}
	variable := core.NewVariable(&core.CustomFunction{Scope: scope.CloneWithName(node.name + "Function"), EntryNode: node.entryNode, Parameters: node.parameters, Typ: typ, ReturnType: node.returnType})
	if !node.lambda {
		scope.DeclareAndSet(node.name, &core.Pointer{Typ: typ, Variable: variable})
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: typ, Variable: variable}}
}

type IntegerNode struct {
	value int64
}

func (node *IntegerNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: builtin.NewIntegerPointer(node.value)}
}

type SetNode struct {
	leftSide  core.Node
	rightSide core.Node
}

func (node *SetNode) Execute(scope *core.Scope) *core.Return {
	r := node.rightSide.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	l := node.leftSide.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	l.Pointer.Variable = r.Pointer.Variable
	return &core.Return{ReturnType: core.NOTHING, Pointer: r.Pointer}
}

type DeclarationNode struct {
	typ        *core.Type
	identifier string
}

func (node *DeclarationNode) Execute(scope *core.Scope) *core.Return {
	p := &core.Pointer{Typ: node.typ, Variable: nil}
	scope.DeclareAndSet(node.identifier, p)
	return &core.Return{ReturnType: core.NOTHING, Pointer: p}
}

type ConditionNode struct {
	condition core.Node
	root      core.Node
}

func (node *ConditionNode) Execute(scope *core.Scope) *core.Return {
	internalReturn := node.condition.Execute(scope)
	if internalReturn.ReturnType != core.NOTHING {
		return internalReturn
	}
	if (internalReturn.Pointer.Variable).VariableInterface.(*builtin.Boolean).Value {
		scope.CreateBlock()
		defer scope.ReleaseBlock()
		current := node.root
		for current != nil {
			internalReturn = current.Execute(scope)
			if internalReturn.ReturnType != core.NOTHING {
				return internalReturn
			}
			current = current.Next()
		}
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
}

type ToLoopNode struct {
	from core.Node
	to   core.Node
	as   string
	root core.Node
}

func (node *ToLoopNode) Execute(scope *core.Scope) *core.Return {
	scope.CreateBlock()
	defer scope.ReleaseBlock()
	fromReturn := node.from.Execute(scope)
	if fromReturn.ReturnType != core.NOTHING {
		return fromReturn
	}
	fromReturnInteger := fromReturn.Pointer.Variable.ConvertTo(builtin.IntegerType)
	if fromReturnInteger.ReturnType != core.NOTHING {
		return fromReturn
	}
	from := fromReturnInteger.Pointer.Variable.VariableInterface.(*builtin.Integer).Value

	toReturn := node.to.Execute(scope)
	if toReturn.ReturnType != core.NOTHING {
		return toReturn
	}
	toReturnInteger := toReturn.Pointer.Variable.ConvertTo(builtin.IntegerType)
	if toReturnInteger.ReturnType != core.NOTHING {
		return toReturnInteger
	}
	to := toReturnInteger.Pointer.Variable.VariableInterface.(*builtin.Integer).Value

	if node.as != "" {
		scope.Declare(node.as, builtin.IntegerType)
	}

	for i := from; i <= to; i++ {
		if node.as != "" {
			scope.Set(node.as, builtin.NewIntegerPointer(i))
		}
		current := node.root
		for current != nil {
			internalReturn := current.Execute(scope)
			if internalReturn.ReturnType == core.BREAK {
				return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
			}
			if internalReturn.ReturnType == core.CONTINUE {
				break
			}
			if internalReturn.ReturnType != core.NOTHING {
				return internalReturn
			}
			current = current.Next()
		}
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
}

type ConditionLoopNode struct {
	condition core.Node
	root      core.Node
}

func (node *ConditionLoopNode) Execute(scope *core.Scope) *core.Return {
	scope.CreateBlock()
	defer scope.ReleaseBlock()
	for {
		internalReturn := node.condition.Execute(scope)
		if internalReturn.ReturnType != core.NOTHING {
			return internalReturn
		}
		if !(internalReturn.Pointer.Variable).VariableInterface.(*builtin.Boolean).Value {
			break
		}
		current := node.root
		for current != nil {
			internalReturn = current.Execute(scope)
			if internalReturn.ReturnType == core.BREAK {
				return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
			}
			if internalReturn.ReturnType == core.CONTINUE {
				break
			}
			if internalReturn.ReturnType != core.NOTHING {
				return internalReturn
			}
			current = current.Next()
		}
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
}

type CsvNode struct {
	children []core.Node
}

func (node *CsvNode) Execute(scope *core.Scope) *core.Return {
	var children []*core.Pointer
	for _, child := range node.children {
		subResult := child.Execute(scope)
		if subResult.ReturnType == core.EXCEPTION {
			return subResult
		}
		children = append(children, subResult.Pointer)
	}
	return &core.Return{
		ReturnType: core.NOTHING,
		Pointer:    core.NewSetPointer(children),
	}
}

type OrNode struct {
	left  core.Node
	right core.Node
}

func (node *OrNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	ll := l.Pointer.Variable.ConvertTo(builtin.BooleanType)
	if ll.ReturnType != core.NOTHING {
		return ll
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	rr := r.Pointer.Variable.ConvertTo(builtin.BooleanType)
	if rr.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Boolean{Value: ll.Pointer.Variable.VariableInterface.(*builtin.Boolean).Value || r.Pointer.Variable.VariableInterface.(*builtin.Boolean).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

type SummationNode struct {
	left  core.Node
	right core.Node
}

func (node *SummationNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Integer{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value + (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

type SubtractionNode struct {
	left  core.Node
	right core.Node
}

func (node *SubtractionNode) Execute(scope *core.Scope) *core.Return {
	var l *core.Return
	if node.left != nil {
		l = node.left.Execute(scope)
		if l.ReturnType != core.NOTHING {
			return l
		}
	} else {
		l = &core.Return{
			ReturnType: core.NOTHING,
			Pointer:    builtin.NewIntegerPointer(0),
		}
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Integer{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value - (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

type DivisionNode struct {
	left  core.Node
	right core.Node
}

func (node *DivisionNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	divider := (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value
	if divider == 0 {
		return core.NewExceptionReturn("division by zero")
	}
	variable := core.NewVariable(&builtin.Integer{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value / divider})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

type EqualityNode struct {
	left  core.Node
	right core.Node
}

func (node *EqualityNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Boolean{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value == (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

type GreaterNode struct {
	left  core.Node
	right core.Node
}

func (node *GreaterNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Boolean{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value > (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

type LessNode struct {
	left  core.Node
	right core.Node
}

func (node *LessNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := core.NewVariable(&builtin.Boolean{Value: (l.Pointer.Variable).VariableInterface.(*builtin.Integer).Value < (r.Pointer.Variable).VariableInterface.(*builtin.Integer).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

type ConcatenationNode struct {
	left  core.Node
	right core.Node
}

func (node *ConcatenationNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	ls := l.Pointer.Variable.ConvertTo(core.StringType)
	if ls.ReturnType != core.NOTHING {
		return ls
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	rs := r.Pointer.Variable.ConvertTo(core.StringType)
	if rs.ReturnType != core.NOTHING {
		return rs
	}
	variable := core.NewVariable(&core.String{Value: ls.Pointer.Variable.VariableInterface.(*core.String).Value + rs.Pointer.Variable.VariableInterface.(*core.String).Value})
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: core.StringType, Variable: variable}}
}

type FunctionCallNode struct {
	name       string
	parameters []core.Node
}

func (node *FunctionCallNode) Execute(localScope *core.Scope) *core.Return {
	scopeResult := localScope.Get(node.name)
	if scopeResult.ReturnType == core.EXCEPTION {
		return scopeResult
	}
	b := scopeResult.Pointer.Variable
	function := b.VariableInterface.(core.Function)
	functionScope := function.GetScope().Clone()
	functionScope.CreateBlock()
	defer functionScope.ReleaseBlock()
	i := 0
	for _, e := range node.parameters {
		t := e.Execute(localScope)
		if t.ReturnType != core.NOTHING {
			return t
		}
		functionScope.Declare(function.GetParameters()[i].Name, function.GetParameters()[i].Typ)
		functionScope.Set(function.GetParameters()[i].Name, t.Pointer)
		i++
	}
	for x := len(function.GetParameters()); i < x; i++ {
		functionScope.Declare(function.GetParameters()[i].Name, function.GetParameters()[i].Typ)
		functionScope.Set(function.GetParameters()[i].Name, function.GetParameters()[i].DefaultValue)
	}
	res := function.Execute(functionScope)
	return res
}

type ReturnNode struct {
	node core.Node
}

func (node *ReturnNode) Execute(scope *core.Scope) *core.Return {
	internalReturn := node.node.Execute(scope)
	if internalReturn.ReturnType != core.NOTHING {
		return internalReturn
	}
	return &core.Return{ReturnType: core.RETURN, Pointer: internalReturn.Pointer}
}

func Compile(node *parser.ParseNode, scope *core.Scope) (core.Node, error) {
	return parseBlock(node, scope, nil)
}

func parseBlock(node *parser.ParseNode, scope *core.Scope, expectedType *core.Type) (core.Node, error) {
	var root core.Node
	var prev core.Node
	var lastNode *parser.ParseNode
	for node != nil {
		current, _, err := createNode(node, scope, false, expectedType)
		if err != nil {
			return nil, err
		}
		if root == nil {
			root = current
		} else {
			prev.SetNext(current)
		}
		prev = current
		lastNode = node
		node = node.Next()
	}
	if expectedType != nil {
		_, b := prev.Root().(*ReturnNode)
		if !b {
			return nil, errors.New("expected return statement " + lastNode.GetMainToken().ToString())
		}
	}
	return root, nil
}

func createNode(node *parser.ParseNode, scope *core.Scope, conditional bool, expectedReturnType *core.Type) (core.Node, *core.Type, error) {
	nodeRoot, typ, err := createNodeRoot(node, scope, conditional, expectedReturnType)
	if err != nil {
		return nil, nil, err
	}
	return core.NewNode(nodeRoot, node.GetMainToken().ToString()), typ, nil
}

func createNodeRoot(node *parser.ParseNode, scope *core.Scope, conditional bool, expectedReturnType *core.Type) (core.NodeRoot, *core.Type, error) {
	switch node.GetType() {
	case parser.Less:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation < " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation < " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
		}
		return &LessNode{left: l, right: r}, builtin.BooleanType, nil
	case parser.Csv:
		var children []core.Node
		var childrenNodeType []*core.Type
		for _, child := range node.GetParseNodesWithKey(parser.Children) {
			childNode, childNodeType, err := createNode(child, scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			children = append(children, childNode)
			childrenNodeType = append(childrenNodeType, childNodeType)
		}
		return &CsvNode{children: children}, core.NewSetSubType(childrenNodeType), nil
	case parser.Or:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if lt.IsConvertable(builtin.BooleanType) && rt.IsConvertable(builtin.BooleanType) {
			return &OrNode{left: l, right: r}, builtin.BooleanType, nil
		}
		return nil, nil, errors.New("incompatible types for operation  " + node.MainLexicalToken.ToString())
	case parser.Summation:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if lt.IsConvertable(builtin.IntegerType) && rt.IsConvertable(builtin.IntegerType) {
			return &SummationNode{left: l, right: r}, builtin.IntegerType, nil
		}
		if lt.IsConvertable(core.StringType) && rt.IsConvertable(core.StringType) {
			return &ConcatenationNode{left: l, right: r}, core.StringType, nil
		}
		return nil, nil, errors.New("incompatible types " + lt.Name + " and " + rt.Name + " for operation  " + node.MainLexicalToken.ToString())
	case parser.Subtraction:
		var l core.Node
		var lt *core.Type
		var err error
		if node.GetParseNodesWithKey(parser.Children)[0] != nil {
			l, lt, err = createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			if !lt.IsCompatible(builtin.IntegerType) {
				return nil, nil, errors.New("incompatible type for operation - " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
			}
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation - " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
		}
		return &SubtractionNode{left: l, right: r}, builtin.IntegerType, nil
	case parser.Divide:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation / " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation / " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
		}
		return &DivisionNode{left: l, right: r}, builtin.IntegerType, nil
	case parser.Equal:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation == " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation == " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
		}
		return &EqualityNode{left: l, right: r}, builtin.BooleanType, nil
	case parser.Greater:
		l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation > " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation > " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
		}
		return &GreaterNode{left: l, right: r}, builtin.BooleanType, nil
	case parser.Variable:
		res := scope.Get(node.GetMainToken().GetValue())
		if res.ReturnType != core.NOTHING {
			return nil, nil, errors.New(node.GetMainToken().GetValue() + " is not declared " + node.GetMainToken().ToString())
		}
		return &VariableNode{name: node.GetMainToken().GetValue()}, res.Pointer.Typ, nil
	case parser.String:
		return &StringNode{value: node.GetMainToken().GetValue()}, core.StringType, nil
	case parser.Integer:
		i, _ := strconv.ParseInt(node.GetMainToken().GetValue(), 10, 64)
		return &IntegerNode{value: i}, builtin.IntegerType, nil
	case parser.FunctionCall:
		t := scope.MustGet(node.GetMainToken().GetValue())
		if t == nil {
			return nil, nil, errors.New("function " + node.GetMainToken().GetValue() + " is not defined " + node.GetMainToken().ToString())
		} else if !t.Typ.IsCompatible(builtin.FunctionType) {
			return nil, nil, errors.New(node.GetMainToken().GetValue() + " is not a function " + node.GetMainToken().ToString())
		}
		types := t.Typ.Generics
		parameters := types[1:]
		returnType := types[0]
		i := 0
		givenParametersLength := len(node.GetParseNodesWithKey(parser.Parameters))
		expectedParametersLength := len(parameters)
		if givenParametersLength > expectedParametersLength {
			return nil, nil, errors.New(fmt.Sprintf("too manny parameters %d-%d for function %s", givenParametersLength, expectedParametersLength, node.GetMainToken().ToString()))
		}
		if givenParametersLength < expectedParametersLength {
			return nil, nil, errors.New(fmt.Sprintf("not enough parameters %d-%d for function %s", givenParametersLength, expectedParametersLength, node.GetMainToken().ToString()))
		}
		suppliedParameters := make([]core.Node, 0)
		if givenParametersLength > 0 {
			for _, parameter := range node.GetParseNodesWithKey(parser.Parameters) {
				g, typ, err := createNode(parameter, scope, false, nil)
				if err != nil {
					return nil, nil, err
				}
				if !typ.IsCompatible(parameters[i]) {
					return nil, nil, errors.New("incompatible parameter type " + parameter.GetMainToken().ToString())
				}
				suppliedParameters = append(suppliedParameters, g)
				i++
			}

		}
		/*
			for x := expectedParametersLength; i < x; i++ {
				//p := parameters[i]
				//if p.defaultValue == nil{
				return nil, nil, errors.New("not enough parameters for function " + node.GetMainToken().ToString())
				//}
			}
		*/
		return &FunctionCallNode{name: node.GetMainToken().GetValue(), parameters: suppliedParameters}, returnType, nil
	case parser.Declaration:
		switch node.GetMainToken().GetValue() {
		case "var":
			scope.Declare(node.GetTokenWithKey(parser.Identifier).GetValue(), core.VariableType)
			return &DeclarationNode{typ: core.VariableType, identifier: node.GetTokenWithKey(parser.Identifier).GetValue()}, core.VariableType, nil
		case "string":
			scope.Declare(node.GetTokenWithKey(parser.Identifier).GetValue(), core.StringType)
			return &DeclarationNode{typ: core.StringType, identifier: node.GetTokenWithKey(parser.Identifier).GetValue()}, core.StringType, nil
		case "int":
			scope.Declare(node.GetTokenWithKey(parser.Identifier).GetValue(), builtin.IntegerType)
			return &DeclarationNode{typ: builtin.IntegerType, identifier: node.GetTokenWithKey(parser.Identifier).GetValue()}, builtin.IntegerType, nil
		}
		return nil, nil, errors.New("unknown declaration type " + node.GetMainToken().ToString())
	case parser.Gets:
		if conditional {
			l, lt, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			r, rt, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			if lt != builtin.IntegerType {
				return nil, nil, errors.New("incompatible type for operation = " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
			}
			if rt != builtin.IntegerType {
				return nil, nil, errors.New("incompatible type for operation = " + node.GetParseNodesWithKey(parser.Children)[1].GetMainToken().ToString())
			}
			return &EqualityNode{left: l, right: r}, builtin.BooleanType, nil
		}
		r, t2, err := createNode(node.GetParseNodesWithKey(parser.Children)[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		l, t1, err := createLeftSideForSet(node.GetParseNodesWithKey(parser.Children)[0], scope, t2)
		if err != nil {
			return nil, nil, err
		}
		if t2 == nil {
			return nil, nil, errors.New("right side does not return a variable " + node.GetMainToken().ToString())
		}
		if !t2.IsCompatible(t1) {
			return nil, nil, errors.New("incompatible types " + node.GetMainToken().ToString())
		}
		return &SetNode{rightSide: r, leftSide: l}, t2, nil
	case parser.If:
		condition, t1, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, true, nil)
		if err != nil {
			return nil, nil, err
		}
		if t1 != builtin.BooleanType {
			return nil, nil, errors.New("expected boolean " + node.GetParseNodesWithKey(parser.Children)[0].GetMainToken().ToString())
		}
		root, err := parseBlock(node.GetParseNodesWithKey(parser.Children)[1], scope, expectedReturnType)
		if err != nil {
			return nil, nil, err
		}
		return &ConditionNode{condition: condition, root: root}, nil, nil
	case parser.ToLoop:
		fromNode, t1, err := createNode(node.GetParseNodesWithKey(parser.From)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !t1.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("expected integer " + node.GetParseNodesWithKey(parser.From)[0].GetMainToken().ToString())
		}
		toNode, t2, err := createNode(node.GetParseNodesWithKey(parser.To)[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !t2.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("expected integer " + node.GetParseNodesWithKey(parser.To)[0].GetMainToken().ToString())
		}
		scope.CreateBlock()
		defer scope.ReleaseBlock()
		as := node.GetTokenWithKey(parser.Identifier)
		if as != nil {
			scope.Declare(as.GetValue(), builtin.IntegerType)
		}
		root, err := parseBlock(node.GetParseNodesWithKey(parser.Children)[0], scope, expectedReturnType)
		if err != nil {
			return nil, nil, err
		}
		return &ToLoopNode{
			from: fromNode,
			to:   toNode,
			as:   node.GetTokenWithKey(parser.Identifier).GetValue(),
			root: root,
		}, nil, nil
	case parser.Boolean:
		return &BooleanNode{value: node.GetMainToken().GetValue() == "true"}, builtin.BooleanType, nil
	case parser.Function:
		var parameters []*core.Parameter
		for _, child := range node.GetParseNodesWithKey(parser.Parameters) {
			p, err := parameterize(child, scope)
			if err != nil {
				return nil, nil, err
			}
			parameters = append(parameters, p)
		}
		var returnType *core.Type
		t3 := node.GetTokenWithKey(parser.ReturnType)
		if t3 != nil {
			t4 := scope.MustGet(t3.GetValue())
			if t4 == nil || !t4.Typ.IsCompatible(core.TypeType) {
				return nil, nil, errors.New("unknown return type for function " + t3.ToString())
			}
			returnType = (t4.Variable).VariableInterface.(*core.TypeVariable).Value
		}
		generics := []*core.Type{returnType}
		for _, parameter := range parameters {
			generics = append(generics, parameter.Typ)
		}
		scope.Declare(node.GetTokenWithKey(parser.Identifier).GetValue(), &core.Type{Name: node.GetTokenWithKey(parser.Identifier).GetValue(), Parent: builtin.FunctionType, Generic: true, Generics: generics})
		scope.CreateBlock()
		for _, parameter := range parameters {
			scope.Declare(parameter.Name, parameter.Typ)
		}
		root, err := parseBlock(node.GetParseNodesWithKey(parser.Children)[0], scope, returnType)
		if err != nil {
			return nil, nil, err
		}
		scope.ReleaseBlock()
		return &FunctionNode{name: node.GetTokenWithKey(parser.Identifier).GetValue(), lambda: false, parameters: parameters, returnType: returnType, entryNode: root}, builtin.FunctionType, nil
	case parser.Return:
		if len(node.GetParseNodesWithKey(parser.Children)) == 0 && expectedReturnType != nil {
			return nil, nil, errors.New("expected expression after " + node.GetMainToken().ToString())
		}
		temp, typ, err := createNode(node.GetParseNodesWithKey(parser.Children)[0], scope, false, expectedReturnType)
		if err != nil {
			return nil, nil, err
		}
		if expectedReturnType == nil {
			return nil, nil, errors.New("unexpected return statement " + node.GetMainToken().ToString())
		}
		if !typ.IsCompatible(expectedReturnType) {
			return nil, nil, errors.New("unexpected return type for the function " + node.GetMainToken().ToString())
		}
		return &ReturnNode{node: temp}, typ, nil
	}
	return nil, nil, errors.New("unknown node type " + node.GetMainToken().ToString())
}

func createLeftSideForSet(node *parser.ParseNode, scope *core.Scope, typ *core.Type) (core.Node, *core.Type, error) {
	if node.GetType() == parser.Variable && scope.Get(node.GetMainToken().GetValue()) == nil {
		scope.Declare(node.GetMainToken().GetValue(), typ)
	}
	return createNode(node, scope, false, nil)
}

func parameterize(node *parser.ParseNode, scope *core.Scope) (*core.Parameter, error) {
	var typ *core.Type
	switch node.GetMainToken().GetValue() {
	case "var":
		typ = core.VariableType
	case "string":
		typ = core.StringType
	case "int":
		typ = builtin.IntegerType
	default:
		return nil, errors.New("unknown parameter type " + node.GetMainToken().ToString())
	}
	//scope.Declare(node.GetToken2().ToString(), typ)
	return &core.Parameter{Typ: typ, Name: node.GetTokenWithKey(parser.Identifier).GetValue(), DefaultValue: nil}, nil
}
