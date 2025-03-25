# Dependency Injection Example

This example demonstrates how to implement explicit dependency injection using the `gioc` package.

## What it demonstrates

- Manual dependency injection pattern
- Initializing dependencies in the correct order
- Using anonymous functions to inject dependencies into constructors
- Maintaining control over the dependency resolution process

## Code explanation

The example creates three components with explicit dependencies:

1. `Database` - A simple component representing a database connection
2. `UserRepository` - Depends on `Database` for data access
3. `UserService` - Depends on `UserRepository` for business logic

Unlike the basic example where dependencies are resolved automatically through nested `IOC()` calls, this example shows how to manually control the dependency injection process:

1. First, we register the database instance with `gioc.IOC(NewDatabase)`
2. Then, we create a `UserRepository` with an anonymous function that injects the database
3. Finally, we create a `UserService` with an anonymous function that injects the repository

This approach gives more control over how dependencies are created and injected, while still leveraging the singleton behavior of the IoC container.

## Output

When run, you'll see:

- Each component is initialized in the correct order
- The dependencies are properly injected into each component
- The final service object contains the complete object graph
