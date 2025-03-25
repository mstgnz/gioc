package main

import (
	"fmt"

	"github.com/mstgnz/gioc"
)

// Logger interface
type Logger interface {
	Log(message string)
}

// ConsoleLogger implements Logger
type ConsoleLogger struct{}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) Log(message string) {
	fmt.Println("Log:", message)
}

// Service that depends on Logger interface
type Service struct {
	logger Logger
}

func NewService(logger Logger) *Service {
	return &Service{logger: logger}
}

func main() {
	// Register concrete implementation
	logger := gioc.IOC(NewConsoleLogger)

	// Use the logger in a service
	service := gioc.IOC(func() *Service { return NewService(logger) })

	// Use the service
	service.logger.Log("Service initialized successfully")
}
