package pan

import (
	"fmt"
	"os"
)

// PrintSuccess prints a success message with consistent formatting
func PrintSuccess(message string) {
	fmt.Printf("[✓] %s\n", message)
}

// PrintError prints an error message with consistent formatting
func PrintError(message string) {
	fmt.Printf("[×] %s\n", message)
}

// PrintErrorAndExit prints an error message and exits with code 1
func PrintErrorAndExit(message string) {
	PrintError(message)
	os.Exit(1)
}