package log

import (
	// "fmt"

	"os"
	"time"

	// "gopkg.in/natefinch/lumberjack.v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var log zerolog.Logger

func init() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

	logLevel := -1

	// TODO: Come back to the file + stdout logger - for now we will just
	// log to the console, but we will want a file logger for full debug in
	// the future

	// fileLogger := &lumberjack.Logger{
	// 	Filename:   fmt.Sprintf("%s.log", os.Getenv("HOSTNAME")),
	// 	MaxSize:    5, // megabytes
	// 	MaxBackups: 10,
	// 	MaxAge:     14, // days
	// 	Compress:   true,
	// }
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	log = zerolog.New(output).
		Level(zerolog.Level(logLevel)).
		With().
		Timestamp().
		Logger()
}

// Return the logger
func Get() zerolog.Logger {
	return log
}

// Create a new logger with the desired level
// TODO: Can we use textual log levels instead of int?
func SetLevel(level int) {
	log = log.Level(zerolog.Level(level))
}
