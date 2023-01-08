package parser

import (
	"errors"
	"github.com/cevatbarisyilmaz/selinus/lexer"
	"strconv"
)

type ParseToken struct {
	Group  bool
	Tokens []ParseToken
	Token  lexer.LexicalToken
}

type ParseNodeType int

const (
	String ParseNodeType = iota
	Integer
	Boolean
	Function
	Summation
	Subtraction
	Equal
	Greater
	Less
	FunctionCall
	Declaration
	Gets
	Variable
	If
	While
	Return
	Csv
)

type ParseNode struct {
	NodeType           ParseNodeType
	Children           []*ParseNode
	Parameters         []*ParseNode
	next               *ParseNode
	LexicelToken       lexer.LexicalToken
	secondLexicalToken lexer.LexicalToken
	thirdLexicalToken  lexer.LexicalToken
}

func (node *ParseNode) GetType() ParseNodeType {
	return node.NodeType
}

func (node *ParseNode) SetType(typ ParseNodeType) {
	node.NodeType = typ
}

func (node *ParseNode) GetToken() lexer.LexicalToken {
	return node.LexicelToken
}

func (node *ParseNode) GetToken2() lexer.LexicalToken {
	return node.secondLexicalToken
}

func (node *ParseNode) GetToken3() lexer.LexicalToken {
	return node.thirdLexicalToken
}

func (node *ParseNode) GetChildren() []*ParseNode {
	return node.Children
}

func (node *ParseNode) GetParameters() []*ParseNode {
	return node.Parameters
}

func (node *ParseNode) Next() *ParseNode {
	return node.next
}

func Parse(tokens []lexer.LexicalToken) (*ParseNode, error) {
	parseTokens, err := group(tokens)
	if err != nil {
		return nil, err
	}
	statements := divide(parseTokens)
	root, err := createParseNodes(statements)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func group(tokens []lexer.LexicalToken) ([]ParseToken, error) {
	var stack []ParseToken
	var inside []lexer.LexicalToken
	level := 0
	for _, e := range tokens {
		if e.GetType() == lexer.LeftParenthesis {
			if level > 0 {
				inside = append(inside, e)
			}
			level++
		} else if e.GetType() == lexer.RightParenthesis {
			if level > 0 {
				level--
				if level == 0 {
					t, err := group(inside)
					if err != nil {
						return nil, err
					}
					inside = make([]lexer.LexicalToken, 0)
					stack = append(stack, ParseToken{Group: true, Tokens: t, Token: e})
				} else {
					inside = append(inside, e)
				}
			} else {
				return nil, errors.New("unexpected right parenthesis at line " + strconv.Itoa(e.GetLine()) + " position " + strconv.Itoa(e.GetPosition()))
			}
		} else if level == 0 {
			stack = append(stack, ParseToken{Group: false, Token: e})
		} else if e.GetType() != lexer.NewLine {
			inside = append(inside, e)
		}
	}
	if level > 0 {
		e := tokens[len(tokens)-1]
		return nil, errors.New("expected right parenthesis at line " + strconv.Itoa(e.GetLine()) + " position " + strconv.Itoa(e.GetPosition()+1))
	}
	return stack, nil
}

func divide(tokens []ParseToken) [][]ParseToken {
	var statements [][]ParseToken
	var statement []ParseToken
	expecting := false
	for _, e := range tokens {
		if e.Group {
			expecting = false
			statement = append(statement, e)
		} else if e.Token.GetType() == lexer.SemiColon || !expecting && e.Token.GetType() == lexer.NewLine {
			expecting = false
			if len(statement) > 0 {
				statements = append(statements, statement)
				statement = make([]ParseToken, 0)
			}
		} else if e.Token.GetType() == lexer.Operator && e.Token.GetValue() != lexer.Increase && e.Token.GetValue() != lexer.Decrease {
			expecting = true
			statement = append(statement, e)
		} else {
			expecting = false
			statement = append(statement, e)
		}
	}
	if len(statement) > 0 {
		statements = append(statements, statement)
	}
	return statements
}

func createParseNodes(statements [][]ParseToken) (*ParseNode, error) {
	node, _, err := formBlock(statements, 0)
	return node, err
}

func getPrecedence(token ParseToken) int {
	if token.Group {
		return 8
	}
	switch token.Token.GetType() {
	case lexer.Keyword:
		switch token.Token.GetValue() {
		case lexer.If:
			fallthrough
		case lexer.While:
			fallthrough
		case lexer.End:
			fallthrough
		case lexer.Else:
			fallthrough
		case lexer.Return:
			return 1
		case lexer.True:
			fallthrough
		case lexer.False:
			return 7
		}
		return 4
	case lexer.Operator:
		switch token.Token.GetValue() {
		case lexer.Gets:
			return 2
		case lexer.Equal:
			fallthrough
		case lexer.Greater:
			fallthrough
		case lexer.GreaterOrEqual:
			fallthrough
		case lexer.Less:
			fallthrough
		case lexer.LessOrEqual:
			return 5
		case lexer.Plus:
			fallthrough
		case lexer.Minus:
			return 6
		case lexer.Multiply:
			fallthrough
		case lexer.Divide:
			return 7
		}
	case lexer.Coma:
		return 3
	}
	return 8
}

func formBlock(statements [][]ParseToken, i int) (*ParseNode, int, error) {
	var root *ParseNode
	var temp *ParseNode
	var pre *ParseNode
	var err error
	for length := len(statements); i < length; i++ {
		if statements[i][0].Token.GetType() == lexer.Keyword && statements[i][0].Token.GetValue() == lexer.End {
			return root, i, nil
		}
		temp, err = formParseNode(statements[i])
		if err != nil {
			return nil, i, err
		}
		if root == nil {
			root = temp
		}
		if pre != nil {
			pre.next = temp
		}
		pre = temp
		if temp.NodeType == Return {
			return root, i, nil
		}
		if temp.NodeType == If || temp.NodeType == While || temp.NodeType == Function {
			child, a, err := formBlock(statements, i+1)
			i = a
			if err != nil {
				return nil, i, err
			}
			temp.Children = append(temp.Children, child)
		}
	}
	return root, i, nil
}

func formParseNode(tokens []ParseToken) (*ParseNode, error) {
	currentPrecedence := -1
	var currentIndex int
	for i, t := range tokens {
		precedence := getPrecedence(t)
		if currentPrecedence == -1 || precedence < currentPrecedence {
			currentPrecedence = precedence
			currentIndex = i
		}
	}
	if currentPrecedence == -1 {
		return nil, nil
	}
	t := tokens[currentIndex]
	t2 := t.Token
	if t.Group {
		if len(tokens) != 1 {
			return nil, errors.New("unexpected Token " + tokens[1].Token.ToString() + " / 1")
		}
		return formParseNode(t.Tokens)
	}
	typ := t2.GetType()
	switch typ {
	case lexer.Coma:
		if currentIndex == 0 {
			return nil, errors.New("expected token before coma " + t2.ToString())
		} else if currentIndex == len(tokens)-1 {
			return nil, errors.New("expected token after coma " + t2.ToString())
		}
		leftChild, err := formParseNode(tokens[:currentIndex])
		if err != nil {
			return nil, err
		}
		rightChild, err := formParseNode(tokens[currentIndex+1:])
		if err != nil {
			return nil, err
		}
		children := []*ParseNode{leftChild}
		if rightChild.NodeType == Csv {
			children = append(children, rightChild.Children...)
		} else {
			children = append(children, rightChild)
		}
		return &ParseNode{
			NodeType:     Csv,
			Children:     children,
			next:         nil,
			LexicelToken: t2,
		}, nil
	case lexer.Operator:
		if currentIndex == 0 {
			return nil, errors.New("expected token before operation " + t2.ToString())
		} else if currentIndex == len(tokens)-1 {
			return nil, errors.New("expected token after operation " + t2.ToString())
		}
		leftChild, err := formParseNode(tokens[:currentIndex])
		if err != nil {
			return nil, err
		}
		rightChild, err := formParseNode(tokens[currentIndex+1:])
		if err != nil {
			return nil, err
		}
		var nodeType ParseNodeType
		switch t2.GetValue() {
		case lexer.Gets:
			nodeType = Gets
		case lexer.Plus:
			nodeType = Summation
		case lexer.Minus:
			nodeType = Subtraction
		case lexer.Equal:
			nodeType = Equal
		case lexer.Greater:
			nodeType = Greater
		case lexer.Less:
			nodeType = Less
		}
		return &ParseNode{NodeType: nodeType, Children: []*ParseNode{leftChild, rightChild}, next: nil, LexicelToken: t2}, nil
	case lexer.Text:
		if len(tokens) != 1 {
			if len(tokens) != 1 {
				return nil, errors.New("unexpected Token " + tokens[1].Token.ToString() + " / 2")
			}
		}
		return &ParseNode{NodeType: String, Children: nil, next: nil, LexicelToken: t2}, nil
	case lexer.Integer:
		if len(tokens) != 1 {
			return nil, errors.New("unexpected Token " + tokens[1].Token.ToString() + " / 3")
		}
		return &ParseNode{NodeType: Integer, Children: nil, next: nil, LexicelToken: t2}, nil
	case lexer.Identifier:
		if len(tokens) > 1 {
			if !tokens[1].Group && tokens[1].Token.GetType() == lexer.Identifier {
				if len(tokens) > 2 {
					return nil, errors.New("was not expecting more tokens after variable declaration " + tokens[2].Token.ToString())
				}
				return &ParseNode{NodeType: Declaration, LexicelToken: t2, secondLexicalToken: tokens[1].Token}, nil
			}
			if tokens[1].Group {
				if len(tokens) > 2 {
					return nil, errors.New("unexpected Token " + tokens[2].Token.ToString() + " / 4")
				}
				parameter, err := formParseNode(tokens[1].Tokens)
				if err != nil {
					return nil, err
				}
				parameters := make([]*ParseNode, 0)
				if parameter != nil {
					if parameter.NodeType == Csv {
						parameters = parameter.Children
					} else {
						parameters = []*ParseNode{parameter}
					}
				}
				return &ParseNode{NodeType: FunctionCall, Parameters: parameters, next: nil, LexicelToken: t2}, nil
			}
			return nil, errors.New("unexpected token, was expecting a parenthesis: " + tokens[1].Token.ToString())
		}
		return &ParseNode{NodeType: Variable, Children: nil, next: nil, LexicelToken: t2}, nil
	case lexer.Keyword:
		loop := false
		switch t2.GetValue() {
		case lexer.While:
			loop = true
			fallthrough
		case lexer.If:
			if len(tokens) > 1 {
				condition, err := formParseNode(tokens[currentIndex+1:])
				if err != nil {
					return nil, err
				}
				typ := If
				if loop {
					typ = While
				}
				return &ParseNode{NodeType: typ, Children: []*ParseNode{condition}, next: nil, LexicelToken: t2}, nil
			}
			return nil, errors.New("expected a condition after keyword if " + t2.ToString())
		case lexer.Function:
			if len(tokens) == 3 || len(tokens) == 4 {
				base := 1
				var third lexer.LexicalToken
				if tokens[2].Token.GetType() == lexer.Identifier {
					if tokens[1].Token.GetType() != lexer.Identifier && !(tokens[1].Token.GetType() == lexer.Keyword &&
						tokens[1].Token.GetValue() == "int") {
						return nil, errors.New("expected identifier " + tokens[1].Token.ToString())
					}
					base = 2
					third = tokens[1].Token
				}
				if tokens[base].Token.GetType() == lexer.Identifier {
					if tokens[base+1].Group {
						parametersNode, err := formParseNode(tokens[base+1].Tokens)
						if err != nil {
							return nil, err
						}
						var children []*ParseNode
						switch parametersNode.NodeType {
						case Declaration:
							children = append(children, parametersNode)
						case Csv:
							for _, child := range parametersNode.Children {
								if child.NodeType != Declaration {
									return nil, errors.New("was expecting parameter declaration " + child.LexicelToken.ToString())
								}
								children = append(children, child)
							}
						default:
							return nil, errors.New("was expecting parameter declaration " + parametersNode.LexicelToken.ToString())
						}
						return &ParseNode{NodeType: Function, LexicelToken: t2, secondLexicalToken: tokens[base].Token, thirdLexicalToken: third, Parameters: children}, nil
					}
					return nil, errors.New("expected parameters " + tokens[base+1].Token.ToString())
				}
				return nil, errors.New("expected identifier after keyword function " + t2.ToString())
			}
			return nil, errors.New("non complete function declaration")
		case lexer.True:
			fallthrough
		case lexer.False:
			return &ParseNode{NodeType: Boolean, LexicelToken: t2}, nil
		case lexer.Return:
			temp, err := formParseNode(tokens[1:])
			if err != nil {
				return nil, err
			}
			return &ParseNode{NodeType: Return, LexicelToken: t2, Children: []*ParseNode{temp}}, nil
		}
	}
	return nil, nil
}
