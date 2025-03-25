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
)

// IOC registers and initializes instances of components
// It ensures each component is initialized only once
func IOC[T any](fn func() T) T {
	// Initialize the instances map only once
	once.Do(func() {
		instances = make(map[uintptr]any)
	})

	// Get the function pointer as the key
	key := reflect.ValueOf(fn).Pointer()

	// Try to get existing instance with read lock first
	mu.RLock()
	if instance, exists := instances[key]; exists {
		mu.RUnlock()
		if typed, ok := instance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type mismatch: expected %T, got %T", *new(T), instance))
	}
	mu.RUnlock()

	// Double-check pattern with write lock
	mu.Lock()
	defer mu.Unlock()

	// Check again after acquiring write lock
	if instance, exists := instances[key]; exists {
		if typed, ok := instance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type mismatch: expected %T, got %T", *new(T), instance))
	}

	// Create and store the new instance
	instance := fn()
	instances[key] = instance

	// Set up finalizer for cleanup
	runtime.SetFinalizer(instance, func(interface{}) {
		mu.Lock()
		delete(instances, key)
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
		fmt.Printf("Key: %v, Instance: %v\n", key, instance)
	}
}

// ClearInstances removes all registered instances
func ClearInstances() {
	mu.Lock()
	defer mu.Unlock()

	instances = make(map[uintptr]any)
}
