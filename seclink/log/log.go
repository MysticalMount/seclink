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

func InitLog(logLevel int) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339Nano

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
