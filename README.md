# **gioc** - Go Inversion of Control (IoC) Container

`gioc` is a simple and lightweight **Inversion of Control (IoC)** container for Go. It provides a way to manage dependencies in your application while following the **lazy initialization** and **singleton** design pattern. With `gioc`, you can ensure that each component is initialized only once and reused whenever needed, helping to manage your application's dependencies efficiently.

## Features

- **Lazy Initialization**: Dependencies are only initialized when they are first needed.
- **Singleton Pattern**: Each dependency is created only once and shared throughout the application.
- **Type Safety**: Uses Go's type system to ensure correct dependency injection.
- **Simple API**: Easy-to-use and minimalistic interface to manage your dependencies.

## Installation

To install `gioc`, you can use Go's `go get`:

```bash
go get github.com/mstgnz/gioc
```

## Usage

### Basic Example

Here's a simple example demonstrating how to use `gioc` to manage your components.

```go
package main

import (
	"fmt"
	"github.com/mstgnz/gioc"
)

// UserHandler - A handler for user-related operations
type UserHandler struct {
	service string
}

// NewUserHandler - Creates a new UserHandler instance
func NewUserHandler() *UserHandler {
	fmt.Println("Creating UserHandler...")
	return &UserHandler{service: "UserService"}
}

func main() {
	// Lazy load the UserHandler with IoC
	handler := gioc.IOC(NewUserHandler)

	// You can use the handler as needed
	fmt.Println("UserHandler service:", handler.service)
}
```

In this example:

- The `NewUserHandler` function is provided to `gioc.IOC()`.
- The `UserHandler` will be created only when it's needed (lazy initialization), and only one instance will exist throughout the application's lifetime (singleton).

### How It Works

- **IOC Function**: `gioc.IOC()` accepts a function that returns an instance of a component and ensures that it is initialized only once.
- **Return Singleton Instance**: If the instance has already been created, `gioc` will return the same instance.
- **Thread-Safe**: The container uses Go's `sync.Once` to ensure that initialization occurs only once.

### Registering and Using Other Components

You can easily extend this pattern to work with other components like services, repositories, etc. Just create the component initialization functions and call them with `gioc.IOC()`.

```go
type UserService struct {
	repository string
}

func NewUserService() *UserService {
	fmt.Println("Creating UserService...")
	return &UserService{repository: "UserRepository"}
}

func main() {
	// Lazy load and get the same instance of UserService
	service := gioc.IOC(NewUserService)

	fmt.Println("UserService repository:", service.repository)
}
```

### Advanced Examples

#### 1. Dependency Injection with Multiple Components

```go
package main

import (
	"fmt"
	"github.com/mstgnz/gioc"
)

// Database represents a database connection
type Database struct {
	connection string
}

func NewDatabase() *Database {
	fmt.Println("Initializing database connection...")
	return &Database{connection: "postgresql://localhost:5432/mydb"}
}

// UserRepository handles user data operations
type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	fmt.Println("Creating UserRepository...")
	return &UserRepository{db: db}
}

// UserService handles user business logic
type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	fmt.Println("Creating UserService...")
	return &UserService{repo: repo}
}

func main() {
	// Initialize dependencies in the correct order
	db := gioc.IOC(NewDatabase)
	repo := gioc.IOC(func() *UserRepository { return NewUserRepository(db) })
	service := gioc.IOC(func() *UserService { return NewUserService(repo) })

	fmt.Printf("Service initialized with repository and database: %+v\n", service)
}
```

#### 2. Interface-based Dependency Injection

```go
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
```

### Best Practices

1. **Dependency Order**

   - Initialize dependencies in the correct order
   - Use function literals for complex dependency chains
   - Avoid circular dependencies

2. **Error Handling**

   - Handle initialization errors gracefully
   - Consider using error return values in factory functions
   - Implement proper cleanup mechanisms

3. **Testing**

   - Use interfaces for better testability
   - Consider creating test doubles (mocks) for dependencies
   - Use dependency injection to swap implementations

4. **Performance**

   - Keep initialization functions lightweight
   - Avoid expensive operations during initialization
   - Use lazy initialization for resource-intensive components

5. **Thread Safety**
   - Ensure thread-safe initialization of shared resources
   - Handle concurrent access to shared state
   - Use proper synchronization mechanisms

### Extended API Reference

#### Core Functions

```go
// IOC registers and retrieves a singleton instance of a component
func IOC[T any](fn func() T) T

// Reset clears all registered instances (useful for testing)
func Reset()
```

#### Type Parameters

- `T`: The type of the component being registered/retrieved
- Must be a concrete type (not an interface)

#### Return Values

- Returns the singleton instance of type `T`
- Thread-safe: guaranteed to return the same instance across goroutines

#### Usage Patterns

1. **Simple Registration**

```go
instance := gioc.IOC(NewComponent)
```

2. **Dependency Chain**

```go
instance := gioc.IOC(func() *Component {
	return NewComponent(dependency)
})
```

3. **Interface Implementation**

```go
var instance Interface = gioc.IOC(NewConcreteType)
```

4. **Testing Setup**

```go
// Before each test
gioc.Reset()

// Test code
instance := gioc.IOC(NewTestComponent)
```

### Common Pitfalls and Solutions

1. **Circular Dependencies**

   - Problem: Components depending on each other
   - Solution: Break the cycle using interfaces or restructuring

2. **Resource Management**

   - Problem: Unmanaged resources in singletons
   - Solution: Implement cleanup methods and proper resource management

3. **Testing Difficulties**

   - Problem: Hard to mock dependencies
   - Solution: Use interfaces and dependency injection

4. **Concurrency Issues**
   - Problem: Race conditions in initialization
   - Solution: Rely on gioc's built-in thread safety

## Contributing

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, feel free to open an issue or create a pull request.

## License

`gioc` is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
