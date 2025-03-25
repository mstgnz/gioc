package main

import (
	"fmt"
	"time"

	"github.com/mstgnz/gioc"
)

// Database represents a database connection
type Database struct {
	connection string
}

func NewDatabase() *Database {
	fmt.Println("Initializing database connection...")
	return &Database{connection: "postgresql://localhost:5432/mydb"}
}

// UserRepository handles user data operations
type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	fmt.Println("Creating UserRepository...")
	return &UserRepository{db: db}
}

// UserService handles user business logic
type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	fmt.Println("Creating UserService...")
	return &UserService{repo: repo}
}

// RequestContext represents a request-scoped context
type RequestContext struct {
	requestID string
}

func NewRequestContext() *RequestContext {
	fmt.Println("Creating new RequestContext...")
	return &RequestContext{requestID: fmt.Sprintf("req-%d", time.Now().UnixNano())}
}

// RequestService is a scoped service that depends on RequestContext
type RequestService struct {
	context *RequestContext
	db      *Database
}

func NewRequestService(ctx *RequestContext, db *Database) *RequestService {
	fmt.Println("Creating RequestService...")
	return &RequestService{
		context: ctx,
		db:      db,
	}
}

func main() {
	// Clear any existing instances to start fresh
	gioc.ClearInstances()

	// Example 1: Singleton scope (default)
	fmt.Println("\n=== Singleton Scope Example ===")
	db1 := gioc.IOC(NewDatabase)
	db2 := gioc.IOC(NewDatabase)
	fmt.Printf("db1 == db2: %v (should be true, singleton)\n", db1 == db2)

	// Example 2: Transient scope
	fmt.Println("\n=== Transient Scope Example ===")
	ctx1 := gioc.IOC(NewRequestContext, gioc.Transient)
	ctx2 := gioc.IOC(NewRequestContext, gioc.Transient)
	fmt.Printf("ctx1 == ctx2: %v (should be false, transient)\n", ctx1 == ctx2)

	// Example 3: Properly implemented Scoped lifetime
	fmt.Println("\n=== Scoped Lifetime Example ===")

	// First scope
	fmt.Println("\nFirst scope:")
	scopeCleanup1 := gioc.BeginScope()
	fmt.Printf("Active scope: %s\n", gioc.GetActiveScope())

	service1 := gioc.IOC(func() *RequestService {
		ctx := gioc.IOC(NewRequestContext, gioc.Scoped)
		return NewRequestService(ctx, db1)
	}, gioc.Scoped)

	// Get it again, should be the same instance within this scope
	service1Again := gioc.IOC(func() *RequestService {
		ctx := gioc.IOC(NewRequestContext, gioc.Scoped)
		return NewRequestService(ctx, db1)
	}, gioc.Scoped)

	fmt.Printf("service1 == service1Again: %v (should be true, same scope)\n", service1 == service1Again)

	// List instances in this scope
	gioc.ListScopedInstances()

	// End first scope
	scopeCleanup1()

	// Second scope
	fmt.Println("\nSecond scope:")
	// Use the WithScope helper
	gioc.WithScope(func() {
		fmt.Printf("Active scope: %s\n", gioc.GetActiveScope())

		service2 := gioc.IOC(func() *RequestService {
			ctx := gioc.IOC(NewRequestContext, gioc.Scoped)
			return NewRequestService(ctx, db1)
		}, gioc.Scoped)

		// service1 and service2 should be different as they're in different scopes
		fmt.Printf("service1 == service2: %v (should be false, different scopes)\n", service1 == service2)

		// List instances in this scope
		gioc.ListScopedInstances()
	})

	// After second scope ends
	fmt.Println("\nAfter all scopes:")
	fmt.Printf("Active scope: %s (should be empty)\n", gioc.GetActiveScope())
	gioc.ListScopedInstances() // Should say "No active scope"

	// List all registered instances
	fmt.Println("\n=== Registered Instances ===")
	gioc.ListInstances()
}
