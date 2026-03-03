package utils

import (
	"fmt"
	"log"
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

func Abs[T int | float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

type FIFO[T any] struct {
	data []*T
}

func (f *FIFO[T]) Enqueue(x *T) {
	f.data = append(f.data, x)
}

func (f *FIFO[T]) Dequeue() *T {
	if len(f.data) == 0 {
		return nil
	}
	x := f.data[0]
	f.data = f.data[1:]
	return x
}

func (f *FIFO[T]) IsEmpty() bool {
	return len(f.data) == 0
}

func InitLogFile() {
	f, err := os.OpenFile("hnreader.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(f)
	log.Println("Starting TUI")
}
