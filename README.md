# **gioc** - Go Inversion of Control (IoC) Container

`gioc` is a simple and lightweight **Inversion of Control (IoC)** container for Go. It provides a way to manage dependencies in your application while following the **lazy initialization** and **singleton** design pattern. With `gioc`, you can ensure that each component is initialized only once and reused whenever needed, helping to manage your application's dependencies efficiently.

## Features

- **Lazy Initialization**: Dependencies are only initialized when they are first needed.
- **Singleton Pattern**: Each dependency is created only once and shared throughout the application.
- **Type Safety**: Uses Go's type system to ensure correct dependency injection.
- **Simple API**: Easy-to-use and minimalistic interface to manage your dependencies.

## Installation

To install `gioc`, you can use Go's `go get`:

```bash
go get github.com/mstgnz/gioc
```

## Usage

### Basic Example

Here's a simple example demonstrating how to use `gioc` to manage your components.

```go
package main

import (
	"fmt"
	"github.com/mstgnz/gioc"
)

// UserHandler - A handler for user-related operations
type UserHandler struct {
	service string
}

// NewUserHandler - Creates a new UserHandler instance
func NewUserHandler() *UserHandler {
	fmt.Println("Creating UserHandler...")
	return &UserHandler{service: "UserService"}
}

func main() {
	// Lazy load the UserHandler with IoC
	handler := gioc.IOC(NewUserHandler)

	// You can use the handler as needed
	fmt.Println("UserHandler service:", handler.service)
}
```

In this example:

- The `NewUserHandler` function is provided to `gioc.IOC()`.
- The `UserHandler` will be created only when it's needed (lazy initialization), and only one instance will exist throughout the application's lifetime (singleton).

### How It Works

- **IOC Function**: `gioc.IOC()` accepts a function that returns an instance of a component and ensures that it is initialized only once.
- **Return Singleton Instance**: If the instance has already been created, `gioc` will return the same instance.
- **Thread-Safe**: The container uses Go's `sync.Once` to ensure that initialization occurs only once.

### Registering and Using Other Components

You can easily extend this pattern to work with other components like services, repositories, etc. Just create the component initialization functions and call them with `gioc.IOC()`.

```go
type UserService struct {
	repository string
}

func NewUserService() *UserService {
	fmt.Println("Creating UserService...")
	return &UserService{repository: "UserRepository"}
}

func main() {
	// Lazy load and get the same instance of UserService
	service := gioc.IOC(NewUserService)

	fmt.Println("UserService repository:", service.repository)
}
```

## Contributing

Contributions are welcome! If you have any suggestions, improvements, or bug fixes, feel free to open an issue or create a pull request.

## License

`gioc` is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
