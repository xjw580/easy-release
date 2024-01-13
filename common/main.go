package common

const (
	ProgramName = "easy-release"
)

type Logs interface {
	AppendLog(log string)
}
