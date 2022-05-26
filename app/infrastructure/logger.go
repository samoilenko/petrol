package infrastructure

import "fmt"

type Logger struct {
}

func (l *Logger) Error(args ...interface{}) {
	fmt.Println(args...)
}

func (l *Logger) Info(args ...interface{}) {
	fmt.Println(args...)
}

func NewLogger() *Logger {
	return &Logger{}
}
