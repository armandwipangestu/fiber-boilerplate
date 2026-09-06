package console

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

var errAborted = errors.New("aborted")

// confirm prompts the user on stdin for a yes/no answer. It returns false for
// any non-yes input, including EOF. Empty confirmation on non-interactive
// terminals is treated as "no".
func confirm(prompt string) bool {
	fmt.Fprint(os.Stderr, prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}
