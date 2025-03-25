// Package gioc provides a simple and lightweight Inversion of Control (IoC) container for Go.
// It implements lazy initialization and singleton pattern for managing application dependencies.
//
// Example:
//
//	type Database struct {
//	    connection string
//	}
//
//	func NewDatabase() *Database {
//	    return &Database{connection: "localhost:5432"}
//	}
//
//	func main() {
//	    // Get a singleton instance of Database
//	    db := gioc.IOC(NewDatabase)
//	    // Use the database instance
//	}
package gioc

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// Scope represents the lifetime of a component in the IoC container
type Scope int

const (
	// Singleton scope (default): One instance per application lifetime
	Singleton Scope = iota
	// Transient scope: New instance each time
	Transient
	// Scoped scope: One instance per scope (e.g., per request)
	Scoped
)

var (
	once      sync.Once
	mu        sync.RWMutex
	instances = make(map[uintptr]any)
	types     = make(map[uintptr]reflect.Type)
	scopes    = make(map[uintptr]Scope)
	// Track dependency graph for cycle detection
	dependencyGraph = make(map[uintptr]map[uintptr]bool)
	// Track current resolution path for cycle detection
	currentPath = make([]uintptr, 0)
)

// IOC registers and initializes instances of components using lazy initialization.
// It ensures each component is initialized only once and returns the same instance
// for subsequent calls with the same factory function.
//
// The function is thread-safe and uses double-check locking pattern for optimal performance.
//
// Type parameter T represents the type of the component to be initialized.
//
// Example:
//
//	type Service struct {
//	    name string
//	}
//
//	func NewService() *Service {
//	    return &Service{name: "my-service"}
//	}
//
//	func main() {
//	    // First call creates the instance
//	    svc1 := gioc.IOC(NewService)
//	    // Second call returns the same instance
//	    svc2 := gioc.IOC(NewService)
//	    // svc1 and svc2 are the same instance
//	}
func IOC[T any](fn func() T, scope ...Scope) T {
	// Initialize the instances map only once
	once.Do(func() {
		instances = make(map[uintptr]any)
		types = make(map[uintptr]reflect.Type)
		scopes = make(map[uintptr]Scope)
		dependencyGraph = make(map[uintptr]map[uintptr]bool)
		currentPath = make([]uintptr, 0)
	})

	// Get the function pointer using runtime instead of full reflection
	fnPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Entry()

	// Check for dependency cycles
	if hasCycle := checkForCycle(fnPtr); hasCycle {
		cyclePath := getCyclePath()
		panic(fmt.Sprintf("circular dependency detected: %v", cyclePath))
	}

	// Determine the scope (default to Singleton if not specified)
	var componentScope Scope = Singleton
	if len(scope) > 0 {
		componentScope = scope[0]
	}

	// For Transient scope, always create a new instance
	if componentScope == Transient {
		return fn()
	}

	// For Scoped scope, check if we're in a new scope
	if componentScope == Scoped {
		// TODO: Implement scope tracking
		// For now, behave like Transient
		return fn()
	}

	// Try to get existing instance with read lock first
	mu.RLock()
	if instance, exists := instances[fnPtr]; exists {
		mu.RUnlock()
		if typed, ok := instance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type assertion failed: expected %T, got %T", *new(T), instance))
	}
	mu.RUnlock()

	// Add current component to resolution path
	currentPath = append(currentPath, fnPtr)

	// Create the instance before acquiring the write lock
	instance := fn()

	// Remove current component from resolution path
	currentPath = currentPath[:len(currentPath)-1]

	// Double-check pattern with write lock
	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if existingInstance, exists := instances[fnPtr]; exists {
		if typed, ok := existingInstance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type assertion failed: expected %T, got %T", *new(T), existingInstance))
	}

	// Store the new instance
	instances[fnPtr] = instance
	// Store type information only when needed
	if _, ok := types[fnPtr]; !ok {
		types[fnPtr] = reflect.TypeOf(instance)
	}
	scopes[fnPtr] = componentScope

	// Set up finalizer for cleanup
	runtime.SetFinalizer(instance, func(interface{}) {
		mu.Lock()
		delete(instances, fnPtr)
		delete(types, fnPtr)
		delete(scopes, fnPtr)
		delete(dependencyGraph, fnPtr)
		mu.Unlock()
	})

	return instance
}

// DirectIOC is a minimal reflection version of IOC
// It provides the same functionality with less reflection use
func DirectIOC[T any](fn func() T, scope ...Scope) T {
	// Get function pointer directly
	fnPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Entry()

	// Determine scope
	var componentScope Scope = Singleton
	if len(scope) > 0 {
		componentScope = scope[0]
	}

	// For Transient scope, always create a new instance
	if componentScope == Transient {
		return fn()
	}

	// Try to get existing instance with read lock first
	mu.RLock()
	if instance, exists := instances[fnPtr]; exists {
		mu.RUnlock()
		return instance.(T)
	}
	mu.RUnlock()

	// Create new instance
	instance := fn()

	// Only store if singleton
	if componentScope == Singleton {
		mu.Lock()
		defer mu.Unlock()

		// Double-check after lock
		if existingInstance, exists := instances[fnPtr]; exists {
			return existingInstance.(T)
		}

		instances[fnPtr] = instance
	}

	return instance
}

// checkForCycle checks if adding the given key would create a cycle in the dependency graph
func checkForCycle(key uintptr) bool {
	// If the key is already in the current path, we have a cycle
	for _, pathKey := range currentPath {
		if pathKey == key {
			return true
		}
	}
	return false
}

// getCyclePath returns a string representation of the cycle path
func getCyclePath() string {
	if len(currentPath) == 0 {
		return "empty path"
	}

	// Find the start of the cycle
	cycleStart := 0
	for i, key := range currentPath {
		if key == currentPath[len(currentPath)-1] {
			cycleStart = i
			break
		}
	}

	// Build the cycle path string
	path := make([]string, 0)
	for i := cycleStart; i < len(currentPath); i++ {
		key := currentPath[i]
		if t, exists := types[key]; exists {
			path = append(path, t.String())
		} else {
			path = append(path, fmt.Sprintf("unknown(%d)", key))
		}
	}

	return fmt.Sprintf("%v", path)
}

// ListInstances prints all currently registered instances in the IoC container.
// This is useful for debugging and understanding the current state of the container.
//
// Example:
//
//	func main() {
//	    // Register some instances
//	    db := gioc.IOC(NewDatabase)
//	    svc := gioc.IOC(NewService)
//	    // List all instances
//	    gioc.ListInstances()
//	}
func ListInstances() {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("Registered instances:")
	for key, instance := range instances {
		scope := scopes[key]
		scopeName := "Singleton"
		switch scope {
		case Transient:
			scopeName = "Transient"
		case Scoped:
			scopeName = "Scoped"
		}
		fmt.Printf("Key: %v, Type: %v, Scope: %s, Instance: %v\n", key, types[key], scopeName, instance)
	}
}

// ClearInstances removes all registered instances from the IoC container.
// This function is useful for testing or when you need to reset the container state.
// Note that this operation is not thread-safe and should be used with caution.
//
// Example:
//
//	func TestCleanup(t *testing.T) {
//	    // Register some instances
//	    db := gioc.IOC(NewDatabase)
//	    // Clear all instances
//	    gioc.ClearInstances()
//	    // Verify container is empty
//	    if gioc.GetInstanceCount() != 0 {
//	        t.Error("Container should be empty")
//	    }
//	}
func ClearInstances() {
	mu.Lock()
	defer mu.Unlock()

	instances = make(map[uintptr]any)
	types = make(map[uintptr]reflect.Type)
	scopes = make(map[uintptr]Scope)
	dependencyGraph = make(map[uintptr]map[uintptr]bool)
	currentPath = make([]uintptr, 0)
}

// GetInstanceCount returns the number of currently registered instances in the IoC container.
// This is useful for monitoring and debugging purposes.
//
// Example:
//
//	func main() {
//	    // Register some instances
//	    db := gioc.IOC(NewDatabase)
//	    svc := gioc.IOC(NewService)
//	    // Get the count
//	    count := gioc.GetInstanceCount()
//	    fmt.Printf("Number of instances: %d\n", count)
//	}
func GetInstanceCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(instances)
}

// ConstructorOptions represents options for constructor injection
type ConstructorOptions struct {
	// Dependencies is a map of parameter names to their factory functions
	Dependencies map[string]interface{}
}

// ConstructorOption is a function that modifies ConstructorOptions
type ConstructorOption func(*ConstructorOptions)

// WithDependency adds a dependency to the constructor options
func WithDependency(name string, factory interface{}) ConstructorOption {
	return func(o *ConstructorOptions) {
		if o.Dependencies == nil {
			o.Dependencies = make(map[string]interface{})
		}
		o.Dependencies[name] = factory
	}
}

// TypedInjectConstructor is a less reflection heavy alternative to InjectConstructor
// It requires explicit dependency creation but avoids runtime reflection for parameter name discovery
// This approach follows the pattern from examples/constructor_injection/main.go "Approach 3"
func TypedInjectConstructor[T any, D1 any](
	constructor func(D1) T,
	dep1 func() D1,
) T {
	// Get dependencies with minimal reflection
	d1 := IOC(dep1)

	// Call constructor directly without reflection
	return constructor(d1)
}

// TypedInjectConstructor2 handles two dependencies
func TypedInjectConstructor2[T any, D1 any, D2 any](
	constructor func(D1, D2) T,
	dep1 func() D1,
	dep2 func() D2,
) T {
	// Get dependencies with minimal reflection
	d1 := IOC(dep1)
	d2 := IOC(dep2)

	// Call constructor directly without reflection
	return constructor(d1, d2)
}

// TypedInjectConstructor3 handles three dependencies
func TypedInjectConstructor3[T any, D1 any, D2 any, D3 any](
	constructor func(D1, D2, D3) T,
	dep1 func() D1,
	dep2 func() D2,
	dep3 func() D3,
) T {
	// Get dependencies with minimal reflection
	d1 := IOC(dep1)
	d2 := IOC(dep2)
	d3 := IOC(dep3)

	// Call constructor directly without reflection
	return constructor(d1, d2, d3)
}

// CreateFactory creates a factory function that uses IOC container to resolve dependencies
// This is a helper function to reduce reflection usage in common scenarios
func CreateFactory[T any, D1 any](
	constructor func(D1) T,
	dep1 func() D1,
) func() T {
	return func() T {
		d1 := IOC(dep1)
		return constructor(d1)
	}
}

// CreateFactory2 creates a factory function for two dependencies
func CreateFactory2[T any, D1 any, D2 any](
	constructor func(D1, D2) T,
	dep1 func() D1,
	dep2 func() D2,
) func() T {
	return func() T {
		d1 := IOC(dep1)
		d2 := IOC(dep2)
		return constructor(d1, d2)
	}
}

// CreateFactory3 creates a factory function for three dependencies
func CreateFactory3[T any, D1 any, D2 any, D3 any](
	constructor func(D1, D2, D3) T,
	dep1 func() D1,
	dep2 func() D2,
	dep3 func() D3,
) func() T {
	return func() T {
		d1 := IOC(dep1)
		d2 := IOC(dep2)
		d3 := IOC(dep3)
		return constructor(d1, d2, d3)
	}
}

// InjectConstructor registers and initializes instances using constructor injection.
// It automatically resolves dependencies and creates instances using the provided constructor function.
//
// Example:
//
//	type UserService struct {
//	    db *Database
//	    logger *Logger
//	}
//
//	func NewUserService(db *Database, logger *Logger) *UserService {
//	    return &UserService{db: db, logger: logger}
//	}
//
//	func main() {
//	    // Register dependencies
//	    db := gioc.IOC(NewDatabase)
//	    logger := gioc.IOC(NewLogger)
//
//	    // Create UserService with constructor injection
//	    userService := gioc.InjectConstructor(NewUserService,
//	        gioc.WithDependency("db", NewDatabase),
//	        gioc.WithDependency("logger", NewLogger),
//	    )
//	}
func InjectConstructor[T any](constructor interface{}, opts ...ConstructorOption) T {
	// Create options
	options := &ConstructorOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Get constructor function type
	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		panic("constructor must be a function")
	}

	// Get constructor parameters
	numIn := constructorType.NumIn()
	args := make([]reflect.Value, numIn)

	// Resolve each parameter
	for i := 0; i < numIn; i++ {
		paramType := constructorType.In(i)
		paramName := getParamName(constructor, i)

		// Try to get dependency from options
		if factory, exists := options.Dependencies[paramName]; exists {
			factoryValue := reflect.ValueOf(factory)
			if factoryValue.Kind() != reflect.Func {
				panic(fmt.Sprintf("dependency factory for %s must be a function", paramName))
			}

			// Call factory function
			result := factoryValue.Call(nil)
			if len(result) != 1 {
				panic(fmt.Sprintf("dependency factory for %s must return exactly one value", paramName))
			}

			// Check type compatibility
			if !result[0].Type().AssignableTo(paramType) {
				panic(fmt.Sprintf("dependency type mismatch for %s: expected %v, got %v",
					paramName, paramType, result[0].Type()))
			}

			args[i] = result[0]
			continue
		}

		// If no explicit dependency provided, try to find a registered instance
		found := false
		mu.RLock()
		for _, instance := range instances {
			instType := reflect.TypeOf(instance)
			if instType.AssignableTo(paramType) {
				args[i] = reflect.ValueOf(instance)
				found = true
				break
			}
		}
		mu.RUnlock()

		if !found {
			// Try to find a constructor with standard naming convention

			// First check if the dependency is registered with IOC
			// Try to dynamically create the constructor for the dependency
			// We'll use a workaround by creating a factory function for each dependency

			// For test mocking, we'll allow dependency lookup by type if it exists in the options
			for _, factory := range options.Dependencies {
				factoryValue := reflect.ValueOf(factory)
				if factoryValue.Kind() != reflect.Func {
					continue
				}

				result := factoryValue.Call(nil)
				if len(result) != 1 {
					continue
				}

				if result[0].Type().AssignableTo(paramType) {
					args[i] = result[0]
					found = true
					break
				}
			}

			if !found {
				panic(fmt.Sprintf("no dependency found for parameter %s of type %v", paramName, paramType))
			}
		}
	}

	// Call constructor with resolved arguments
	constructorValue := reflect.ValueOf(constructor)
	result := constructorValue.Call(args)

	if len(result) != 1 {
		panic("constructor must return exactly one value")
	}

	return result[0].Interface().(T)
}

// getParamName returns the name of the parameter at the given index
func getParamName(fn interface{}, index int) string {
	// Get function file and line
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)

	// Read the file
	fileHandle, err := os.Open(file)
	if err != nil {
		return fmt.Sprintf("param%d", index)
	}
	defer fileHandle.Close()

	// Create scanner
	scanner := bufio.NewScanner(fileHandle)
	currentLine := 0
	var functionLine string

	// Find the function definition
	for scanner.Scan() {
		currentLine++
		if currentLine == line {
			functionLine = scanner.Text()
			break
		}
	}

	// Extract parameter names using regex
	re := regexp.MustCompile(`func\s+\w+\s*\((.*?)\)`)
	matches := re.FindStringSubmatch(functionLine)
	if len(matches) != 2 {
		return fmt.Sprintf("param%d", index)
	}

	// Split parameters
	params := strings.Split(matches[1], ",")
	if index >= len(params) {
		return fmt.Sprintf("param%d", index)
	}

	// Clean up parameter name
	param := strings.TrimSpace(params[index])
	if strings.Contains(param, " ") {
		parts := strings.Split(param, " ")
		if len(parts) > 1 {
			return parts[1]
		}
	}

	return param
}
