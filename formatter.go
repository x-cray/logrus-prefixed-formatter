package prefixed

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mgutz/ansi"
)

const reset = ansi.Reset

var (
	baseTimestamp time.Time
)

func init() {
	baseTimestamp = time.Now()
}

func miniTS() int {
	return int(time.Since(baseTimestamp) / time.Second)
}

type TextFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Use formatted layout, but without colors for a non-TTY output
	ForceFormatting bool

	// Force disabling colors.
	DisableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging of just the time passed since beginning of execution.
	ShortTimestamp bool

	// Timestamp format to use for display when a full timestamp is printed.
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// Pad msg field with spaces on the right for display.
	// The value for this parameter will be the size of padding.
	// Its default value is zero, which means no padding will be applied for msg.
	SpacePadding int

	// Whether the logger's out is to a terminal
	isTerminal   bool
	terminalOnce sync.Once
}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var keys []string = make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		if k != "prefix" {
			keys = append(keys, k)
		}
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}

	b := &bytes.Buffer{}

	prefixFieldClashes(entry.Data)

	f.terminalOnce.Do(func() {
		if entry.Logger != nil {
			f.isTerminal = logrus.IsTerminal(entry.Logger.Out)
		}
	})

	isColored := (f.ForceColors || f.isTerminal) && !f.DisableColors
	isFormatted := f.ForceFormatting || f.isTerminal

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.Stamp
	}
	if isColored || isFormatted {
		f.printFormatted(b, entry, keys, timestampFormat, isColored)
	} else {
		if !f.DisableTimestamp {
			f.appendKeyValue(b, "time", entry.Time.Format(timestampFormat))
		}
		f.appendKeyValue(b, "level", entry.Level.String())
		if entry.Message != "" {
			f.appendKeyValue(b, "msg", entry.Message)
		}
		for _, key := range keys {
			f.appendKeyValue(b, key, entry.Data[key])
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) printFormatted(b *bytes.Buffer, entry *logrus.Entry, keys []string, timestampFormat string, isColored bool) {
	var levelColor string
	var levelText string
	if isColored {
		switch entry.Level {
		case logrus.InfoLevel:
			levelColor = ansi.Green
		case logrus.WarnLevel:
			levelColor = ansi.Yellow
		case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
			levelColor = ansi.Red
		default:
			levelColor = ansi.Blue
		}
	}

	if entry.Level != logrus.WarnLevel {
		levelText = strings.ToUpper(entry.Level.String())
	} else {
		levelText = "WARN"
	}

	prefix := ""
	message := entry.Message

	resetColor, prefixColor, tsColor := f.setupColors(isColored)

	if prefixValue, ok := entry.Data["prefix"]; ok {
		prefix = fmt.Sprint(" ", prefixColor, prefixValue, ":", resetColor)
	} else {
		prefixValue, trimmedMsg := extractPrefix(entry.Message)
		if len(prefixValue) > 0 {
			prefix = fmt.Sprint(" ", prefixColor, prefixValue, ":", resetColor)
			message = trimmedMsg
		}
	}

	messageFormat := "%s"
	if f.SpacePadding != 0 {
		messageFormat = fmt.Sprintf("%%-%ds", f.SpacePadding)
	}

	if f.DisableTimestamp {
		fmt.Fprintf(b, "%s%s %s%+5s%s%s "+messageFormat, tsColor, resetColor, levelColor, levelText, resetColor, prefix, message)
	} else {
		if f.ShortTimestamp {
			fmt.Fprintf(b, "%s[%04d]%s %s%+5s%s%s "+messageFormat, tsColor, miniTS(), resetColor, levelColor, levelText, resetColor, prefix, message)
		} else {
			fmt.Fprintf(b, "%s[%s]%s %s%+5s%s%s "+messageFormat, tsColor, entry.Time.Format(timestampFormat), resetColor, levelColor, levelText, resetColor, prefix, message)
		}
	}
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s%s%s=%+v", levelColor, k, resetColor, v)
	}
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func extractPrefix(msg string) (string, string) {
	prefix := ""
	regex := regexp.MustCompile("^\\[(.*?)\\]")
	if regex.MatchString(msg) {
		match := regex.FindString(msg)
		prefix, msg = match[1:len(match)-1], strings.TrimSpace(msg[len(match):])
	}
	return prefix, msg
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	case error:
		errmsg := value.Error()
		if needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

func prefixFieldClashes(data logrus.Fields) {
	_, ok := data["time"]
	if ok {
		data["fields.time"] = data["time"]
	}
	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}
	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}
}

func (f *TextFormatter) setupColors(isColored bool) (resetColor string, prefixColor string, tsColor string) {
	if isColored {
		resetColor = reset
		prefixColor = ansi.Cyan
		tsColor = ansi.LightBlack
		return
	} else {
		// leave as empty strings to "disable" coloring
		return
	}
}
