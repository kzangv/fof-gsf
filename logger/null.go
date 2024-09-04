package logger

// ToNull

type ToNull int

func (output ToNull) Level() int          { return int(output) }
func (output *ToNull) SetLevel(v int) int { *output = ToNull(v); return int(*output) }

func (output ToNull) Debug(string, ...interface{}) {}
func (output ToNull) Info(string, ...interface{})  {}
func (output ToNull) Warn(string, ...interface{})  {}
func (output ToNull) Error(string, ...interface{}) {}

func (output ToNull) DebugForce(string, ...interface{}) {}
func (output ToNull) InfoForce(string, ...interface{})  {}
func (output ToNull) WarnForce(string, ...interface{})  {}
func (output ToNull) ErrorForce(string, ...interface{}) {}
