# Constructor Injection Example

This example demonstrates different approaches to constructor injection using the `gioc` package.

## What it demonstrates

- Three different approaches to constructor injection
- Using the `InjectConstructor` function with `WithDependency` option
- Automatic dependency resolution
- Explicit dependency specification
- Minimizing reflection usage in dependency resolution

## Code explanation

The example shows three different ways to perform constructor injection:

1. **Approach 1: Explicit dependency injection with `WithDependency`**

   - Uses `InjectConstructor` to create a `UserService` with explicitly defined dependencies
   - Dependencies are specified by parameter name using `WithDependency`

2. **Approach 2: Using pre-registered instances**

   - Dependencies are first registered in the IoC container
   - Then they're injected using `InjectConstructor`
   - Shows how to reuse existing singleton instances

3. **Approach 3: Minimum reflection with explicit type declaration**
   - Uses a helper function (`CreateUserService`) that manually resolves dependencies
   - Wraps the constructor call with all dependencies already resolved
   - Reduces the use of reflection by relying on Go's type system
   - More type-safe approach with better IDE support

## Output

When run, you'll see:

- All three approaches create functionally equivalent `UserService` instances
- Each approach demonstrates a different way to manage dependencies
- The `ListInstances()` function shows registered instances after each approach
- The performance and type-safety characteristics of each approach
