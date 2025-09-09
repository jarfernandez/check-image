package main

import (
	"check-image/cmd/check-image/commands"
	"fmt"
	"os"
)

func main() {
	commands.Execute()

	if commands.Result == commands.ValidationFailed {
		fmt.Println("Validation failed")
		os.Exit(1)
	}

	if commands.Result == commands.ValidationSucceeded {
		fmt.Println("Validation succeeded")
	}
}
