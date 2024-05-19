package myersdiff

type operation uint

const (
	INSERT operation = iota
	DELETE
	MOVE
)
