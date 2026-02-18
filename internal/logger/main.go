package logger

import "fmt"

func Info(service, message string) {
	fmt.Printf("[INFO] [%s] %s\n", service, message)
}
