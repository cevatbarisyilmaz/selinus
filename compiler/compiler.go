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
	next core.Node
}

func (node *VariableNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: scope.Get(node.name)}
}

func (node *VariableNode) Next() core.Node {
	return node.next
}

func (node *VariableNode) SetNext(o core.Node) {
	node.next = o
}

type StringNode struct {
	value string
	next  core.Node
}

func (node *StringNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.StringType, Variable: &builtin.String{Value: node.value}}}
}

func (node *StringNode) Next() core.Node {
	return node.next
}

func (node *StringNode) SetNext(o core.Node) {
	node.next = o
}

type BooleanNode struct {
	value bool
	next  core.Node
}

func (node *BooleanNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: &builtin.Boolean{Value: node.value}}}
}

func (node *BooleanNode) Next() core.Node {
	return node.next
}

func (node *BooleanNode) SetNext(o core.Node) {
	node.next = o
}

type FunctionNode struct {
	name       string
	lambda     bool
	returnType *core.Type
	parameters []*core.Parameter
	entryNode  core.Node
	next       core.Node
}

func (node *FunctionNode) Execute(scope *core.Scope) *core.Return {
	generics := []*core.Type{node.returnType}
	for _, parameter := range node.parameters {
		generics = append(generics, parameter.Typ)
	}
	typ := &core.Type{Name: "CustomFunction", Parent: builtin.FunctionType, Generic: true, Generics: generics}
	variable := &core.CustomFunction{EntryNode: node.entryNode, Parameters: node.parameters, Typ: typ, ReturnType: node.returnType}
	if !node.lambda {
		scope.DeclareAndSet(node.name, &core.Pointer{Typ: typ, Variable: variable})
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: typ, Variable: variable}}
}

func (node *FunctionNode) Next() core.Node {
	return node.next
}

func (node *FunctionNode) SetNext(o core.Node) {
	node.next = o
}

type IntegerNode struct {
	value int
	next  core.Node
}

func (node *IntegerNode) Execute(scope *core.Scope) *core.Return {
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: &builtin.Integer{Value: node.value}}}
}

func (node *IntegerNode) Next() core.Node {
	return node.next
}

func (node *IntegerNode) SetNext(o core.Node) {
	node.next = o
}

type SetNode struct {
	leftSide  core.Node
	rightSide core.Node
	next      core.Node
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

func (node *SetNode) Next() core.Node {
	return node.next
}

func (node *SetNode) SetNext(o core.Node) {
	node.next = o
}

type DeclarationNode struct {
	typ        *core.Type
	identifier string
	next       core.Node
}

func (node *DeclarationNode) Execute(scope *core.Scope) *core.Return {
	p := &core.Pointer{Typ: node.typ, Variable: nil}
	scope.DeclareAndSet(node.identifier, p)
	return &core.Return{ReturnType: core.NOTHING, Pointer: p}
}

func (node *DeclarationNode) Next() core.Node {
	return node.next
}

func (node *DeclarationNode) SetNext(o core.Node) {
	node.next = o
}

type ConditionNode struct {
	condition core.Node
	root      core.Node
	next      core.Node
}

func (node *ConditionNode) Execute(scope *core.Scope) *core.Return {
	internalReturn := node.condition.Execute(scope)
	if internalReturn.ReturnType != core.NOTHING {
		return internalReturn
	}
	if (internalReturn.Pointer.Variable).(*builtin.Boolean).Value {
		scope.CreateBlock()
		current := node.root
		for current != nil {
			internalReturn = current.Execute(scope)
			if internalReturn.ReturnType != core.NOTHING {
				return internalReturn
			}
			current = current.Next()
		}
		scope.ReleaseBlock()
	}
	return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
}

func (node *ConditionNode) Next() core.Node {
	return node.next
}

func (node *ConditionNode) SetNext(o core.Node) {
	node.next = o
}

type LoopNode struct {
	condition core.Node
	root      core.Node
	next      core.Node
}

func (node *LoopNode) Execute(scope *core.Scope) *core.Return {
	scope.CreateBlock()
	for {
		internalReturn := node.condition.Execute(scope)
		if internalReturn.ReturnType != core.NOTHING {
			return internalReturn
		}
		if !(internalReturn.Pointer.Variable).(*builtin.Boolean).Value {
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
	scope.ReleaseBlock()
	return &core.Return{ReturnType: core.NOTHING, Pointer: nil}
}

func (node *LoopNode) Next() core.Node {
	return node.next
}

func (node *LoopNode) SetNext(o core.Node) {
	node.next = o
}

type CsvNode struct {
	children []core.Node
	next     core.Node
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

func (node CsvNode) Next() core.Node {
	return node.next
}

func (node CsvNode) SetNext(nextNode core.Node) {
	node.next = nextNode
}

type SummationNode struct {
	left  core.Node
	right core.Node
	next  core.Node
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
	variable := &builtin.Integer{Value: (l.Pointer.Variable).(*builtin.Integer).Value + (r.Pointer.Variable).(*builtin.Integer).Value}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

func (node *SummationNode) Next() core.Node {
	return node.next
}

func (node *SummationNode) SetNext(o core.Node) {
	node.next = o
}

type SubtractionNode struct {
	left  core.Node
	right core.Node
	next  core.Node
}

func (node *SubtractionNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := &builtin.Integer{Value: (l.Pointer.Variable).(*builtin.Integer).Value - (r.Pointer.Variable).(*builtin.Integer).Value}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

func (node *SubtractionNode) Next() core.Node {
	return node.next
}

func (node *SubtractionNode) SetNext(o core.Node) {
	node.next = o
}

type EqualityNode struct {
	left  core.Node
	right core.Node
	next  core.Node
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
	variable := &builtin.Boolean{Value: (l.Pointer.Variable).(*builtin.Integer).Value == (r.Pointer.Variable).(*builtin.Integer).Value}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

func (node *EqualityNode) Next() core.Node {
	return node.next
}

func (node *EqualityNode) SetNext(o core.Node) {
	node.next = o
}

type GreaterNode struct {
	left  core.Node
	right core.Node
	next  core.Node
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
	variable := &builtin.Boolean{Value: (l.Pointer.Variable).(*builtin.Integer).Value > (r.Pointer.Variable).(*builtin.Integer).Value}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

func (node *GreaterNode) Next() core.Node {
	return node.next
}

func (node *GreaterNode) SetNext(o core.Node) {
	node.next = o
}

type LessNode struct {
	left  core.Node
	right core.Node
	next  core.Node
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
	variable := &builtin.Boolean{Value: (l.Pointer.Variable).(*builtin.Integer).Value < (r.Pointer.Variable).(*builtin.Integer).Value}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.BooleanType, Variable: variable}}
}

func (node *LessNode) Next() core.Node {
	return node.next
}

func (node *LessNode) SetNext(o core.Node) {
	node.next = o
}

type ConcatenationNode struct {
	left  core.Node
	right core.Node
	next  core.Node
}

func (node *ConcatenationNode) Execute(scope *core.Scope) *core.Return {
	l := node.left.Execute(scope)
	if l.ReturnType != core.NOTHING {
		return l
	}
	r := node.right.Execute(scope)
	if r.ReturnType != core.NOTHING {
		return r
	}
	variable := &builtin.String{Value: (l.Pointer.Variable).(builtin.StringInterface).GetStringValue() + (r.Pointer.Variable).(builtin.StringInterface).GetStringValue()}
	return &core.Return{ReturnType: core.NOTHING, Pointer: &core.Pointer{Typ: builtin.IntegerType, Variable: variable}}
}

func (node *ConcatenationNode) Next() core.Node {
	return node.next
}

func (node *ConcatenationNode) SetNext(o core.Node) {
	node.next = o
}

type FunctionCallNode struct {
	name       string
	parameters []core.Node
	next       core.Node
}

func (node *FunctionCallNode) Execute(scope *core.Scope) *core.Return {
	a := scope.Get(node.name)
	b := a.Variable
	function := b.(core.Function)
	scope.CreateBlock()
	i := 0
	for _, e := range node.parameters {
		t := e.Execute(scope)
		if t.ReturnType != core.NOTHING {
			return t
		}
		scope.Declare(function.GetParameters()[i].Name, function.GetParameters()[i].Typ)
		scope.Set(function.GetParameters()[i].Name, t.Pointer)
		i++
	}
	for x := len(function.GetParameters()); i < x; i++ {
		scope.Declare(function.GetParameters()[i].Name, function.GetParameters()[i].Typ)
		scope.Set(function.GetParameters()[i].Name, function.GetParameters()[i].DefaultValue)
	}
	return function.Execute(scope)
}

func (node *FunctionCallNode) Next() core.Node {
	return node.next
}

func (node *FunctionCallNode) SetNext(o core.Node) {
	node.next = o
}

type ReturnNode struct {
	node core.Node
	next core.Node
}

func (node *ReturnNode) Execute(scope *core.Scope) *core.Return {
	internalReturn := node.node.Execute(scope)
	if internalReturn.ReturnType != core.NOTHING {
		return internalReturn
	}
	return &core.Return{ReturnType: core.RETURN, Pointer: internalReturn.Pointer}
}

func (node *ReturnNode) Next() core.Node {
	return node.next
}

func (node *ReturnNode) SetNext(o core.Node) {
	node.next = o
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
		_, b := prev.(*ReturnNode)
		if !b {
			return nil, errors.New("expected return statement " + lastNode.GetToken().ToString())
		}
	}
	return root, nil
}

func createNode(node *parser.ParseNode, scope *core.Scope, conditional bool, expectedReturnType *core.Type) (core.Node, *core.Type, error) {
	switch node.GetType() {
	case parser.Csv:
		var children []core.Node
		var childrenNodeType []*core.Type
		for _, child := range node.GetChildren() {
			childNode, childNodeType, err := createNode(child, scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			children = append(children, childNode)
			childrenNodeType = append(childrenNodeType, childNodeType)
		}
		return &CsvNode{children: children}, core.NewSetSubType(childrenNodeType), nil
	case parser.Summation:
		l, lt, err := createNode(node.GetChildren()[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetChildren()[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if lt.IsCompatible(builtin.IntegerType) && rt.IsCompatible(builtin.IntegerType) {
			return &SummationNode{left: l, right: r}, builtin.IntegerType, nil
		}
		if lt.IsCompatible(builtin.StringType) && rt.IsCompatible(builtin.StringType) {
			return &ConcatenationNode{left: l, right: r}, builtin.StringType, nil
		}
		if !lt.IsCompatible(builtin.StringType) {
			return nil, nil, errors.New("incompatible type for operation  " + node.GetChildren()[0].GetToken().ToString())
		}
		return nil, nil, errors.New("incompatible type for operation  " + node.GetChildren()[1].GetToken().ToString())
	case parser.Subtraction:
		l, lt, err := createNode(node.GetChildren()[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetChildren()[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation - " + node.GetChildren()[0].GetToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation - " + node.GetChildren()[1].GetToken().ToString())
		}
		return &SubtractionNode{left: l, right: r}, builtin.IntegerType, nil
	case parser.Equal:
		l, lt, err := createNode(node.GetChildren()[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetChildren()[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation == " + node.GetChildren()[0].GetToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation == " + node.GetChildren()[1].GetToken().ToString())
		}
		return &EqualityNode{left: l, right: r}, builtin.BooleanType, nil
	case parser.Greater:
		l, lt, err := createNode(node.GetChildren()[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		r, rt, err := createNode(node.GetChildren()[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if !lt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation > " + node.GetChildren()[0].GetToken().ToString())
		}
		if !rt.IsCompatible(builtin.IntegerType) {
			return nil, nil, errors.New("incompatible type for operation > " + node.GetChildren()[1].GetToken().ToString())
		}
		return &GreaterNode{left: l, right: r}, builtin.BooleanType, nil
	case parser.Variable:
		if scope.Get(node.GetToken().GetValue()) == nil {
			return nil, nil, errors.New(node.GetToken().GetValue() + " is not declared " + node.GetToken().ToString())
		}
		return &VariableNode{name: node.GetToken().GetValue()}, scope.Get(node.GetToken().GetValue()).Typ, nil
	case parser.String:
		return &StringNode{value: node.GetToken().GetValue()}, builtin.StringType, nil
	case parser.Integer:
		i, _ := strconv.Atoi(node.GetToken().GetValue())
		return &IntegerNode{value: i}, builtin.IntegerType, nil
	case parser.FunctionCall:
		t := scope.Get(node.GetToken().GetValue())
		if t == nil {
			return nil, nil, errors.New("function " + node.GetToken().GetValue() + " is not defined " + node.GetToken().ToString())
		} else if !t.Typ.IsCompatible(builtin.FunctionType) {
			return nil, nil, errors.New(node.GetToken().GetValue() + " is not a function " + node.GetToken().ToString())
		}
		types := t.Typ.Generics
		parameters := types[1:]
		returnType := types[0]
		i := 0
		givenParametersLength := len(node.GetParameters())
		expectedParametersLength := len(parameters)
		if givenParametersLength > expectedParametersLength {
			return nil, nil, errors.New(fmt.Sprintf("too manny parameters %d-%d for function %s", givenParametersLength, expectedParametersLength, node.GetToken().ToString()))
		}
		if givenParametersLength < expectedParametersLength {
			return nil, nil, errors.New(fmt.Sprintf("not enough parameters %d-%d for function %s", givenParametersLength, expectedParametersLength, node.GetToken().ToString()))
		}
		suppliedParameters := make([]core.Node, 0)
		if givenParametersLength > 0 {
			for _, parameter := range node.GetParameters() {
				g, typ, err := createNode(parameter, scope, false, nil)
				if err != nil {
					return nil, nil, err
				}
				if !typ.IsCompatible(parameters[i]) {
					return nil, nil, errors.New("incompatible parameter type " + parameter.GetToken().ToString())
				}
				suppliedParameters = append(suppliedParameters, g)
				i++
			}

		}
		/*
			for x := expectedParametersLength; i < x; i++ {
				//p := parameters[i]
				//if p.defaultValue == nil{
				return nil, nil, errors.New("not enough parameters for function " + node.GetToken().ToString())
				//}
			}
		*/
		return &FunctionCallNode{name: node.GetToken().GetValue(), parameters: suppliedParameters}, returnType, nil
	case parser.Declaration:
		switch node.GetToken().GetValue() {
		case "var":
			scope.Declare(node.GetToken2().GetValue(), builtin.VariableType)
			return &DeclarationNode{typ: builtin.VariableType, identifier: node.GetToken2().GetValue()}, builtin.VariableType, nil
		case "string":
			scope.Declare(node.GetToken2().GetValue(), builtin.StringType)
			return &DeclarationNode{typ: builtin.StringType, identifier: node.GetToken2().GetValue()}, builtin.StringType, nil
		case "int":
			scope.Declare(node.GetToken2().GetValue(), builtin.IntegerType)
			return &DeclarationNode{typ: builtin.IntegerType, identifier: node.GetToken2().GetValue()}, builtin.IntegerType, nil
		}
		return nil, nil, errors.New("unknown declaration type " + node.GetToken().ToString())
	case parser.Gets:
		if conditional {
			l, lt, err := createNode(node.GetChildren()[0], scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			r, rt, err := createNode(node.GetChildren()[1], scope, false, nil)
			if err != nil {
				return nil, nil, err
			}
			if lt != builtin.IntegerType {
				return nil, nil, errors.New("incompatible type for operation = " + node.GetChildren()[0].GetToken().ToString())
			}
			if rt != builtin.IntegerType {
				return nil, nil, errors.New("incompatible type for operation = " + node.GetChildren()[1].GetToken().ToString())
			}
			return &EqualityNode{left: l, right: r}, builtin.BooleanType, nil
		}
		r, t2, err := createNode(node.GetChildren()[1], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		l, t1, err := createLeftSideForSet(node.GetChildren()[0], scope, t2)
		if err != nil {
			return nil, nil, err
		}
		if t2 == nil {
			return nil, nil, errors.New("right side does not return a variable " + node.GetToken().ToString())
		}
		if !t2.IsCompatible(t1) {
			return nil, nil, errors.New("incompatible types " + node.GetToken().ToString())
		}
		return &SetNode{rightSide: r, leftSide: l}, t2, nil
	case parser.If:
		condition, t1, err := createNode(node.GetChildren()[0], scope, true, nil)
		if err != nil {
			return nil, nil, err
		}
		if t1 != builtin.BooleanType {
			return nil, nil, errors.New("expected boolean " + node.GetChildren()[0].GetToken().ToString())
		}
		root, err := parseBlock(node.GetChildren()[1], scope, nil)
		return &ConditionNode{condition: condition, root: root}, nil, nil
	case parser.While:
		condition, t1, err := createNode(node.GetChildren()[0], scope, true, nil)
		if err != nil {
			return nil, nil, err
		}
		if t1 != builtin.BooleanType {
			return nil, nil, errors.New("expected boolean " + node.GetChildren()[0].GetToken().ToString())
		}
		root, err := parseBlock(node.GetChildren()[1], scope, nil)
		return &LoopNode{condition: condition, root: root}, nil, nil
	case parser.Boolean:
		return &BooleanNode{value: node.GetToken().GetValue() == "true"}, builtin.BooleanType, nil
	case parser.Function:
		var parameters []*core.Parameter
		for _, child := range node.GetParameters() {
			p, err := parameterize(child, scope)
			if err != nil {
				return nil, nil, err
			}
			parameters = append(parameters, p)
		}
		var returnType *core.Type
		t3 := node.GetToken3()
		if t3.GetValue() != "" {
			t4 := scope.Get(t3.GetValue())
			if t4 == nil || !t4.Typ.IsCompatible(core.TypeType) {
				return nil, nil, errors.New("unknown return type for function " + t3.ToString())
			}
			returnType = (t4.Variable).(*core.TypeVariable).Value
		}
		generics := []*core.Type{returnType}
		for _, parameter := range parameters {
			generics = append(generics, parameter.Typ)
		}
		scope.Declare(node.GetToken2().GetValue(), &core.Type{Name: node.GetToken2().GetValue(), Parent: builtin.FunctionType, Generic: true, Generics: generics})
		scope.CreateBlock()
		for _, parameter := range parameters {
			scope.Declare(parameter.Name, parameter.Typ)
		}
		root, err := parseBlock(node.GetChildren()[0], scope, returnType)
		if err != nil {
			return nil, nil, err
		}
		scope.ReleaseBlock()
		return &FunctionNode{name: node.GetToken2().GetValue(), lambda: false, parameters: parameters, returnType: returnType, entryNode: root}, builtin.FunctionType, nil
	case parser.Return:
		temp, typ, err := createNode(node.GetChildren()[0], scope, false, nil)
		if err != nil {
			return nil, nil, err
		}
		if expectedReturnType == nil {
			return nil, nil, errors.New("unexpected return statement " + node.GetToken().ToString())
		}
		if !typ.IsCompatible(expectedReturnType) {
			return nil, nil, errors.New("unexpected return type for the function " + node.GetToken().ToString())
		}
		return &ReturnNode{node: temp}, typ, nil
	}
	return nil, nil, errors.New("unknown node type " + node.GetToken().ToString())
}

func createLeftSideForSet(node *parser.ParseNode, scope *core.Scope, typ *core.Type) (core.Node, *core.Type, error) {
	if node.GetType() == parser.Variable && scope.Get(node.GetToken().GetValue()) == nil {
		scope.Declare(node.GetToken().GetValue(), typ)
	}
	return createNode(node, scope, false, nil)
}

func parameterize(node *parser.ParseNode, scope *core.Scope) (*core.Parameter, error) {
	var typ *core.Type
	switch node.GetToken().GetValue() {
	case "var":
		typ = builtin.VariableType
	case "string":
		typ = builtin.StringType
	case "int":
		typ = builtin.IntegerType
	default:
		return nil, errors.New("unknown parameter type " + node.GetToken().ToString())
	}
	//scope.Declare(node.GetToken2().ToString(), typ)
	return &core.Parameter{Typ: typ, Name: node.GetToken2().GetValue(), DefaultValue: nil}, nil
}
