# Minimal Reflection Example

This example demonstrates approaches to dependency injection that minimize the use of reflection, providing better performance and type safety.

## What it demonstrates

- Using `DirectIOC` for reduced reflection overhead
- Type-safe dependency injection with explicit type parameters
- Using `TypedInjectConstructor` functions for strongly-typed dependency resolution
- Factory function creation with `CreateFactory` helpers
- Managing multi-level dependencies with minimal reflection

## Code explanation

The example shows four different approaches to dependency injection with minimal reflection:

1. **DirectIOC**

   - Uses the `DirectIOC` function that reduces reflection overhead
   - Manual wiring of dependencies
   - Provides full control over the dependency resolution process

2. **TypedInjectConstructor2**

   - Uses generics to provide strong typing for dependency injection
   - Automatically resolves and injects dependencies
   - Compile-time type checking for better safety

3. **CreateFactory**

   - Creates strongly-typed factory functions
   - Registers the factory with the IoC container
   - Combines the benefits of lazy initialization with type safety

4. **Three-level Dependency**
   - Demonstrates managing complex dependency chains with `TypedInjectConstructor3`
   - Shows how to handle services that depend on other services
   - Minimizes reflection through the entire dependency chain

Each approach reduces the amount of runtime reflection used, providing better performance, stronger type safety, and improved IDE support compared to more reflection-heavy approaches.

## Output

When run, you'll see:

- Four different approaches to create services with dependencies
- How each approach initializes and wires dependencies
- The shared instances being reused across different approaches
- Type-safe dependency resolution with minimal reflection overhead
