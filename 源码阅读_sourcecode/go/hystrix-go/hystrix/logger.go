package myhystrix

type logger interface {
	Printf(format string, items ...interface{})
}

type NoopLogger struct{}

func (l NoopLogger) Printf(format string, items ...interface{}) {}
