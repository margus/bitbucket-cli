package util

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

const (
	reset   = "\033[0m"
	red     = "\033[0;31m"
	green   = "\033[0;32m"
	yellow  = "\033[0;33m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"
	white   = "\033[0;37m"
	gray    = "\033[0;90m"
)

var colors = map[string]string{
	"nocolor": reset,
	"red":     red,
	"green":   green,
	"yellow":  yellow,
	"blue":    blue,
	"magenta": magenta,
	"cyan":    cyan,
	"white":   white,
	"gray":    gray,
}

// disabled reports whether colors are disabled (NO_COLOR or not a TTY).
func disabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return true
	}
	return false
}

func colorize(color, s string) string {
	if disabled() {
		return s
	}
	c, ok := colors[color]
	if !ok {
		c = white
	}
	return c + s + reset
}

// O prints data colored, supporting strings, maps (key/value pairs),
// slices/arrays, and nil. Maps print as "Key: value" with the key
// Title-cased in cyan and the value in yellow.
func O(data any, color string) {
	o(data, color, "", "\n")
}

// ORaw prints without a trailing newline.
func ORaw(s, color string) {
	fmt.Print(colorize(color, s))
}

func o(data any, color, prefix, end string) {
	if data == nil {
		fmt.Print(colorize(color, prefix) + end)
		return
	}

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.String:
		fmt.Print(colorize(color, prefix+v.String()) + end)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Print(colorize(color, fmt.Sprintf("%s%d", prefix, v.Int())) + end)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Print(colorize(color, fmt.Sprintf("%s%d", prefix, v.Uint())) + end)
	case reflect.Float32, reflect.Float64:
		fmt.Print(colorize(color, fmt.Sprintf("%s%v", prefix, v.Float())) + end)
	case reflect.Bool:
		fmt.Print(colorize(color, fmt.Sprintf("%s%v", prefix, v.Bool())) + end)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			o(v.Index(i).Interface(), color, prefix, end)
		}
	case reflect.Map:
		// Stable iteration order to keep output deterministic.
		keys := v.MapKeys()
		sort.SliceStable(keys, func(i, j int) bool {
			return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
		})
		for _, k := range keys {
			ks := fmt.Sprint(k.Interface())
			val := v.MapIndex(k).Interface()
			fmt.Print(colorize("cyan", prefix+ucfirst(ks)+": "))
			printValue(val)
		}
	default:
		fmt.Print(colorize(color, fmt.Sprintf("%s%v", prefix, data)) + end)
	}
}

// printValue prints a single map-value in yellow, handling nested
// structures by recursing through O.
func printValue(v any) {
	if v == nil {
		fmt.Println()
		return
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		fmt.Println(colorize("yellow", rv.String()))
	case reflect.Slice, reflect.Array, reflect.Map:
		fmt.Println()
		O(v, "yellow")
	default:
		fmt.Println(colorize("yellow", fmt.Sprintf("%v", v)))
	}
}

func ucfirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// Errorln prints an error message in red to stderr.
func Errorln(format string, args ...any) {
	fmt.Fprintln(os.Stderr, colorize("red", strings.TrimSpace(fmt.Sprintf(format, args...))))
}
