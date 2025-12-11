package logger

type NopLogger struct{}

func (NopLogger) Debug(...interface{})          {}
func (NopLogger) Info(...interface{})           {}
func (NopLogger) Warn(...interface{})           {}
func (NopLogger) Error(...interface{})          {}
func (NopLogger) Fatal(...interface{})          {}
func (NopLogger) Debugf(string, ...interface{}) {}
func (NopLogger) Infof(string, ...interface{})  {}
func (NopLogger) Warnf(string, ...interface{})  {}
func (NopLogger) Errorf(string, ...interface{}) {}
func (NopLogger) Fatalf(string, ...interface{}) {}
func (NopLogger) Sync() error                   { return nil }

func NewNop() Logger {
	return NopLogger{}
}
