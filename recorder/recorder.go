package recorder

import (
    "github.com/fatih/color"
    "hometown-bot/build"
    "hometown-bot/util"
    "log"
)

type Recorder interface {
    print(v ...any)
    printf(format string, v ...any)
    println(v ...any)
}

type InfoRecorder struct {
    logger *log.Logger
}

type DebugRecorder struct {
    logger *log.Logger
}

type WarningRecorder struct {
    logger *log.Logger
}

type ErrorRecorder struct {
    logger *log.Logger
}

func (i *InfoRecorder) print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgBlue, v...))
}

func (i *InfoRecorder) printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgBlue, format, v...))
}

func (i *InfoRecorder) println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgBlue, v...))
}

func (i *DebugRecorder) print(v ...any) {
    if !build.IsDebug { return }
    i.logger.Print(util.WrapInColor(color.FgGreen, v...))
}

func (i *DebugRecorder) printf(format string, v ...any) {
    if !build.IsDebug { return }
    i.logger.Print(util.WrapInColorf(color.FgGreen, format, v...))
}

func (i *DebugRecorder) println(v ...any) {
    if !build.IsDebug { return }
    i.logger.Print(util.WrapInColorln(color.FgGreen, v...))
}

func (i *WarningRecorder) print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgYellow, v...))
}

func (i *WarningRecorder) printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgYellow, format, v...))
}

func (i *WarningRecorder) println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgYellow, v...))
}

func (i *ErrorRecorder) print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgRed, v...))
}

func (i *ErrorRecorder) printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgRed, format, v...))
}

func (i *ErrorRecorder) println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgRed, v...))
}
