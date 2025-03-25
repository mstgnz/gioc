package gioc

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// initializeContainer initializes the global container state
func initializeContainer() {
	instances = make(map[uintptr]any, 16)
	types = make(map[uintptr]reflect.Type, 16)
	scopes = make(map[uintptr]Scope, 16)
	dependencyGraph = make(map[uintptr]map[uintptr]bool, 16)
	resolutionPathMap = sync.Map{}
}

// GetCurrentScopeContext returns the current active scope context.
// Returns nil if no scope context is active.
func getCurrentScopeContext() *ScopeContext {
	scopeContextMutex.RLock()
	defer scopeContextMutex.RUnlock()
	return currentScopeContext
}

// getCurrentResolutionPath gets the current goroutine's resolution path
func getCurrentResolutionPath() []uintptr {
	resolutionPathMutex.Lock()
	defer resolutionPathMutex.Unlock()

	// Get current goroutine ID - we'll use the goroutine ID as a key
	gid := getGoroutineID()

	// Get or create the path for this goroutine
	if path, ok := resolutionPathMap.Load(gid); ok {
		return path.([]uintptr)
	}

	// Create a new path for this goroutine
	path := make([]uintptr, 0, 8)
	resolutionPathMap.Store(gid, path)
	return path
}

// updateResolutionPath updates the current goroutine's resolution path
func updateResolutionPath(path []uintptr) {
	resolutionPathMutex.Lock()
	defer resolutionPathMutex.Unlock()

	gid := getGoroutineID()
	resolutionPathMap.Store(gid, path)
}

// clearAllResolutionPaths removes all resolution paths (for ClearInstances)
func clearAllResolutionPaths() {
	// Lock to ensure no other goroutine is using resolutionPathMap
	resolutionPathMutex.Lock()
	defer resolutionPathMutex.Unlock()

	// Create a new sync.Map instead of trying to clear the existing one
	// This is more thread-safe in concurrent environments
	resolutionPathMap = sync.Map{}
}

// getGoroutineID returns a unique identifier for the current goroutine
func getGoroutineID() int64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Parse goroutine ID from the stack trace
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.ParseInt(idField, 10, 64)
	return id
}

// checkForCycle checks if adding the given key would create a cycle in the dependency graph
func checkForCycle(key uintptr) bool {
	// Get the current goroutine's resolution path
	path := getCurrentResolutionPath()

	// Make a copy to avoid race conditions with other goroutines
	pathCopy := make([]uintptr, len(path))
	copy(pathCopy, path)

	// If the key is already in the current path, we have a cycle
	for _, pathKey := range pathCopy {
		if pathKey == key {
			return true
		}
	}
	return false
}

// getCyclePath returns a string representation of the cycle path
func getCyclePath() string {
	// Get the current goroutine's resolution path
	path := getCurrentResolutionPath()

	// Make a copy to avoid race conditions
	pathCopy := make([]uintptr, len(path))
	copy(pathCopy, path)

	if len(pathCopy) == 0 {
		return "empty path"
	}

	// Find the start of the cycle
	cycleStart := 0
	for i, key := range pathCopy {
		if key == pathCopy[len(pathCopy)-1] {
			cycleStart = i
			break
		}
	}

	// Create a local buffer to avoid races with the global one
	localBuffer := make([]string, 0, 8)

	// Build the cycle path string
	for i := cycleStart; i < len(pathCopy); i++ {
		key := pathCopy[i]
		mu.RLock() // Lock while accessing the types map
		t, exists := types[key]
		mu.RUnlock()

		if exists {
			localBuffer = append(localBuffer, t.String())
		} else {
			localBuffer = append(localBuffer, fmt.Sprintf("unknown(%d)", key))
		}
	}

	return fmt.Sprintf("%v", localBuffer)
}

// getParamName returns the name of the parameter at the given index
func getParamName(fn interface{}, index int) string {
	fnPtr := reflect.ValueOf(fn).Pointer()

	// First try to get from cache
	paramNameCacheMutex.RLock()
	if params, ok := paramNameCache[fnPtr]; ok {
		paramNameCacheMutex.RUnlock()
		if index < len(params) {
			return params[index]
		}
		return fmt.Sprintf("param%d", index)
	}
	paramNameCacheMutex.RUnlock()

	// Get function file and line
	file, line := runtime.FuncForPC(fnPtr).FileLine(0)

	// Read the file
	fileHandle, err := os.Open(file)
	if err != nil {
		return fmt.Sprintf("param%d", index)
	}
	defer fileHandle.Close()

	// Create scanner with appropriate buffer size to reduce allocations
	scanner := bufio.NewScanner(fileHandle)
	// Set a larger buffer for the scanner to avoid extra allocations
	const bufferSize = 64 * 1024
	buffer := make([]byte, bufferSize)
	scanner.Buffer(buffer, bufferSize)

	currentLine := 0
	var functionLine string

	// Find the function definition
	for scanner.Scan() {
		currentLine++
		if currentLine == line {
			functionLine = scanner.Text()
			break
		}
		// No need to continue if we've passed the target line
		if currentLine > line {
			break
		}
	}

	// Extract parameter names using the precompiled regex
	matches := paramRegex.FindStringSubmatch(functionLine)
	if len(matches) != 2 {
		return fmt.Sprintf("param%d", index)
	}

	// Split parameters
	paramStr := matches[1]
	var params []string

	// Don't use strings.Split for large strings as it creates a new array
	// More efficient to parse directly
	if strings.IndexByte(paramStr, ',') == -1 {
		// Only one parameter
		params = []string{strings.TrimSpace(paramStr)}
	} else {
		// Multiple parameters
		parts := strings.Split(paramStr, ",")
		params = make([]string, 0, len(parts))
		for _, part := range parts {
			// Clean up parameter name
			part = strings.TrimSpace(part)
			if strings.Contains(part, " ") {
				nameParts := strings.Split(part, " ")
				if len(nameParts) > 1 {
					params = append(params, nameParts[1])
					continue
				}
			}
			params = append(params, part)
		}
	}

	// Make sure we actually have extracted names before caching
	if len(params) > 0 {
		// Store in cache
		paramNameCacheMutex.Lock()
		paramNameCache[fnPtr] = params
		paramNameCacheMutex.Unlock()
	}

	if index < len(params) {
		return params[index]
	}

	return fmt.Sprintf("param%d", index)
}
