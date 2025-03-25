package gioc

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

var (
	once      sync.Once
	mu        sync.RWMutex
	instances = make(map[uintptr]any)
	types     = make(map[uintptr]reflect.Type)
)

// IOC registers and initializes instances of components
// It ensures each component is initialized only once
func IOC[T any](fn func() T) T {
	// Initialize the instances map only once
	once.Do(func() {
		instances = make(map[uintptr]any)
		types = make(map[uintptr]reflect.Type)
	})

	// Get the function pointer and type information
	fnValue := reflect.ValueOf(fn)
	key := fnValue.Pointer()
	fnType := fnValue.Type()
	returnType := fnType.Out(0)
	expectedType := reflect.TypeOf((*T)(nil)).Elem()

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

	// Create the instance before acquiring the write lock
	instance := fn()

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

	// Set up finalizer for cleanup
	runtime.SetFinalizer(instance, func(interface{}) {
		mu.Lock()
		delete(instances, key)
		delete(types, key)
		mu.Unlock()
	})

	return instance
}

// ListInstances lists all the registered instances
func ListInstances() {
	mu.RLock()
	defer mu.RUnlock()

	fmt.Println("Registered instances:")
	for key, instance := range instances {
		fmt.Printf("Key: %v, Type: %v, Instance: %v\n", key, types[key], instance)
	}
}

// ClearInstances removes all registered instances
func ClearInstances() {
	mu.Lock()
	defer mu.Unlock()

	instances = make(map[uintptr]any)
	types = make(map[uintptr]reflect.Type)
}

// GetInstanceCount returns the number of registered instances
func GetInstanceCount() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(instances)
}
