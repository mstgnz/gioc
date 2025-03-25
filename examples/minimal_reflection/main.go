package main

import (
	"fmt"

	"github.com/mstgnz/gioc"
)

// Database represents a database connection
type Database struct {
	connection string
}

// NewDatabase creates a new database connection
func NewDatabase() *Database {
	fmt.Println("Creating new Database instance")
	return &Database{connection: "localhost:5432"}
}

// Logger represents a logger service
type Logger struct {
	level string
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	fmt.Println("Creating new Logger instance")
	return &Logger{level: "info"}
}

// UserService represents a user service with dependencies
type UserService struct {
	db     *Database
	logger *Logger
}

// NewUserService creates a new user service with injected dependencies
func NewUserService(db *Database, logger *Logger) *UserService {
	fmt.Println("Creating new UserService instance")
	return &UserService{
		db:     db,
		logger: logger,
	}
}

// OrderService represents a service with three dependencies
type OrderService struct {
	db          *Database
	logger      *Logger
	userService *UserService
}

// NewOrderService creates a new order service
func NewOrderService(db *Database, logger *Logger, userService *UserService) *OrderService {
	fmt.Println("Creating new OrderService instance")
	return &OrderService{
		db:          db,
		logger:      logger,
		userService: userService,
	}
}

func main() {
	fmt.Println("=== Minimal Reflection Example ===")

	// Example 1: Using DirectIOC - minimal reflection approach
	fmt.Println("\nExample 1: Using DirectIOC")
	db1 := gioc.DirectIOC(NewDatabase)
	logger1 := gioc.DirectIOC(NewLogger)

	// Manually wire up the dependencies
	userService1 := NewUserService(db1, logger1)
	fmt.Printf("UserService1: %+v\n", userService1)
	fmt.Printf("Database connection: %s\n", userService1.db.connection)
	fmt.Printf("Logger level: %s\n", userService1.logger.level)

	// Example 2: Using TypedInjectConstructor
	fmt.Println("\nExample 2: Using TypedInjectConstructor")
	userService2 := gioc.TypedInjectConstructor2(NewUserService, NewDatabase, NewLogger)
	fmt.Printf("UserService2: %+v\n", userService2)
	fmt.Printf("Database connection: %s\n", userService2.db.connection)
	fmt.Printf("Logger level: %s\n", userService2.logger.level)

	// Example 3: Using CreateFactory
	fmt.Println("\nExample 3: Using CreateFactory")
	// Create a factory function for UserService
	userServiceFactory := gioc.CreateFactory2(NewUserService, NewDatabase, NewLogger)

	// Register the factory with IOC
	userService3 := gioc.IOC(userServiceFactory)
	fmt.Printf("UserService3: %+v\n", userService3)
	fmt.Printf("Database connection: %s\n", userService3.db.connection)
	fmt.Printf("Logger level: %s\n", userService3.logger.level)

	// Example 4: Three-level dependency with TypedInjectConstructor3
	fmt.Println("\nExample 4: Three-level dependency with TypedInjectConstructor3")
	// Creating with explicit factory
	userServiceFactory2 := gioc.CreateFactory2(NewUserService, NewDatabase, NewLogger)
	orderService := gioc.TypedInjectConstructor3(NewOrderService, NewDatabase, NewLogger, userServiceFactory2)

	fmt.Printf("OrderService: %+v\n", orderService)
	fmt.Printf("OrderService.UserService: %+v\n", orderService.userService)
}
