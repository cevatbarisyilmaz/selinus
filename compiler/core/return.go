package core

type ReturnType int

const (
	NOTHING ReturnType = iota
	BREAK
	CONTINUE
	RETURN
	EXCEPTION
)

type Return struct {
	ReturnType ReturnType
	Pointer    *Pointer
}
