package main

import (
	"os"
)

func DoesPathExist(path string) bool {
	result := false
	if _, err := os.Stat(path); err == nil {
		result = true
	}
	return result
}

func main() {
	Start(os.Args)
}
