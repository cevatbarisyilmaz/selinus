package core

var VariableType = &Type{Name: "VariableInterface", Parent: nil, Methods: map[string]Function{}, Converters: map[*Type]Function{}, Scope: NewScope()}

type VariableInterface interface {
	GetType() *Type
}

const Self = "Self"

func NewVariable(variableInterface VariableInterface) *Variable {
	return &Variable{VariableInterface: variableInterface}
}

type Variable struct {
	VariableInterface
}

func (variable *Variable) ToPointer() *Pointer {
	return &Pointer{
		Typ:      variable.GetType(),
		Variable: variable,
	}
}

func (variable *Variable) ConvertTo(typ *Type) *Return {
	if variable.GetType().Is(typ) {
		return &Return{
			ReturnType: NOTHING,
			Pointer:    variable.ToPointer(),
		}
	}
	converter := variable.GetType().Converters[typ]
	if converter != nil {
		var res *Return
		variable.GetType().Scope.CloneWithNewBlock(func(scope *Scope) {
			scope.DeclareAndSet(Self, variable.ToPointer())
			res = converter.Execute(scope)
		})
		return res
	}
	return NewExceptionReturn("conversion from " + variable.GetType().Name + " to " + typ.Name + " is not possible")
}

func (variable *Variable) CallMethod(method string) (res *Return) {
	variable.GetType().Scope.CloneWithNewBlock(func(scope *Scope) {
		scope.DeclareAndSet(Self, variable.ToPointer())
		res = variable.GetType().Methods[method].Execute(scope)
	})
	return
}
