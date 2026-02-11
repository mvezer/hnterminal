package utils

import (
	"fmt"
	"os"
)

type ErrorSeverity int

const (
	ErrorSeverityWarn = iota
	ErrorSeverityError
	ErrorSeverityFatal
)

func HandleError(err error, severity ErrorSeverity) {
	switch severity {
	case ErrorSeverityWarn:
		fmt.Printf("WARN: %s", err.Error())
	case ErrorSeverityError:
		fmt.Printf("ERROR: %s", err.Error())
	case ErrorSeverityFatal:
		fmt.Printf("FATAL: %s", err.Error())
		os.Exit(1)
	}
}

func Max[T int | float64](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func Min[T int | float64](x, y T) T {
	if x < y {
		return x
	}
	return y
}
