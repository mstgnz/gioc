# Interface-Based Example

This example demonstrates how to use the `gioc` package with interfaces for more flexible and decoupled design.

## What it demonstrates

- Using interfaces for dependency contracts
- Registering concrete implementations of interfaces
- Injecting interface dependencies into services
- Improved testability through interface-based design

## Code explanation

The example implements a simple interface-based design pattern:

1. `Logger` interface - Defines a contract for logging functionality
2. `ConsoleLogger` - A concrete implementation of the `Logger` interface
3. `Service` - A component that depends on the `Logger` interface, not a concrete implementation

The key aspect of this example is that the `Service` depends on the `Logger` interface rather than a specific implementation like `ConsoleLogger`. This pattern enables:

1. **Decoupling** - The service doesn't know or care about which specific logger implementation it uses
2. **Testability** - In tests, you can easily substitute a mock implementation of the `Logger` interface
3. **Flexibility** - You can switch logger implementations without changing the service code

## Output

When run, you'll see:

- The console logger is registered in the IoC container
- The service is created with the logger injected
- The service logs a message using the injected logger
- The message is displayed through the console logger implementation
