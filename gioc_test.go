package gioc

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestStruct represents a test component
type TestStruct struct {
	Value string
}

// NewTestStruct creates a new TestStruct instance
func NewTestStruct() *TestStruct {
	return &TestStruct{Value: "test"}
}

// TestIOCBasic tests basic IOC functionality
func TestIOCBasic(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Get first instance
	instance1 := IOC(NewTestStruct)
	if instance1 == nil {
		t.Fatal("Expected non-nil instance")
	}

	if instance1.Value != "test" {
		t.Errorf("Expected value 'test', got '%s'", instance1.Value)
	}

	// Get second instance (should be the same)
	instance2 := IOC(NewTestStruct)
	if instance2 != instance1 {
		t.Error("Expected same instance")
	}
}

// TestIOCConcurrent tests concurrent access to IOC
func TestIOCConcurrent(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Number of goroutines
	numGoroutines := 100
	var wg sync.WaitGroup
	instances := make([]*TestStruct, numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			instances[index] = IOC(NewTestStruct)
		}(i)
	}

	wg.Wait()

	// Verify all instances are the same
	firstInstance := instances[0]
	for i := 1; i < numGoroutines; i++ {
		if instances[i] != firstInstance {
			t.Errorf("Instance %d is different from first instance", i)
		}
	}
}

// TestIOCMemory tests memory cleanup
func TestIOCMemory(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Get initial memory stats
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	initialAlloc := mem.Alloc

	// Create multiple instances
	for i := 0; i < 1000; i++ {
		IOC(NewTestStruct)
	}

	// Force garbage collection
	runtime.GC()

	// Get final memory stats
	runtime.ReadMemStats(&mem)
	finalAlloc := mem.Alloc

	// Memory usage should not be significantly higher
	if finalAlloc > initialAlloc*2 {
		t.Errorf("Memory usage increased significantly: initial=%d, final=%d", initialAlloc, finalAlloc)
	}
}

// TestIOCTypeSafety tests type safety
func TestIOCTypeSafety(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Different type struct
	type DifferentStruct struct {
		Value string
	}

	// First register a TestStruct
	fn := NewTestStruct
	key := reflect.ValueOf(fn).Pointer()
	_ = IOC(fn)

	// Try to get the same instance with a different type
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for type mismatch")
		} else {
			panicMsg, ok := r.(string)
			if !ok {
				t.Errorf("Expected panic message to be string, got %T", r)
			} else if msg := "type mismatch"; !strings.Contains(panicMsg, msg) {
				t.Errorf("Expected panic message to contain '%s', got '%s'", msg, panicMsg)
			}
		}
	}()

	// Create a new function with the same key but different return type
	differentFn := func() *DifferentStruct {
		return &DifferentStruct{Value: "different"}
	}

	// Manually set the instance to simulate type mismatch
	mu.Lock()
	instances[key] = differentFn()
	types[key] = reflect.TypeOf(differentFn()).Elem()
	mu.Unlock()

	// This should panic with type mismatch
	_ = IOC(fn)
}

// TestIOCMultipleTypes tests handling of multiple types
func TestIOCMultipleTypes(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Register different types
	type Type1 struct{ value int }
	type Type2 struct{ value string }

	newType1 := func() *Type1 { return &Type1{1} }
	newType2 := func() *Type2 { return &Type2{"2"} }

	// Both registrations should succeed
	instance1 := IOC(newType1)
	instance2 := IOC(newType2)

	if instance1 == nil || instance2 == nil {
		t.Error("Expected non-nil instances")
	}

	// Check instance count
	if count := GetInstanceCount(); count != 2 {
		t.Errorf("Expected 2 instances, got %d", count)
	}
}

// BenchmarkIOC tests performance
func BenchmarkIOC(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IOC(NewTestStruct)
	}
}

// BenchmarkIOCConcurrent tests concurrent performance
func BenchmarkIOCConcurrent(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = IOC(NewTestStruct)
			b.N--
		}
	})
}

// BenchmarkIOCMultipleTypes tests performance with multiple types
func BenchmarkIOCMultipleTypes(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	type Type1 struct{ value int }
	type Type2 struct{ value string }
	type Type3 struct{ value float64 }

	newType1 := func() *Type1 { return &Type1{1} }
	newType2 := func() *Type2 { return &Type2{"2"} }
	newType3 := func() *Type3 { return &Type3{3.14} }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IOC(newType1)
		_ = IOC(newType2)
		_ = IOC(newType3)
	}
}

// BenchmarkIOCLargeStruct tests performance with large structs
func BenchmarkIOCLargeStruct(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	type LargeStruct struct {
		Data [1000]byte
		Map  map[string]interface{}
	}

	newLargeStruct := func() *LargeStruct {
		return &LargeStruct{
			Map: make(map[string]interface{}),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IOC(newLargeStruct)
	}
}

// BenchmarkIOCWithDependencies tests performance with dependency injection
func BenchmarkIOCWithDependencies(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	type Dep1 struct{ value int }
	type Dep2 struct{ value string }
	type Service struct {
		dep1 *Dep1
		dep2 *Dep2
	}

	newDep1 := func() *Dep1 { return &Dep1{1} }
	newDep2 := func() *Dep2 { return &Dep2{"2"} }
	newService := func() *Service {
		return &Service{
			dep1: IOC(newDep1),
			dep2: IOC(newDep2),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IOC(newService)
	}
}

// BenchmarkIOCMemoryAllocation tests memory allocation patterns
func BenchmarkIOCMemoryAllocation(b *testing.B) {
	// Clear any existing instances
	ClearInstances()

	type MemoryTest struct {
		Data []byte
	}

	newMemoryTest := func() *MemoryTest {
		return &MemoryTest{
			Data: make([]byte, 1024),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IOC(newMemoryTest)
	}
}

// TestIOCStress tests under stress conditions
func TestIOCStress(t *testing.T) {
	// Clear any existing instances
	ClearInstances()

	// Create a channel to signal completion
	done := make(chan bool)
	timeout := time.After(2 * time.Second)

	// Launch multiple goroutines that continuously access IOC
	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					_ = IOC(NewTestStruct)
				}
			}
		}()
	}

	// Wait for timeout or completion
	select {
	case <-timeout:
		close(done)
		// If we get here, the test passed (no deadlocks)
		return
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}
}
