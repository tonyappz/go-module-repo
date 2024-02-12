package nlib

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type NLogConfig struct {
	OutputConsole bool   // Enable console logging
	OutputFile    bool   // FileLoggingEnabled makes the framework log to a file
	LogPath       string // Directory to log to when file logging is enabled
	LogFile       string // Filename is the name of the logfile which will be placed inside the directory
	MaxSize       int    // MaxSize the max size in MB of the logfile before it's rolled
	MaxBackups    int    // MaxBackups the max number of rolled files to keep
	MaxAge        int    // MaxAge the max age in days to keep a logfile
}

type NLog struct {
	*zerolog.Logger
}

func NewLog(config NLogConfig) *NLog {
	var writers []io.Writer
	if config.OutputFile {
		writer := zerolog.ConsoleWriter{
			Out: newRollingFile(config),
		}
		writer.FormatTimestamp = defaultTimestamp()
		writer.FormatCaller = defaultCaller(false)
		writer.FormatLevel = defaultFormatLevel(false)
		writers = append(writers, writer)
	}
	if config.OutputConsole {
		writer := zerolog.ConsoleWriter{
			Out: os.Stderr,
		}
		writer.FormatTimestamp = defaultTimestamp()
		writer.FormatCaller = defaultCaller(false)
		writer.FormatLevel = defaultFormatLevel(false)
		writers = append(writers, writer)
	}
	multiWriter := io.MultiWriter(writers...)
	logger := zerolog.New(multiWriter).With().Timestamp().Caller().Logger()
	logger.Info().Bool("fileLogging", config.OutputFile).
		Bool("consoleLogging", config.OutputConsole).
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

func newRollingFile(config NLogConfig) io.Writer {
	if err := os.MkdirAll(config.LogPath, 0744); err != nil {
		log.Error().Err(err).Str("path", config.LogPath).Msg("can't create log directory")
		return nil
	}
	writer := &lumberjack.Logger{
		Filename:   path.Join(config.LogPath, config.LogFile),
		MaxBackups: config.MaxBackups,
		MaxSize:    config.MaxSize,
		MaxAge:     config.MaxAge,
	}
	return writer
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

func defaultCaller(noColor bool) zerolog.Formatter {
	return func(i interface{}) string {
		var caller string
		if cc, ok := i.(string); ok {
			caller = cc
		}
		if len(caller) > 0 {
			if cwd, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(cwd, caller); err == nil {
					caller = rel
				}
			}
		}
		return fmt.Sprintf("%26s >", caller)
	}
}

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
				l = colorizeBoth("ERR", colorRed, colorBold, noColor)
			case zerolog.LevelFatalValue:
				l = colorizeBoth("FTL", colorRed, colorBold, noColor)
			case zerolog.LevelPanicValue:
				l = colorizeBoth("PNC", colorRed, colorBold, noColor)
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

func defaultTimestamp() zerolog.Formatter {
	return func(i interface{}) string {
		t := time.Now().Local()
		return fmt.Sprintf("%02d:%02d:%02d.%03d", t.Hour(), t.Minute(), t.Second(), t.UnixMilli()%1000)
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

func colorizeBoth(s interface{}, c1 int, c2 int, disabled bool) string {
	e := os.Getenv("NO_COLOR")
	if e != "" {
		disabled = true
	}
	if disabled {
		return fmt.Sprintf("|%s|", s)
	}
	return fmt.Sprintf("\x1b[%dm\x1b[%dm|%v|\x1b[0m", c1, c2, s)
}
