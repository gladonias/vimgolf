package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	cyan   = color.New(color.FgCyan)
)

func Info(msg string) {
	green.Println(msg)
}

func Warn(msg string) {
	yellow.Println(msg)
}

func Error(msg string) {
	red.Fprintln(os.Stderr, msg)
}

func Debug(msg string) {
	if os.Getenv("GOLFDEBUG") != "" {
		cyan.Println(msg)
	}
}

func Ask(prompt string) (string, error) {
	yellow.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func AskChoice(prompt string, choices []string) (string, error) {
	valid := make(map[string]bool)
	for _, c := range choices {
		if len(c) >= 3 && c[0] == '[' && c[2] == ']' {
			valid[strings.ToLower(string(c[1]))] = true
		}
	}

	for {
		for _, c := range choices {
			Warn(c)
		}
		answer, err := Ask(prompt)
		if err != nil {
			return "", err
		}
		answer = strings.ToLower(strings.TrimSpace(answer))
		if valid[answer] {
			return answer, nil
		}
		fmt.Println("Invalid choice, try again.")
	}
}

func PrintKeystroke(stroke string) {
	if len(stroke) == 1 {
		fmt.Print(stroke)
	} else {
		red.Print(stroke)
	}
}
