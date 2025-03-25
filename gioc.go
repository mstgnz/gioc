package gioc

import (
	"fmt"
	"reflect"
	"sync"
)

var (
	once      sync.Once
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

	// Return the existing instance if available
	if instance, exists := instances[key]; exists {
		if typed, ok := instance.(T); ok {
			return typed
		}
		panic(fmt.Sprintf("type mismatch: expected %T, got %T", *new(T), instance))
	}

	// Create and store the new instance
	instance := fn()
	instances[key] = instance
	return instance
}

// ListInstances lists all the registered instances
func ListInstances() {
	fmt.Println("Registered instances:")
	for key, instance := range instances {
		fmt.Printf("Key: %v, Instance: %v\n", key, instance)
	}
}
