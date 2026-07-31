package main

import (
	"fmt"
	"os"

	"github.com/xiaojingming/agent-cli/agent"
)

func main() {
	message := agent.BuildTaskMessage(os.Args[1])
	fmt.Println(message)

}
