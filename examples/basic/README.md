# Basic Example

This example demonstrates the core functionality of the `gioc` package, showing how to use the IoC (Inversion of Control) container for basic dependency management.

## What it demonstrates

- Basic singleton pattern through the `IOC()` function
- Factory functions for object creation
- Automatic dependency resolution
- Singleton behavior - objects are created only once

## Code explanation

The example creates a simple hierarchy of objects:

1. `UserRepository` - The lowest-level component that simulates a database connection
2. `UserService` - Middle-tier service that depends on `UserRepository`
3. `UserHandler` - Top-level component that depends on `UserService`

When requesting a `UserHandler` using `gioc.IOC()`, the container automatically resolves the entire dependency chain, creating each component only once.

The code also demonstrates that multiple calls to `gioc.IOC()` with the same factory function return the same instance, confirming that the singleton pattern is working correctly.

## Output

When run, you'll see:

- Each factory function is called only once, even though we retrieve the handler twice
- Both `handler1` and `handler2` are the same instance
- The `ListInstances()` function shows all registered instances in the container
