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
	"time"
)

// BeginScope creates and activates a new scope context.
// Any subsequent requests for scoped instances will be resolved within this scope.
//
// Returns a cleanup function that should be called when the scope ends to properly
// clean up resources.
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    cleanup := gioc.BeginScope()
//	    defer cleanup()
//
//	    // Get scoped instance
//	    requestService := gioc.IOC(NewRequestService, gioc.Scoped)
//	    // Use requestService...
//	}
func BeginScope() func() {
	scopeContextMutex.Lock()
	defer scopeContextMutex.Unlock()

	previousScope := currentScopeContext
	currentScopeContext = NewScopeContext()

	return func() {
		scopeContextMutex.Lock()
		defer scopeContextMutex.Unlock()

		// Cleanup the scope
		if currentScopeContext != nil {
			currentScopeContext.Cleanup()
		}

		// Restore previous scope
		currentScopeContext = previousScope
	}
}

// GetActiveScope returns the ID of the current active scope.
// Returns an empty string if no scope is active.
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    cleanup := gioc.BeginScope()
//	    defer cleanup()
//
//	    scopeID := gioc.GetActiveScope()
//	    fmt.Printf("Active scope: %s\n", scopeID)
//	}
func GetActiveScope() string {
	scopeContextMutex.RLock()
	defer scopeContextMutex.RUnlock()

	if currentScopeContext == nil {
		return ""
	}
	return string(currentScopeContext.id)
}

// ListScopedInstances prints all instances in the current scope.
// If no scope is active, it prints a message indicating that no scope is active.
//
// Example:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    cleanup := gioc.BeginScope()
//	    defer cleanup()
//
//	    // Get scoped instance
//	    requestService := gioc.IOC(NewRequestService, gioc.Scoped)
//
//	    // List all instances in the current scope
//	    gioc.ListScopedInstances()
//	}
func ListScopedInstances() {
	scopeCtx := getCurrentScopeContext()
	if scopeCtx == nil {
		fmt.Println("No active scope")
		return
	}

	scopeCtx.mu.RLock()
	defer scopeCtx.mu.RUnlock()

	fmt.Printf("Instances in scope %s:\n", scopeCtx.id)
	if len(scopeCtx.instances) == 0 {
		fmt.Println("  No instances in this scope")
		return
	}

	for key, instance := range scopeCtx.instances {
		instanceType := reflect.TypeOf(instance)
		fmt.Printf("  Key: %v, Type: %v, Instance: %v\n", key, instanceType, instance)
	}
}

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
	once.Do(initializeContainer)

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

	// For Scoped scope, check if we're in a scope
	if componentScope == Scoped {
		scopeCtx := getCurrentScopeContext()
		if scopeCtx != nil {
			// Try to get from current scope
			if instance, exists := scopeCtx.Get(fnPtr); exists {
				if typed, ok := instance.(T); ok {
					return typed
				}
				funcName := runtime.FuncForPC(fnPtr).Name()
				panic(fmt.Sprintf("type assertion failed in scoped instance: expected %T, got %T for function %s", *new(T), instance, funcName))
			}

			// Create new instance for this scope
			// Add to resolution path for cycle detection
			currentPath := getCurrentResolutionPath()
			newPath := append(append([]uintptr(nil), currentPath...), fnPtr)
			updateResolutionPath(newPath)

			instance := fn()

			// Remove from resolution path
			updateResolutionPath(currentPath)

			scopeCtx.Set(fnPtr, instance)
			return instance
		}
		// No active scope, behave like Transient
		return fn()
	}

	// Singleton scope handling

	// Try to get existing instance with read lock first
	mu.RLock()
	if instance, exists := instances[fnPtr]; exists {
		mu.RUnlock()
		if typed, ok := instance.(T); ok {
			return typed
		}
		funcName := runtime.FuncForPC(fnPtr).Name()
		panic(fmt.Sprintf("type assertion failed in singleton instance: expected %T, got %T for function %s", *new(T), instance, funcName))
	}
	mu.RUnlock()

	// Get the current resolution path for this goroutine
	currentPath := getCurrentResolutionPath()

	// Create a new path with the current function (deep copy to avoid modifying the original)
	newPath := append(append([]uintptr(nil), currentPath...), fnPtr)
	updateResolutionPath(newPath)

	// Create the instance before acquiring the write lock
	instance := fn()

	// Restore the previous path
	updateResolutionPath(currentPath)

	// Double-check pattern with write lock
	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if existingInstance, exists := instances[fnPtr]; exists {
		if typed, ok := existingInstance.(T); ok {
			return typed
		}
		funcName := runtime.FuncForPC(fnPtr).Name()
		panic(fmt.Sprintf("type assertion failed in singleton double-check: expected %T, got %T for function %s", *new(T), existingInstance, funcName))
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
	// Initialize the instances map only once
	once.Do(initializeContainer)

	// Get function pointer directly
	fnPtr := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Entry()

	// Check for dependency cycles the same way as IOC
	if hasCycle := checkForCycle(fnPtr); hasCycle {
		cyclePath := getCyclePath()
		panic(fmt.Sprintf("circular dependency detected: %v", cyclePath))
	}

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
		if typed, ok := instance.(T); ok {
			return typed
		}
		funcName := runtime.FuncForPC(fnPtr).Name()
		panic(fmt.Sprintf("type assertion failed in DirectIOC: expected %T, got %T for function %s", *new(T), instance, funcName))
	}
	mu.RUnlock()

	// Get the current resolution path for this goroutine
	currentPath := getCurrentResolutionPath()

	// Create a new path with the current function
	newPath := append(append([]uintptr(nil), currentPath...), fnPtr)
	updateResolutionPath(newPath)

	// Create new instance
	instance := fn()

	// Restore the previous path
	updateResolutionPath(currentPath)

	// Only store if singleton
	if componentScope == Singleton {
		mu.Lock()
		defer mu.Unlock()

		// Double-check after lock
		if existingInstance, exists := instances[fnPtr]; exists {
			if typed, ok := existingInstance.(T); ok {
				return typed
			}
			funcName := runtime.FuncForPC(fnPtr).Name()
			panic(fmt.Sprintf("type assertion failed in DirectIOC double-check: expected %T, got %T for function %s", *new(T), existingInstance, funcName))
		}

		instances[fnPtr] = instance
		// Store type information for better error messages
		if _, ok := types[fnPtr]; !ok {
			types[fnPtr] = reflect.TypeOf(instance)
		}
		scopes[fnPtr] = componentScope
	}

	return instance
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

// RegisterInstance manually registers an instance with the container by type.
// This is useful when you want to provide a pre-created instance instead of a factory function.
//
// Example:
//
//	db := &Database{connection: "localhost:5432"}
//	gioc.RegisterInstance(db)
//
//	// Later, retrieve the same instance
//	sameDd := gioc.GetInstance[*Database]()
func RegisterInstance(instance interface{}) {
	// Initialize the container if not already initialized
	once.Do(initializeContainer)

	instanceType := reflect.TypeOf(instance)
	typeKey := instanceType.String() // Use the full type name as key

	typeRegistryMutex.Lock()
	defer typeRegistryMutex.Unlock()

	// Store in the type registry
	typeRegistry[typeKey] = instance
}

// GetInstance retrieves a registered instance by type.
// This is useful when you want to get an instance without providing a factory function.
//
// Example:
//
//	db := &Database{connection: "localhost:5432"}
//	gioc.RegisterInstance(db)
//
//	// Later, retrieve the same instance
//	sameDd := gioc.GetInstance[*Database]()
func GetInstance[T any]() T {
	// Initialize the container if not already initialized
	once.Do(initializeContainer)

	// Get the type of T
	var zero T
	instanceType := reflect.TypeOf(zero)
	if instanceType == nil {
		// For interface types or nil, use the type information from reflect
		instanceType = reflect.TypeOf((*T)(nil)).Elem()
	}

	typeKey := instanceType.String()

	typeRegistryMutex.RLock()
	instance, exists := typeRegistry[typeKey]
	typeRegistryMutex.RUnlock()

	if !exists {
		panic(fmt.Sprintf("no instance registered for type %v", instanceType))
	}

	// Convert to the correct type
	if typed, ok := instance.(T); ok {
		return typed
	}

	panic(fmt.Sprintf("type assertion failed: expected %T, got %T", zero, instance))
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

// MemoryStats returns statistics about the container's memory usage
func MemoryStats() map[string]int {
	mu.RLock()
	paramNameCacheMutex.RLock()
	directMutex.RLock()

	stats := map[string]int{
		"instances":         len(instances),
		"types":             len(types),
		"scopes":            len(scopes),
		"dependencyGraph":   len(dependencyGraph),
		"paramNameCache":    len(paramNameCache),
		"directInstances":   len(directInstances),
		"currentPathCap":    cap(getCurrentResolutionPath()),
		"currentPathLen":    len(getCurrentResolutionPath()),
		"tempPathBufferCap": cap(tempPathBuffer),
	}

	directMutex.RUnlock()
	paramNameCacheMutex.RUnlock()
	mu.RUnlock()

	return stats
}

// CompactMaps compacts the internal maps to reduce memory usage
// This is helpful after removing many instances
func CompactMaps() {
	mu.Lock()
	defer mu.Unlock()

	// Maps don't have a cap() function, so we'll use a threshold for compaction
	// Only compact if maps have at least this many entries deleted
	const deletionThreshold = 100

	// Check if container had significant churn
	totalSize := len(instances) + len(types) + len(scopes) + len(dependencyGraph)

	if totalSize > deletionThreshold {
		// Create new maps to compact memory usage
		newInstances := make(map[uintptr]any, len(instances))
		for k, v := range instances {
			newInstances[k] = v
		}
		instances = newInstances

		newTypes := make(map[uintptr]reflect.Type, len(types))
		for k, v := range types {
			newTypes[k] = v
		}
		types = newTypes

		newScopes := make(map[uintptr]Scope, len(scopes))
		for k, v := range scopes {
			newScopes[k] = v
		}
		scopes = newScopes

		newDependencyGraph := make(map[uintptr]map[uintptr]bool, len(dependencyGraph))
		for k, v := range dependencyGraph {
			newNodeDeps := make(map[uintptr]bool, len(v))
			for dep, val := range v {
				newNodeDeps[dep] = val
			}
			newDependencyGraph[k] = newNodeDeps
		}
		dependencyGraph = newDependencyGraph
	}

	// Compact parameter name cache
	paramNameCacheMutex.Lock()
	if len(paramNameCache) > deletionThreshold {
		newParamNameCache := make(map[uintptr][]string, len(paramNameCache))
		for k, v := range paramNameCache {
			newParamNameCache[k] = v
		}
		paramNameCache = newParamNameCache
	}
	paramNameCacheMutex.Unlock()
}

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
	// Initialize the container if not already initialized
	once.Do(initializeContainer)

	// Create options with preallocated map to reduce allocations
	options := &ConstructorOptions{
		Dependencies: make(map[string]interface{}, len(opts)),
	}
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

	// Create a map for instance type lookups to avoid unnecessary iterations
	var instanceTypeMap map[reflect.Type]reflect.Value

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

		// Lazy initialize the instance type map only when needed
		if instanceTypeMap == nil {
			instanceTypeMap = make(map[reflect.Type]reflect.Value)
			mu.RLock()
			for _, instance := range instances {
				instType := reflect.TypeOf(instance)
				instanceTypeMap[instType] = reflect.ValueOf(instance)
			}
			mu.RUnlock()
		}

		// Try to find a matching instance by type (more efficient than looping through all instances)
		if val, ok := instanceTypeMap[paramType]; ok {
			args[i] = val
			found = true
		} else {
			// If no exact match, check for assignable types
			for t, val := range instanceTypeMap {
				if t.AssignableTo(paramType) {
					args[i] = val
					found = true
					break
				}
			}
		}

		if !found {
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

	resultInterface := result[0].Interface()
	castedResult, ok := resultInterface.(T)
	if !ok {
		panic(fmt.Sprintf("type assertion failed in InjectConstructor: expected %T, got %T", *new(T), resultInterface))
	}

	return castedResult
}

// RegisterType directly registers an instance by type
func RegisterType(instance interface{}) {
	// Get the type name as key
	typ := reflect.TypeOf(instance)
	key := typ.String()

	// Store the instance
	directMutex.Lock()
	directInstances[key] = instance
	directMutex.Unlock()
}

// GetType retrieves an instance by type
func GetType[T any]() T {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ == nil {
		// Handle interface or nil
		typ = reflect.TypeOf((*T)(nil)).Elem()
	}
	key := typ.String()

	directMutex.RLock()
	instance, exists := directInstances[key]
	directMutex.RUnlock()

	if !exists {
		panic(fmt.Sprintf("No instance registered for type %s", key))
	}

	// Type assert
	result, ok := instance.(T)
	if !ok {
		panic(fmt.Sprintf("Type assertion failed: expected %T, got %T", zero, instance))
	}

	return result
}

// TypeCount returns the number of registered types
func TypeCount() int {
	directMutex.RLock()
	defer directMutex.RUnlock()
	return len(directInstances)
}

// ClearInstances removes all instances from the container.
// This is primarily useful for testing.
//
// Example:
//
//	func TestMyService(t *testing.T) {
//	    // Start with a clean state
//	    gioc.ClearInstances()
//	    // Register a mock database
//	    gioc.RegisterInstance(&MockDatabase{})
//	    // Run tests...
//	}
func ClearInstances() {
	mu.Lock()
	paramNameCacheMutex.Lock()
	directMutex.Lock()
	scopeContextMutex.Lock()
	defer mu.Unlock()
	defer paramNameCacheMutex.Unlock()
	defer directMutex.Unlock()
	defer scopeContextMutex.Unlock()

	// Clear all instances
	instances = make(map[uintptr]any, 16)
	types = make(map[uintptr]reflect.Type, 16)
	scopes = make(map[uintptr]Scope, 16)
	dependencyGraph = make(map[uintptr]map[uintptr]bool, 16)

	// Clear parameter name cache
	paramNameCache = make(map[uintptr][]string)

	// Clear direct instances
	directInstances = make(map[string]interface{})

	// Clear type registry
	typeRegistryMutex.Lock()
	typeRegistry = make(map[string]any)
	typeRegistryMutex.Unlock()

	// Clear all resolution paths - use the thread-safe method
	clearAllResolutionPaths()

	// Clear any active scope context
	if currentScopeContext != nil {
		currentScopeContext.Cleanup()
		currentScopeContext = nil
	}
}

// WithScope executes the provided function within a new scope.
// It automatically creates a new scope before executing the function and
// cleans up the scope after the function completes, regardless of whether
// the function panics or not.
//
// Example:
//
//	gioc.WithScope(func() {
//	    // This code runs within a scope
//	    service := gioc.IOC(NewRequestService, gioc.Scoped)
//	    // Use service...
//	})
func WithScope(fn func()) {
	cleanup := BeginScope()
	defer cleanup()

	fn()
}

// NewScopeContext creates a new scope context
func NewScopeContext() *ScopeContext {
	// Using time and a unique value (nanoseconds) for scope ID
	// Also adding a unique counter value to ensure uniqueness for nested scopes

	// Get unique counter value
	scopeCounterMutex.Lock()
	scopeCounter++
	uniqueCounter := scopeCounter
	scopeCounterMutex.Unlock()

	scopeID := fmt.Sprintf("scope-%d-%d", time.Now().UnixNano(), uniqueCounter)

	return &ScopeContext{
		id:        ScopeID(scopeID),
		instances: make(map[uintptr]any),
	}
}

// ListDependencyStatus prints details about the current dependency resolution state.
// This includes information about active resolution paths and cached type information.
// This function is intended for debugging purposes only.
//
// Example:
//
//	func debugMyApp() {
//	    // Print dependency status
//	    gioc.ListDependencyStatus()
//	}
func ListDependencyStatus() {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("IoC Container Status:")
	fmt.Println("=====================")

	// Count of active goroutines with resolution paths
	var pathCount int
	resolutionPathMap.Range(func(_, _ interface{}) bool {
		pathCount++
		return true
	})

	fmt.Printf("Active Resolution Goroutines: %d\n", pathCount)
	fmt.Printf("Registered Types: %d\n", len(types))

	fmt.Println("\nType Registry:")
	for key, t := range types {
		fmt.Printf("  Key: %v, Type: %v\n", key, t)
	}
}
