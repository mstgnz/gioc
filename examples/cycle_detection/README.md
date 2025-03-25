# Cycle Detection Example

This example demonstrates how the `gioc` package detects and prevents circular dependencies in your object graph.

## What it demonstrates

- Detection of direct circular dependencies
- Detection of indirect circular dependencies
- Proper handling of valid dependency chains
- Runtime safety features of the IoC container

## Code explanation

The example contains three scenarios:

1. **Direct Circular Dependency**

   - ServiceA depends on ServiceB
   - ServiceB depends on ServiceA
   - This creates a direct circular dependency that's detected by the container

2. **Indirect Circular Dependency**

   - ServiceC depends on ServiceD
   - ServiceD depends on ServiceC
   - Similar to the first example, but demonstrates the cycle detection in a different way

3. **Valid Dependency Chain**
   - A Database -> Repository -> Service chain
   - This is a valid, acyclic dependency graph that works correctly
   - Shows how proper dependencies should be structured

When a circular dependency is detected, the container throws a panic with detailed information about the cycle, helping developers identify and fix the issue.

## Output

When run, you'll see:

- Two expected panics with cycle detection messages
- A successful initialization of the valid dependency chain
- Clear error messages that identify the specific components involved in the circular dependency
- The path of the dependency cycle as part of the error message
