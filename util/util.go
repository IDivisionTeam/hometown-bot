package util

import (
    "fmt"
    "github.com/fatih/color"
    "sync"
)

var (
    colorsCache   = make(map[color.Attribute]*color.Color)
    colorsCacheMu sync.Mutex
)

func WrapInColor(c color.Attribute, v ...any) string {
    if !inBetween(c, color.FgBlack, color.FgWhite) {
        return fmt.Sprint(v...)
    }

    return getCachedColor(c).Sprint(v...)
}

func WrapInColorf(c color.Attribute, format string, v ...any) string {
    if !inBetween(c, color.FgBlack, color.FgWhite) {
        return fmt.Sprintf(format, v...)
    }

    return getCachedColor(c).Sprintf(format, v...)
}

func WrapInColorln(c color.Attribute, v ...any) string {
    if !inBetween(c, color.FgBlack, color.FgWhite) {
        return fmt.Sprintln(v...)
    }

    return getCachedColor(c).Sprintln(v...)
}

func getCachedColor(p color.Attribute) *color.Color {
    colorsCacheMu.Lock()
    defer colorsCacheMu.Unlock()

    c, ok := colorsCache[p]
    if !ok {
        c = color.New(p)
        colorsCache[p] = c
    }

    return c
}

func inBetween(i, min, max color.Attribute) bool {
    return (i >= min) && (i <= max)
}
