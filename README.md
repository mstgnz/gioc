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

## Contributing

This project is open-source, and contributions are welcome. Feel free to contribute or provide feedback of any kind.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
