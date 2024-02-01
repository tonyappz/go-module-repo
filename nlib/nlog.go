package nlib

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type Config struct {
	// Enable console logging
	OutputConsole bool
	// FileLoggingEnabled makes the framework log to a file
	OutputFile bool
	// EncodeLogsAsJson makes the log framework log JSON
	ModeJson bool
	// Directory to log to when file logging is enabled
	LogPath string
	// Filename is the name of the logfile which will be placed inside the directory
	LogFile string
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
}

type NLog struct {
	*zerolog.Logger
}

func Configure(config Config) *NLog {
	var writers []io.Writer
	if config.OutputFile {
		writer := zerolog.ConsoleWriter{
			Out: newRollingFile(config),
		}
		writer.TimeFormat = time.TimeOnly
		writer.FormatLevel = defaultFormatLevel(false)
		writers = append(writers, writer)
	}
	if config.OutputConsole {
		writer := zerolog.ConsoleWriter{
			Out: os.Stderr,
		}
		writer.TimeFormat = time.TimeOnly
		writer.FormatLevel = defaultFormatLevel(false)
		writers = append(writers, writer)
	}
	multiWriter := io.MultiWriter(writers...)
	logger := zerolog.New(multiWriter).With().Timestamp().Logger()
	logger.Info().Bool("fileLogging", config.OutputFile).
		Bool("modeJson", config.ModeJson).
		Str("logPath", config.LogPath).
		Str("logFile", config.LogFile).
		Int("maxSizeMB", config.MaxSize).
		Int("maxBackup", config.MaxBackups).
		Int("maxAgeInDays", config.MaxAge).
		Msg("logging configured")
	return &NLog{
		Logger: &logger,
	}
}

func newRollingFile(config Config) io.Writer {
	if err := os.MkdirAll(config.LogPath, 0744); err != nil {
		log.Error().Err(err).Str("path", config.LogPath).Msg("can't create log directory")
		return nil
	}
	return &lumberjack.Logger{
		Filename:   path.Join(config.LogPath, config.LogFile),
		MaxBackups: config.MaxBackups,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
	}
}

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
	colorBold     = 1
	colorDarkGray = 90
)

func defaultFormatLevel(noColor bool) zerolog.Formatter {
	return func(i interface{}) string {
		var l string
		if ll, ok := i.(string); ok {
			switch ll {
			case zerolog.LevelTraceValue:
				l = colorize("TRC", colorMagenta, noColor)
			case zerolog.LevelDebugValue:
				l = colorize("DBG", colorYellow, noColor)
			case zerolog.LevelInfoValue:
				l = colorize("INF", colorGreen, noColor)
			case zerolog.LevelWarnValue:
				l = colorize("WRN", colorRed, noColor)
			case zerolog.LevelErrorValue:
				l = colorize(colorize("ERR", colorRed, noColor), colorBold, noColor)
			case zerolog.LevelFatalValue:
				l = colorize(colorize("FTL", colorRed, noColor), colorBold, noColor)
			case zerolog.LevelPanicValue:
				l = colorize(colorize("PNC", colorRed, noColor), colorBold, noColor)
			default:
				l = colorize(ll, colorBold, noColor)
			}
		} else {
			if i == nil {
				l = colorize("???", colorBold, noColor)
			} else {
				l = strings.ToUpper(fmt.Sprintf("%s", i))[0:3]
			}
		}
		return l
	}
}

func colorize(s interface{}, c int, disabled bool) string {
	e := os.Getenv("NO_COLOR")
	if e != "" {
		disabled = true
	}

	if disabled {
		return fmt.Sprintf("|%s|", s)
	}
	return fmt.Sprintf("\x1b[%dm|%v|\x1b[0m", c, s)
}
