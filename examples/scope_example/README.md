# Scope Example

This example demonstrates how to use different lifetime scopes (Singleton, Transient, and Scoped) with the `gioc` package.

## What it demonstrates

- Using the three different lifetime scopes:
  - **Singleton**: One instance per application
  - **Transient**: New instance each time
  - **Scoped**: One instance per scope (e.g., per request)
- Creating and managing scopes
- Proper scope cleanup
- Using the `WithScope` helper function

## Code explanation

The example contains three main scenarios:

1. **Singleton Scope Example**

   - Default scope behavior where a single instance is shared across the entire application
   - Demonstrates that subsequent requests for the same service return the same instance

2. **Transient Scope Example**

   - Shows how the transient scope creates a new instance every time it's requested
   - Useful for non-shared components or when stateful instances are needed

3. **Scoped Lifetime Example**
   - Demonstrates how to create and manage scopes
   - Shows that instances are shared within a scope but not across different scopes
   - Introduces two approaches to scope management:
     - Manual scope management with `BeginScope()` and cleanup function
     - Automatic scope management with `WithScope()` helper function

Scoped lifetime is particularly useful for request-based applications (like web servers) where you want to share state within a request but isolate between different requests.

## Output

When run, you'll see:

- Instances being created only when needed
- How singleton instances are shared throughout the application
- How transient instances are created anew each time
- How scoped instances are shared within a scope but not across scopes
- Scope creation and cleanup
- Lists of instances in each scope
