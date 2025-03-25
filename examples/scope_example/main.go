package main

import (
	"fmt"

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
	return &RequestContext{requestID: "req-123"}
}

func main() {
	// Example 1: Singleton scope (default)
	fmt.Println("\n=== Singleton Scope Example ===")
	db1 := gioc.IOC(NewDatabase)
	db2 := gioc.IOC(NewDatabase)
	fmt.Printf("db1 == db2: %v\n", db1 == db2) // Should be true

	// Example 2: Transient scope
	fmt.Println("\n=== Transient Scope Example ===")
	ctx1 := gioc.IOC(NewRequestContext, gioc.Transient)
	ctx2 := gioc.IOC(NewRequestContext, gioc.Transient)
	fmt.Printf("ctx1 == ctx2: %v\n", ctx1 == ctx2) // Should be false

	// Example 3: Scoped scope (currently behaves like Transient)
	fmt.Println("\n=== Scoped Scope Example ===")
	repo1 := gioc.IOC(func() *UserRepository { return NewUserRepository(db1) }, gioc.Scoped)
	repo2 := gioc.IOC(func() *UserRepository { return NewUserRepository(db1) }, gioc.Scoped)
	fmt.Printf("repo1 == repo2: %v\n", repo1 == repo2) // Should be false

	// Example 4: Dependency chain with mixed scopes
	fmt.Println("\n=== Mixed Scopes Example ===")
	service1 := gioc.IOC(func() *UserService { return NewUserService(repo1) })
	service2 := gioc.IOC(func() *UserService { return NewUserService(repo2) })
	fmt.Printf("service1 == service2: %v\n", service1 == service2) // Should be true (Singleton)

	// List all registered instances with their scopes
	fmt.Println("\n=== Registered Instances ===")
	gioc.ListInstances()
}
