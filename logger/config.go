package logger

type LoggingLevel int

const (
	ERRORLVL = 1
	WARNLVL  = 2
	INFOLVL  = 4 // You might ask, where the hell is the 3. Well, 4 has 1 only in the thrid bit (100). So each level has only a single 1 in its binary representation.
	// Now how is this useful? I don't know to be honest, just thought it is a bit cooler.
)

type Config struct {
	Level LoggingLevel
	// Obviously, there are some other configs, to be stated later
}
