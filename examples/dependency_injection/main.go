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

func main() {
	// Initialize dependencies in the correct order
	db := gioc.IOC(NewDatabase)
	repo := gioc.IOC(func() *UserRepository { return NewUserRepository(db) })
	service := gioc.IOC(func() *UserService { return NewUserService(repo) })

	fmt.Printf("Service initialized with repository and database: %+v\n", service)
}
