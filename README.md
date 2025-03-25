# **gioc** - Go Inversion of Control (IoC) Container

[![Go Reference](https://pkg.go.dev/badge/github.com/mstgnz/gioc.svg)](https://pkg.go.dev/github.com/mstgnz/gioc)
[![Go Report Card](https://goreportcard.com/badge/github.com/mstgnz/gioc)](https://goreportcard.com/report/github.com/mstgnz/gioc)
[![License](https://img.shields.io/github/license/mstgnz/gioc)](LICENSE)
[![codecov](https://codecov.io/gh/mstgnz/gioc/branch/main/graph/badge.svg)](https://codecov.io/gh/mstgnz/gioc)

`gioc` is a simple and lightweight **Inversion of Control (IoC)** container for Go. It provides a way to manage dependencies in your application while following the **lazy initialization** and **singleton** design pattern. With `gioc`, you can ensure that each component is initialized only once and reused whenever needed, helping to manage your application's dependencies efficiently.

## Features

- **Lazy Initialization**: Dependencies are only initialized when they are first needed.
- **Singleton Pattern**: Each dependency is created only once and shared throughout the application.
- **Type Safety**: Uses Go's type system to ensure correct dependency injection.
- **Thread Safety**: Built-in synchronization mechanisms for concurrent access.
- **Simple API**: Easy-to-use and minimalistic interface to manage your dependencies.
- **Resource Cleanup**: Automatic cleanup of resources using Go's finalizer mechanism.

## Installation

To install `gioc`, you can use Go's `go get`:

```bash
go get github.com/mstgnz/gioc
```

## Quick Start

Here's a simple example to get you started:

```go
package main

import "github.com/mstgnz/gioc"

type Database struct {
    connection string
}

func NewDatabase() *Database {
    return &Database{connection: "localhost:5432"}
}

func main() {
    // Get a singleton instance of Database
    db := gioc.IOC(NewDatabase)
    // Use the database instance
}
```

## API Reference

### IOC[T]

```go
func IOC[T any](fn func() T) T
```

The main function for registering and retrieving instances. It ensures that each component is initialized only once.

**Parameters:**

- `fn`: A factory function that creates and returns an instance of type T

**Returns:**

- An instance of type T (singleton)

**Example:**

```go
type Service struct {
    name string
}

func NewService() *Service {
    return &Service{name: "my-service"}
}

func main() {
    svc := gioc.IOC(NewService)
}
```

### ListInstances

```go
func ListInstances()
```

Prints all currently registered instances in the IoC container. Useful for debugging.

**Example:**

```go
func main() {
    db := gioc.IOC(NewDatabase)
    svc := gioc.IOC(NewService)
    gioc.ListInstances()
}
```

### ClearInstances

```go
func ClearInstances()
```

Removes all registered instances from the IoC container. Use with caution as it's not thread-safe.

**Example:**

```go
func TestCleanup(t *testing.T) {
    db := gioc.IOC(NewDatabase)
    gioc.ClearInstances()
    if gioc.GetInstanceCount() != 0 {
        t.Error("Container should be empty")
    }
}
```

### GetInstanceCount

```go
func GetInstanceCount() int
```

Returns the number of currently registered instances.

**Example:**

```go
func main() {
    db := gioc.IOC(NewDatabase)
    svc := gioc.IOC(NewService)
    count := gioc.GetInstanceCount()
    fmt.Printf("Number of instances: %d\n", count)
}
```

## Advanced Usage

### Dependency Injection Example

```go
type UserService struct {
    db *Database
}

func NewUserService(db *Database) *UserService {
    return &UserService{db: db}
}

func main() {
    // Create database instance
    db := gioc.IOC(NewDatabase)
    // Create user service with database dependency
    userService := gioc.IOC(func() *UserService {
        return NewUserService(db)
    })
}
```

### Interface-based Example

```go
type Database interface {
    Connect() error
    Query(string) ([]byte, error)
}

type PostgresDB struct {
    connection string
}

func NewPostgresDB() Database {
    return &PostgresDB{connection: "localhost:5432"}
}

func main() {
    var db Database
    db = gioc.IOC(NewPostgresDB)
}
```

## Best Practices

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

## Common Pitfalls and Solutions

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

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, please:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

`gioc` is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Support

If you encounter any issues or have questions, please open an issue on the GitHub repository.

## Testing

The project includes comprehensive test coverage and benchmarks:

### Test Coverage

Run tests with coverage:

```bash
go test -v -race -cover ./...
```

Current test coverage is available at [Codecov](https://codecov.io/gh/mstgnz/gioc).

### Benchmarks

Run benchmarks:

```bash
go test -bench=. -benchmem ./...
```

Available benchmark tests:

- `BenchmarkIOC`: Basic performance test
- `BenchmarkIOCConcurrent`: Concurrent access performance
- `BenchmarkIOCMultipleTypes`: Performance with multiple types
- `BenchmarkIOCLargeStruct`: Performance with large structs
- `BenchmarkIOCWithDependencies`: Performance with dependency injection
- `BenchmarkIOCMemoryAllocation`: Memory allocation patterns

### Continuous Integration

The project uses GitHub Actions for continuous integration:

- Runs tests on every push and pull request
- Generates and uploads test coverage reports
- Performs race condition detection
- Runs benchmarks
