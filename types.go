package gioc

import (
	"reflect"
	"regexp"
	"sync"
)

const (
	// Singleton scope (default): One instance per application lifetime
	Singleton Scope = iota
	// Transient scope: New instance each time
	Transient
	// Scoped scope: One instance per scope (e.g., per request)
	Scoped
)

// ConstructorOptions represents options for constructor injection
type ConstructorOptions struct {
	// Dependencies is a map of parameter names to their factory functions
	Dependencies map[string]interface{}
}

// ConstructorOption is a function that modifies ConstructorOptions
type ConstructorOption func(*ConstructorOptions)

var (
	once      sync.Once
	mu        sync.RWMutex
	instances = make(map[uintptr]any, 16) // Initialize with capacity hint
	types     = make(map[uintptr]reflect.Type, 16)
	scopes    = make(map[uintptr]Scope, 16)
	// Track dependency graph for cycle detection
	dependencyGraph = make(map[uintptr]map[uintptr]bool, 16)
	// Track current resolution path for cycle detection using goroutine-local storage
	resolutionPathMap = sync.Map{}           // map[goroutineID][]uintptr
	tempPathBuffer    = make([]string, 0, 8) // Reusable buffer for path strings

	// Precompiled regex for parameter name extraction
	paramRegex = regexp.MustCompile(`func\s+\w+\s*\((.*?)\)`)

	// paramNameCache caches parameter names to avoid repeatedly parsing the same function
	paramNameCache      = make(map[uintptr][]string)
	paramNameCacheMutex sync.RWMutex

	// typeRegistry is a separate registry for type-based instance storage
	typeRegistry      = make(map[string]any)
	typeRegistryMutex sync.RWMutex

	// Type registry for storing instances by type
	directInstances = make(map[string]interface{})
	directMutex     sync.RWMutex

	// Current active scope context
	currentScopeContext *ScopeContext
	scopeContextMutex   sync.RWMutex

	// Scope ID için statik sayaç
	scopeCounter      int
	scopeCounterMutex sync.Mutex

	// Create a mutex specifically for resolution path operations
	resolutionPathMutex sync.Mutex
)
