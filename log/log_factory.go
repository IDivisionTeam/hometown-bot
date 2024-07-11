package log

import (
    "hometown-bot/build"
    "log"
    "os"
    "sync"
)

type Type int

const (
    INFO Type = iota
    DEBUG
    WARN
    ERROR
)

var (
    loggers = make(map[Type]Recorder)
    once    sync.Once
)

func createRecorders() {
    loggers[INFO] = newInfoRecorder()
    loggers[DEBUG] = newDebugRecorder()
    loggers[WARN] = newWarningRecorder()
    loggers[ERROR] = newErrorRecorder()
}

func newInfoRecorder() Recorder {
    return &InfoRecorder{
        logger: log.New(os.Stdout, "INFO: ", log.Lmsgprefix|log.LstdFlags),
    }
}

func newDebugRecorder() Recorder {
    if build.IsDebug {
        return &DebugRecorder{
            logger: log.New(os.Stdout, "DEBUG: ", log.Lmsgprefix|log.LstdFlags),
        }
    }

    return &DebugRecorder{}
}

func newWarningRecorder() Recorder {
    return &WarningRecorder{
        logger: log.New(os.Stdout, "WARNING: ", log.Lmsgprefix|log.LstdFlags),
    }
}

func newErrorRecorder() Recorder {
    return &ErrorRecorder{
        logger: log.New(os.Stderr, "ERROR: ", log.Lmsgprefix|log.LstdFlags),
    }
}

func Print(lt Type, v ...any) {
    once.Do(createRecorders)
    loggers[lt].Print(v...)
}

func Printf(lt Type, format string, v ...any) {
    once.Do(createRecorders)
    loggers[lt].Printf(format, v...)
}

func Println(lt Type, v ...any) {
    once.Do(createRecorders)
    loggers[lt].Println(v...)
}

func Info() Recorder {
    once.Do(createRecorders)
    return loggers[INFO]
}

func Debug() Recorder {
    once.Do(createRecorders)
    return loggers[DEBUG]
}

func Warn() Recorder {
    once.Do(createRecorders)
    return loggers[WARN]
}

func Error() Recorder {
    once.Do(createRecorders)
    return loggers[ERROR]
}
