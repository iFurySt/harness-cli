package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func promptLanguage(in io.Reader, out io.Writer) (Language, error) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintln(out, "Select template language:")
		fmt.Fprintln(out, "  1. English")
		fmt.Fprintln(out, "  2. Chinese")
		fmt.Fprint(out, "Choice [1]: ")

		if !scanner.Scan() {
			return "", fmt.Errorf("language is required; pass --language en or --language zh")
		}

		language, err := ParseLanguage(strings.TrimSpace(scanner.Text()))
		if err == nil {
			return language, nil
		}

		fmt.Fprintln(out, err)
	}
}
