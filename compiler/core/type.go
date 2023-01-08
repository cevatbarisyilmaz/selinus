package core

type Type struct {
	Name      string
	Parent    *Type
	Functions []*Function
	Generic   bool
	Generics  []*Type
}

func (typ *Type) IsCompatible(other *Type) bool {
	if typ == other {
		return true
	}
	if typ.Generic && other.Generic && len(typ.Generics) == len(other.Generics) && typ.Parent.IsCompatible(other.Parent) {
		for i, e := range typ.Generics {
			o := other.Generics[i]
			if o == nil {
				continue
			}
			if e == nil || !e.IsCompatible(o) {
				return false
			}
		}
		return true
	}
	t := typ.Parent
	for t != nil {
		if t == other.Parent {
			return true
		}
		t = t.Parent
	}
	return false
}

var TypeType = &Type{Name: "Type", Parent: nil, Functions: nil}

type TypeVariable struct {
	Value *Type
}

func (*TypeVariable) GetType() *Type {
	return TypeType
}

func (t *TypeVariable) GetStringValue() string {
	return t.Value.Name
}

func TypeToVariable(typ *Type) Variable {
	return &TypeVariable{Value: typ}
}
