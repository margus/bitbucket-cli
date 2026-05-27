package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptFn is the function actually called by Prompt. Tests override it
// to stub interactive input without standing up a fake stdin:
//
//	saved := util.PromptFn
//	util.PromptFn = func(q, def string) string { return "alice" }
//	t.Cleanup(func() { util.PromptFn = saved })
var PromptFn = readLineFromStdin

// Prompt writes the question and returns one line of input. If reading
// fails or the user enters nothing, returns def.
func Prompt(question, def string) string {
	return PromptFn(question, def)
}

func readLineFromStdin(question, def string) string {
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
