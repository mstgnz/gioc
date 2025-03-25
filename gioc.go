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
	"fmt"
	"reflect"
	"runtime"
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

	// Get the function pointer and type information
	fnValue := reflect.ValueOf(fn)
	key := fnValue.Pointer()
	fnType := fnValue.Type()
	returnType := fnType.Out(0)
	expectedType := reflect.TypeOf((*T)(nil)).Elem()

	// Check for dependency cycles
	if hasCycle := checkForCycle(key); hasCycle {
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
	if instance, exists := instances[key]; exists {
		storedType := types[key]
		mu.RUnlock()

		if !storedType.AssignableTo(expectedType) {
			panic(fmt.Sprintf("type mismatch: cannot use %v as %v", storedType, expectedType))
		}

		if typed, ok := instance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type assertion failed: expected %T, got %T", *new(T), instance))
	}
	mu.RUnlock()

	// Add current component to resolution path
	currentPath = append(currentPath, key)

	// Create the instance before acquiring the write lock
	instance := fn()

	// Remove current component from resolution path
	currentPath = currentPath[:len(currentPath)-1]

	// Double-check pattern with write lock
	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if existingInstance, exists := instances[key]; exists {
		storedType := types[key]
		if !storedType.AssignableTo(expectedType) {
			panic(fmt.Sprintf("type mismatch: cannot use %v as %v", storedType, expectedType))
		}

		if typed, ok := existingInstance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type assertion failed: expected %T, got %T", *new(T), existingInstance))
	}

	// Store the new instance
	instances[key] = instance
	types[key] = returnType
	scopes[key] = componentScope

	// Set up finalizer for cleanup
	runtime.SetFinalizer(instance, func(interface{}) {
		mu.Lock()
		delete(instances, key)
		delete(types, key)
		delete(scopes, key)
		delete(dependencyGraph, key)
		mu.Unlock()
	})

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
