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

### Examples

We provide several examples to demonstrate different use cases of `gioc`:

1. **Basic Example** - Shows how to use `gioc` with a simple component:

   ```bash
   go run examples/basic/main.go
   ```

2. **Dependency Injection Example** - Demonstrates how to use `gioc` with multiple dependent components:

   ```bash
   go run examples/dependency_injection/main.go
   ```

3. **Interface-based Example** - Shows how to use `gioc` with interfaces:
   ```bash
   go run examples/interface_based/main.go
   ```

### How It Works

- **IOC Function**: `gioc.IOC()` accepts a function that returns an instance of a component and ensures that it is initialized only once.
- **Return Singleton Instance**: If the instance has already been created, `gioc` will return the same instance.
- **Thread-Safe**: The container uses Go's `sync.Once` to ensure that initialization occurs only once.

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
