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

// Açık tip belirtimi ile injection sağlayan yardımcı fonksiyon
func CreateUserService() *UserService {
	db := gioc.IOC(NewDatabase)
	logger := gioc.IOC(NewLogger)
	return NewUserService(db, logger)
}

func main() {
	fmt.Println("=== Constructor Injection Example ===")

	// Approach 1: Explicit dependency injection with WithDependency
	fmt.Println("Approach 1: Explicit dependencies")
	userService1 := gioc.InjectConstructor[*UserService](NewUserService,
		gioc.WithDependency("db", NewDatabase),
		gioc.WithDependency("logger", NewLogger),
	)
	fmt.Printf("UserService1: %+v\n", userService1)
	fmt.Printf("Database connection: %s\n", userService1.db.connection)
	fmt.Printf("Logger level: %s\n\n", userService1.logger.level)

	// List instances
	fmt.Println("Registered instances after approach 1:")
	gioc.ListInstances()
	fmt.Println()

	// Clear instances for demonstration
	gioc.ClearInstances()

	// Approach 2: Using pre-registered instances
	fmt.Println("Approach 2: Using pre-registered instances")
	// Register the dependencies first
	_ = gioc.IOC(NewDatabase)
	_ = gioc.IOC(NewLogger)

	// Use them in the constructor injection
	userService2 := gioc.InjectConstructor[*UserService](NewUserService,
		gioc.WithDependency("db", NewDatabase),
		gioc.WithDependency("logger", NewLogger),
	)
	fmt.Printf("UserService2: %+v\n", userService2)
	fmt.Printf("Database connection: %s\n", userService2.db.connection)
	fmt.Printf("Logger level: %s\n", userService2.logger.level)

	// Approach 3: Reflection kullanımını azaltan yaklaşım
	fmt.Println("\nApproach 3: Minimum reflection with explicit type declaration")
	userService3 := gioc.IOC(CreateUserService)
	fmt.Printf("UserService3: %+v\n", userService3)
	fmt.Printf("Database connection: %s\n", userService3.db.connection)
	fmt.Printf("Logger level: %s\n", userService3.logger.level)
}
