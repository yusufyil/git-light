package myersdiff

type Diff struct {
	PreviousBlobHash string
	Commands         string
	Data             []string
}
