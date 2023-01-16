package core

var StackTraceType = &Type{Name: "StackTrace", Parent: VariableType, Methods: map[string]Function{}, Converters: map[*Type]Function{}, Scope: NewScope()}

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
		Variable: NewVariable(&StackTrace{ExceptionMessage: exceptionMessage}),
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
	r.Pointer.Variable.VariableInterface.(*StackTrace).AddPosition(position)
	return r
}
