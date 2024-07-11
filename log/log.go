package log

import (
    "github.com/fatih/color"
    "hometown-bot/build"
    "hometown-bot/util"
    "log"
)

type Recorder interface {
    Print(v ...any)
    Printf(format string, v ...any)
    Println(v ...any)
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

func (i *InfoRecorder) Print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgBlue, v...))
}

func (i *InfoRecorder) Printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgBlue, format, v...))
}

func (i *InfoRecorder) Println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgBlue, v...))
}

func (i *DebugRecorder) Print(v ...any) {
    if !build.IsDebug {
        return
    }
    i.logger.Print(util.WrapInColor(color.FgGreen, v...))
}

func (i *DebugRecorder) Printf(format string, v ...any) {
    if !build.IsDebug {
        return
    }
    i.logger.Print(util.WrapInColorf(color.FgGreen, format, v...))
}

func (i *DebugRecorder) Println(v ...any) {
    if !build.IsDebug {
        return
    }
    i.logger.Print(util.WrapInColorln(color.FgGreen, v...))
}

func (i *WarningRecorder) Print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgYellow, v...))
}

func (i *WarningRecorder) Printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgYellow, format, v...))
}

func (i *WarningRecorder) Println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgYellow, v...))
}

func (i *ErrorRecorder) Print(v ...any) {
    i.logger.Print(util.WrapInColor(color.FgRed, v...))
}

func (i *ErrorRecorder) Printf(format string, v ...any) {
    i.logger.Print(util.WrapInColorf(color.FgRed, format, v...))
}

func (i *ErrorRecorder) Println(v ...any) {
    i.logger.Print(util.WrapInColorln(color.FgRed, v...))
}
