package myersdiff

type operation uint

const (
	INSERT operation = iota
	DELETE
	MOVE
)

var colors = map[operation]string{
	INSERT: "\033[32m",
	DELETE: "\033[31m",
	MOVE:   "\033[39m",
}
