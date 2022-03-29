package packagelog

//Logger is the module's logging interface. This obe is compatible with the standard os logger,
//but won't be great for a lot of popular logging libraries.
type Logger interface {
	Printf(l string, args ...interface{})
	Fatalf(l string, args ...interface{})
	//Infof and Errorf give a lot of compatibility with existing logging libraries
	//Infof(l string, args ...interface{})
	//Errorf(l string, args ...interface{})
}

//NoopLogger is the default provided logger
type NoopLogger struct{}

func (NoopLogger) Printf(l string, args ...interface{}) {}
func (NoopLogger) Fatalf(l string, args ...interface{}) {}

var moduleLogger Logger = NoopLogger{}

//SetLogger allows the package user to provide his own implementation
func SetLogger(l Logger) {
	moduleLogger = l
}

func MyCoolFunction(name string, age int) {
	//The module just logs through the moduleLogger
	moduleLogger.Printf("Name: %s, Age: %d", name, age)
}
