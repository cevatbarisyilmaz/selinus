package parser

import (
	"errors"
	"fmt"
	"github.com/cevatbarisyilmaz/selinus/lexer"
	"github.com/cevatbarisyilmaz/selinus/util"
	"strconv"
)

type ParseToken struct {
	Group            bool
	Tokens           []*ParseToken
	Token            *lexer.LexicalToken
	LeftParenthesis  *lexer.LexicalToken
	RightParenthesis *lexer.LexicalToken
}

func (parseToken *ParseToken) GetStartPosition() string {
	if !parseToken.Group {
		return parseToken.Token.ToString()
	}
	return parseToken.LeftParenthesis.ToString()
}

func (parseToken *ParseToken) GetEndPosition() string {
	if !parseToken.Group {
		return parseToken.Token.ToString()
	}
	return parseToken.RightParenthesis.ToString()
}

type ParseNodeType int

const (
	Children   = "children"
	Identifier = "identifier"
	ReturnType = "return type"
	Parameters = "parameters"
	To         = "to"
	From       = "from"
)

const (
	String ParseNodeType = iota
	Integer
	Boolean
	Function
	Summation
	Subtraction
	Divide
	Equal
	Greater
	Less
	Or
	FunctionCall
	Declaration
	Gets
	Variable
	If
	ToLoop
	Return
	Csv
)

type ParseNode struct {
	NodeType           ParseNodeType
	ParseNodes         map[string][]*ParseNode
	next               *ParseNode
	MainLexicalToken   *lexer.LexicalToken
	OtherLexicalTokens map[string]*lexer.LexicalToken
}

func (node *ParseNode) GetType() ParseNodeType {
	return node.NodeType
}

func (node *ParseNode) SetType(typ ParseNodeType) {
	node.NodeType = typ
}

func (node *ParseNode) GetMainToken() *lexer.LexicalToken {
	return node.MainLexicalToken
}

func (node *ParseNode) GetTokenWithKey(key string) *lexer.LexicalToken {
	return node.OtherLexicalTokens[key]
}

func (node *ParseNode) GetParseNodesWithKey(key string) []*ParseNode {
	return node.ParseNodes[key]
}

func (node *ParseNode) Next() *ParseNode {
	return node.next
}

func Parse(tokens []*lexer.LexicalToken) (*ParseNode, error) {
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

func group(tokens []*lexer.LexicalToken) ([]*ParseToken, error) {
	var stack []*ParseToken
	var inside []*lexer.LexicalToken
	var leftParenthesis *lexer.LexicalToken
	level := 0
	for _, e := range tokens {
		if e.GetType() == lexer.LeftParenthesis {
			if level > 0 {
				inside = append(inside, e)
			} else {
				leftParenthesis = e
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
					inside = make([]*lexer.LexicalToken, 0)
					stack = append(stack, &ParseToken{Group: true, Token: leftParenthesis, Tokens: t, LeftParenthesis: leftParenthesis, RightParenthesis: e})
				} else {
					inside = append(inside, e)
				}
			} else {
				return nil, errors.New("unexpected right parenthesis at line " + strconv.Itoa(e.GetLine()) + " position " + strconv.Itoa(e.GetPosition()))
			}
		} else if level == 0 {
			stack = append(stack, &ParseToken{Group: false, Token: e})
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

func divide(tokens []*ParseToken) [][]*ParseToken {
	var statements [][]*ParseToken
	var statement []*ParseToken
	expecting := false
	for _, e := range tokens {
		if e.Group {
			expecting = false
			statement = append(statement, e)
		} else if e.Token.GetType() == lexer.SemiColon || !expecting && e.Token.GetType() == lexer.NewLine {
			expecting = false
			if len(statement) > 0 {
				statements = append(statements, statement)
				statement = make([]*ParseToken, 0)
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

func createParseNodes(statements [][]*ParseToken) (*ParseNode, error) {
	node, _, err := formBlock(statements, 0)
	return node, err
}

func getPrecedence(token *ParseToken) int {
	if token.Group {
		return 8
	}
	switch token.Token.GetType() {
	case lexer.Keyword:
		switch token.Token.GetValue() {
		case lexer.Function:
			fallthrough
		case lexer.If:
			fallthrough
		case lexer.Loop:
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
		case lexer.As:
			fallthrough
		case lexer.To:
			return 9
		}
	case lexer.Operator:
		switch token.Token.GetValue() {
		case lexer.Gets:
			return 2
		case lexer.Or:
			return 4
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
	case lexer.Identifier:
		fallthrough
	case lexer.Integer:
		fallthrough
	case lexer.Text:
		return 8
	}
	panic(fmt.Sprint("unknown token ", util.PrettyString(token)))
}

func formBlock(statements [][]*ParseToken, i int) (*ParseNode, int, error) {
	var root *ParseNode
	var temp *ParseNode
	var pre *ParseNode
	var err error
	for length := len(statements); i < length; i++ {
		if statements[i][0].Token.GetType() == lexer.Keyword && statements[i][0].Token.GetValue() == lexer.End {
			return root, i, nil
		}
		temp, err = formParseNode(statements[i], true)
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
		if temp.NodeType == If || temp.NodeType == ToLoop || temp.NodeType == Function {
			child, a, err := formBlock(statements, i+1)
			i = a
			if err != nil {
				return nil, i, err
			}
			temp.ParseNodes[Children] = append(temp.ParseNodes[Children], child)
		}
	}
	return root, i, nil
}

func formParseNode(tokens []*ParseToken, isStatement bool) (*ParseNode, error) {
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
			return nil, errors.New("unexpected token " + tokens[1].Token.ToString() + " / 1")
		}
		if isStatement {
			return nil, errors.New("was expecting a statement " + t.GetStartPosition())
		}
		return formParseNode(t.Tokens, false)
	}
	typ := t2.GetType()
	switch typ {
	case lexer.Coma:
		if currentIndex == 0 {
			return nil, errors.New("expected token before coma " + t2.ToString())
		} else if currentIndex == len(tokens)-1 {
			return nil, errors.New("expected token after coma " + t2.ToString())
		}
		if isStatement {
			return nil, errors.New("was expecting a statement " + t.GetStartPosition())
		}
		leftChild, err := formParseNode(tokens[:currentIndex], false)
		if err != nil {
			return nil, err
		}
		rightChild, err := formParseNode(tokens[currentIndex+1:], false)
		if err != nil {
			return nil, err
		}
		children := []*ParseNode{leftChild}
		if rightChild.NodeType == Csv {
			children = append(children, rightChild.GetParseNodesWithKey(Children)...)
		} else {
			children = append(children, rightChild)
		}
		return &ParseNode{
			NodeType:         Csv,
			ParseNodes:       map[string][]*ParseNode{Children: children},
			next:             nil,
			MainLexicalToken: t2,
		}, nil
	case lexer.Operator:
		if isStatement {
			return nil, errors.New("was expecting a statement " + t.GetStartPosition())
		}
		var leftChild *ParseNode
		leaveLeftChildNil := false
		if currentIndex == 0 {
			if !(len(tokens) > 1 && t2.GetValue() == lexer.Minus) {
				return nil, errors.New("expected token before operation " + t2.ToString())
			}
			leaveLeftChildNil = true
		}
		if currentIndex == len(tokens)-1 {
			return nil, errors.New("expected token after operation " + t2.ToString())
		}
		var err error
		if !leaveLeftChildNil {
			leftChild, err = formParseNode(tokens[:currentIndex], false)
			if err != nil {
				return nil, err
			}
		}
		rightChild, err := formParseNode(tokens[currentIndex+1:], false)
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
		case lexer.Or:
			nodeType = Or
		case lexer.Greater:
			nodeType = Greater
		case lexer.Less:
			nodeType = Less
		case lexer.Divide:
			nodeType = Divide
		}
		return &ParseNode{NodeType: nodeType, ParseNodes: map[string][]*ParseNode{Children: {leftChild, rightChild}}, next: nil, MainLexicalToken: t2}, nil
	case lexer.Text:
		if len(tokens) != 1 {
			if len(tokens) != 1 {
				return nil, errors.New("unexpected Token " + tokens[1].Token.ToString() + " / 2")
			}
		}
		return &ParseNode{NodeType: String, MainLexicalToken: t2}, nil
	case lexer.Integer:
		if len(tokens) != 1 {
			return nil, errors.New("unexpected token " + tokens[1].Token.ToString() + " / 3 " + util.PrettyString(tokens))
		}
		return &ParseNode{NodeType: Integer, MainLexicalToken: t2}, nil
	case lexer.Identifier:
		if len(tokens) > 1 {
			if !tokens[1].Group && tokens[1].Token.GetType() == lexer.Identifier {
				if len(tokens) > 2 {
					return nil, errors.New("was not expecting more tokens after variable declaration " + tokens[2].Token.ToString())
				}
				return &ParseNode{NodeType: Declaration, MainLexicalToken: t2, OtherLexicalTokens: map[string]*lexer.LexicalToken{Identifier: tokens[1].Token}}, nil
			}
			if tokens[1].Group {
				if len(tokens) > 2 {
					return nil, errors.New("unexpected Token " + tokens[2].Token.ToString() + " / 4")
				}
				parameter, err := formParseNode(tokens[1].Tokens, false)
				if err != nil {
					return nil, err
				}
				parameters := make([]*ParseNode, 0)
				if parameter != nil {
					if parameter.NodeType == Csv {
						parameters = parameter.GetParseNodesWithKey(Children)
					} else {
						parameters = []*ParseNode{parameter}
					}
				}
				return &ParseNode{NodeType: FunctionCall, ParseNodes: map[string][]*ParseNode{Parameters: parameters}, MainLexicalToken: t2}, nil
			}
			return nil, errors.New("unexpected token, was expecting a parenthesis: " + tokens[1].Token.ToString())
		}
		if isStatement {
			return nil, errors.New("was expecting a statement " + t.GetStartPosition())
		}
		return &ParseNode{NodeType: Variable, MainLexicalToken: t2}, nil
	case lexer.Keyword:
		switch t2.GetValue() {
		case lexer.As:
			fallthrough
		case lexer.To:
			return nil, errors.New("unexpected " + t.GetStartPosition())
		case lexer.Loop:
			if !isStatement {
				return nil, errors.New("unexpected loop " + t.GetStartPosition())
			}
			if currentIndex != 0 {
				return nil, errors.New("unexpected loop " + t2.ToString())
			}
			if len(tokens) == 1 {
				return nil, errors.New("expected an expression after loop " + t2.ToString())
			}
			var fromTokens []*ParseToken
			toPosition := -1
			for i, token := range tokens[1:] {
				if token.Token.GetType() == lexer.Keyword && token.Token.GetValue() == lexer.To {
					if len(fromTokens) == 0 {
						return nil, errors.New("was expecting an expression instead of " + token.GetStartPosition())
					}
					toPosition = i + 1
					break
				}
				fromTokens = append(fromTokens, token)
			}
			if toPosition == -1 {
				return nil, errors.New("was expecting a \"to\" after" + tokens[len(tokens)-1].GetEndPosition())
			}
			fromNode, err := formParseNode(fromTokens, false)
			if err != nil {
				return nil, err
			}
			var toTokens []*ParseToken
			asPosition := -1
			for i, token := range tokens[toPosition+1:] {
				if token.Token.GetType() == lexer.Keyword && token.Token.GetValue() == lexer.As {
					if len(toTokens) == 0 {
						return nil, errors.New("was expecting an expression instead of " + token.GetStartPosition())
					}
					asPosition = i + toPosition + 1
					break
				}
				toTokens = append(toTokens, token)
			}
			if asPosition == -1 {
				return nil, errors.New("was expecting an \"as\" after" + tokens[len(tokens)-1].GetEndPosition())
			}
			toNode, err := formParseNode(toTokens, false)
			if err != nil {
				return nil, err
			}
			if len(tokens) == asPosition+1 {
				return nil, errors.New("was expecting an identifier after" + tokens[asPosition].GetEndPosition())
			}
			if tokens[asPosition+1].Token.GetType() != lexer.Identifier {
				return nil, errors.New("was expecting an identifier instead of " + tokens[asPosition+1].GetEndPosition())
			}
			asIdentifier := tokens[asPosition+1].Token
			if len(tokens) != asPosition+2 {
				return nil, errors.New("unexpected " + tokens[asPosition+2].GetStartPosition())
			}
			return &ParseNode{
				NodeType: ToLoop,
				ParseNodes: map[string][]*ParseNode{
					From: {fromNode},
					To:   {toNode},
				},
				MainLexicalToken: t2,
				OtherLexicalTokens: map[string]*lexer.LexicalToken{
					Identifier: asIdentifier,
				},
			}, nil
		case lexer.If:
			if currentIndex != 0 {
				return nil, errors.New("unexpected " + t.GetStartPosition())
			}
			if !isStatement {
				return nil, errors.New("unexpected " + t.GetStartPosition())
			}
			if len(tokens) > 1 {
				condition, err := formParseNode(tokens[currentIndex+1:], false)
				if err != nil {
					return nil, err
				}
				return &ParseNode{NodeType: If, ParseNodes: map[string][]*ParseNode{Children: {condition}}, MainLexicalToken: t2}, nil
			}
			return nil, errors.New("expected a condition after keyword if " + t2.ToString())
		case lexer.Function:
			if len(tokens) == 3 || len(tokens) == 4 {
				base := 1
				var third *lexer.LexicalToken
				if tokens[2].Token.GetType() == lexer.Identifier {
					if tokens[1].Token.GetType() != lexer.Identifier {
						return nil, errors.New("expected identifier " + tokens[1].Token.ToString())
					}
					base = 2
					third = tokens[1].Token
				}
				if tokens[base].Token.GetType() == lexer.Identifier {
					if tokens[base+1].Group {
						parametersNode, err := formParseNode(tokens[base+1].Tokens, false)
						if err != nil {
							return nil, err
						}
						var children []*ParseNode
						switch parametersNode.NodeType {
						case Declaration:
							children = append(children, parametersNode)
						case Csv:
							for _, child := range parametersNode.GetParseNodesWithKey(Children) {
								if child.NodeType != Declaration {
									return nil, errors.New("was expecting parameter declaration " + child.GetMainToken().ToString())
								}
								children = append(children, child)
							}
						default:
							return nil, errors.New("was expecting parameter declaration " + parametersNode.MainLexicalToken.ToString())
						}
						return &ParseNode{
								NodeType:           Function,
								MainLexicalToken:   t2,
								OtherLexicalTokens: map[string]*lexer.LexicalToken{Identifier: tokens[base].Token, ReturnType: third},
								ParseNodes:         map[string][]*ParseNode{Parameters: children}},
							nil
					}
					return nil, errors.New("expected parameters " + tokens[base+1].Token.ToString())
				}
				return nil, errors.New("expected identifier after keyword function " + t2.ToString())
			}
			return nil, errors.New("non complete function declaration")
		case lexer.True:
			if isStatement {
				return nil, errors.New("was expecting a statement" + t.GetStartPosition())
			}
			fallthrough
		case lexer.False:
			if isStatement {
				return nil, errors.New("was expecting a statement" + t.GetStartPosition())
			}
			return &ParseNode{NodeType: Boolean, MainLexicalToken: t2}, nil
		case lexer.Return:
			if !isStatement {
				return nil, errors.New("unexpected return" + t.GetStartPosition())
			}
			temp, err := formParseNode(tokens[1:], false)
			if err != nil {
				return nil, err
			}
			var children []*ParseNode
			if temp != nil {
				children = []*ParseNode{temp}
			}
			return &ParseNode{NodeType: Return, MainLexicalToken: t2, ParseNodes: map[string][]*ParseNode{Children: children}}, nil
		}
	}
	return nil, nil
}
