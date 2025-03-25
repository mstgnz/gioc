package main

import (
	"fmt"

	"github.com/mstgnz/gioc"
)

// ServiceA depends on ServiceB
type ServiceA struct {
	serviceB *ServiceB
}

func NewServiceA(serviceB *ServiceB) *ServiceA {
	return &ServiceA{serviceB: serviceB}
}

// ServiceB depends on ServiceA (creating a cycle)
type ServiceB struct {
	serviceA *ServiceA
}

func NewServiceB(serviceA *ServiceA) *ServiceB {
	return &ServiceB{serviceA: serviceA}
}

// ServiceC depends on ServiceD
type ServiceC struct {
	serviceD *ServiceD
}

func NewServiceC(serviceD *ServiceD) *ServiceC {
	return &ServiceC{serviceD: serviceD}
}

// ServiceD depends on ServiceC (creating a cycle)
type ServiceD struct {
	serviceC *ServiceC
}

func NewServiceD(serviceC *ServiceC) *ServiceD {
	return &ServiceD{serviceC: serviceC}
}

func main() {
	// Example 1: Direct circular dependency
	fmt.Println("=== Testing Direct Circular Dependency ===")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Expected panic: %v\n", r)
		}
	}()

	// This will panic with a circular dependency error
	serviceA := gioc.IOC(func() *ServiceA {
		serviceB := gioc.IOC(func() *ServiceB {
			serviceA := gioc.IOC(func() *ServiceA { return NewServiceA(nil) }, gioc.Transient)
			return NewServiceB(serviceA)
		})
		return NewServiceA(serviceB)
	})

	// This line will never be reached
	fmt.Println(serviceA)

	// Example 2: Indirect circular dependency
	fmt.Println("\n=== Testing Indirect Circular Dependency ===")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Expected panic: %v\n", r)
		}
	}()

	// This will panic with a circular dependency error
	serviceC := gioc.IOC(func() *ServiceC {
		serviceD := gioc.IOC(func() *ServiceD {
			serviceC := gioc.IOC(func() *ServiceC { return NewServiceC(nil) }, gioc.Transient)
			return NewServiceD(serviceC)
		})
		return NewServiceC(serviceD)
	})

	// This line will never be reached
	fmt.Println(serviceC)

	// Example 3: Valid dependency chain
	fmt.Println("\n=== Testing Valid Dependency Chain ===")

	// Define types for the valid dependency chain
	type Database struct {
		connection string
	}

	type Repository struct {
		db *Database
	}

	type Service struct {
		repo *Repository
	}

	// Define factory functions
	newDatabase := func() *Database {
		return &Database{connection: "localhost:5432"}
	}

	newRepository := func(db *Database) *Repository {
		return &Repository{db: db}
	}

	newService := func(repo *Repository) *Service {
		return &Service{repo: repo}
	}

	// This should work without any cycle detection errors
	service := gioc.IOC(func() *Service {
		db := gioc.IOC(newDatabase)
		repo := gioc.IOC(func() *Repository { return newRepository(db) })
		return newService(repo)
	})

	fmt.Printf("Service initialized successfully: %+v\n", service)
}
