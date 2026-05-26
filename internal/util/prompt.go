package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompt writes the question, reads one line. If reading fails, or the
// user enters nothing, returns def.
func Prompt(question, def string) string {
	fmt.Printf("%s %s", question, def)
	if !strings.HasSuffix(question, " ") && def == "" {
		fmt.Print(" ")
	}
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return def
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return def
	}
	return line
}
