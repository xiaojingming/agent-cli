package agent

import "fmt"

func BuildTaskMessage(taskName string) string {
	return fmt.Sprintf("Agent task %q is ready.", taskName)
}
