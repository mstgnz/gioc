package main

import (
	"fmt"

	"github.com/mstgnz/gioc"
)

type UserRepository struct {
	db string
}

type UserService struct {
	userRepository *UserRepository
}

type UserHandler struct {
	userService *UserService
}

// Factory Function for UserRepository
func NewUserRepository() *UserRepository {
	fmt.Println("NewUserRepository called")
	return &UserRepository{db: "DB Connection"}
}

// Factory Function for UserService
func NewUserService() *UserService {
	fmt.Println("NewUserService called")
	return &UserService{
		userRepository: gioc.IOC(NewUserRepository),
	}
}

// Factory Function for UserHandler
func NewUserHandler() *UserHandler {
	fmt.Println("NewUserHandler called")
	return &UserHandler{
		userService: gioc.IOC(NewUserService),
	}
}

func main() {
	// We retrieve objects using IOC
	handler1 := gioc.IOC(NewUserHandler)
	handler2 := gioc.IOC(NewUserHandler)

	// Note that different objects are retrieved, but they are created once.
	fmt.Println(handler1)
	fmt.Println(handler2)

	// List instances of registered objects
	gioc.ListInstances()
}
