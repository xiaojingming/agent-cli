package main

import (
	"fmt"
	"os"
)

const agentName = "review-agent"

func main() {
	modelName := "gpt-5"
	maxTokens := 4096
	timeoutSeconds := 30
	maxRetries := 3
	var debug = os.Getenv("DEBUG") == "true"

	if maxTokens < 0 {
		fmt.Printf("maxTokens 应该大于0")
	}

	if maxRetries < 0 {
		fmt.Printf("maxRetries 应该大于0")
	}

	if timeoutSeconds < 0 {
		fmt.Printf("timeoutSeconds 应该大于0")
	}

	fmt.Println("agentName:", agentName)
	fmt.Println("modelName:", modelName)
	fmt.Println("maxTokens:", maxTokens)
	fmt.Println("timeoutSeconds:", timeoutSeconds)
	fmt.Println("maxRetries:", maxRetries)
	fmt.Println("debug:", debug)

}
