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
- **Constructor Injection**: Support for automatic dependency resolution for constructors.
- **Multiple Scopes**: Support for Singleton, Transient, and Scoped lifetimes.

## Installation

To install `gioc`, you can use Go's `go get`:

```bash
go get github.com/mstgnz/gioc
```

## API Overview

### Core Functions

- **IOC[T]**: Main function for registering and retrieving instances.
- **InjectConstructor[T]**: Creates instances with constructor injection.
- **WithDependency**: Adds explicit dependencies to constructors.
- **ListInstances**: Prints all registered instances (for debugging).
- **ClearInstances**: Removes all registered instances.
- **GetInstanceCount**: Returns the count of registered instances.

### Scopes

- **Singleton** (default): One instance per application lifetime.
- **Transient**: New instance created each time.
- **Scoped**: One instance per scope (e.g., per request).

## Examples

For complete examples, see the [examples directory](./examples):

- [Basic Usage](./examples/basic)
- [Dependency Injection](./examples/dependency_injection)
- [Constructor Injection](./examples/constructor_injection)
- [Interface-Based Usage](./examples/interface_based)
- [Cycle Detection](./examples/cycle_detection)
- [Scope Examples](./examples/scope_example)

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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
