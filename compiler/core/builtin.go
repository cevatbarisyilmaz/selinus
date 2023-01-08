package core

//Built-in Types

//Base Type

var VariableType = &Type{Name: "Variable", Parent: nil, Functions: nil}

//Variable Sub-Types

var StringType = &Type{Name: "String", Parent: VariableType, Functions: nil}

var StackTraceType = &Type{Name: "StackTrace", Parent: StringType, Functions: nil}

type String struct {
	Value string
}

func (*String) GetType() *Type {
	return StringType
}

func (s *String) GetStringValue() string {
	return s.Value
}

func NewStringPointer(value string) *Pointer {
	return &Pointer{
		Typ:      StringType,
		Variable: &String{Value: value},
	}
}

type StackTrace struct {
	ExceptionMessage string
	Positions        []string
}

func (s *StackTrace) AddPosition(position string) {
	s.Positions = append(s.Positions, position)
}

func (*StackTrace) GetType() *Type {
	return StackTraceType
}

func (s *StackTrace) GetStringValue() string {
	msg := s.ExceptionMessage
	for _, position := range s.Positions {
		msg += "\n"
		msg += position
	}
	return msg
}

func NewStackTracePointer(exceptionMessage string) *Pointer {
	return &Pointer{
		Typ:      StackTraceType,
		Variable: &StackTrace{ExceptionMessage: exceptionMessage},
	}
}

func NewExceptionReturn(message string) *Return {
	return &Return{
		ReturnType: EXCEPTION,
		Pointer:    NewStackTracePointer(message),
	}
}

func AddPositionToStackTrace(r *Return, position string) *Return {
	if r.ReturnType != EXCEPTION {
		return r
	}
	r.Pointer.Variable.(*StackTrace).AddPosition(position)
	return r
}
