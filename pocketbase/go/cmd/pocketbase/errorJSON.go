package main

import "fmt"

func errorJSON(format string, a ...any) any {
	return map[string]string{
		"error": fmt.Sprintf(format, a...),
	}
}
