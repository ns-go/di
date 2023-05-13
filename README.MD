# NS-Go Dependency Injection Library (ns-go/di)

This is a simple dependency injection library for Go, designed to make it easy to manage dependencies between components in your application.

## Features

- Automatic dependency injection using tags: By using the `di.inject:""` or `di.inject:"{registered name}"` tag, you can indicate which fields in your structs should be automatically injected by the container.
- Generic parameter registration: You can register a parameter with the container by providing its type as a generic parameter to the `Register()` method.
- Custom instantiation: If you need to customize how a particular type is instantiated, you can register a custom factory function using the `RegisterFactory()` method.
- Named instances: You can register multiple instances of the same type using different names, and then resolve them by name using the `ResolveByName()` method.

## Installation

To install NS-GO/DI, use go get:

```bash
go get github.com/ns-go/di
```


## Usage

Here's a quick example of how to use this library:

```go
package main

import (
    "fmt"
    "github.com/ns-go/di/pkg/di"
)

type Foo struct {
    Bar *Bar `di.inject:""`
}

type Bar struct {
    Baz *Baz `di.inject:"baz"`
}

type Baz struct {
    Name string
}

func main() {
    container := di.NewContainer()

    di.RegisterByName(constainer, "baz", Baz{Name: "hello world!"}, false)

    di.RegisterSingleton[Foo](constainer, true)
    di.RegisterSingleton[Bar](constainer, true)

    foo := di.Resolve[Foo](constainer)
    fmt.Println(foo.Bar.Baz.Name)
}
```

In this example, we define three types: `Baz`, `Bar`, and `Foo`. `Baz` is a simple struct with a `Name` field, while `Bar` has a single field `Baz` of type `*Baz`, and `Foo` has a single field `Bar` of type `*Bar`.

We then create a new container using `di.NewContainer()`, register instances of `Baz`, `Bar`, and `Foo` using `container.Register()`, and then use `container.Resolve()` to obtain an instance of `Foo`.

The `di.inject:""` tag is used to indicate that the `Bar` field in `Foo` and the `Baz` field in `Bar` should be automatically injected by the container.

When we run this program, it will print out "hello world!", which is the value of the `Name` field of the `Baz` instance that was injected into `Foo`.

### Instance Lifetime
```go
// Register a singleton instance of Service1.
di.RegisterSingleton[Service1](constainer, false)

// Register a new instance of Service1 every time it is resolved.
di.RegisterTransient[Service1](constainer, false)

// Register a new instance of Service1 for every new scope.
di.RegisterScoped[Service1](constainer, false)

// Register a named instance of Service1 lifetime of instance is singleton.
di.RegisterByName(constainer, "test", Service1{}, false)

// Register a instance of Service1 lifetime of instance is singleton.
di.RegisterInstance(constainer, Service1{}, false)

// Register a factory function of Service1.
di.RegisterFactory(constainer, di.Singleton, func(c di.Container) *Service1 { return &Service1{} }, false)

```

## Contributing

Contributions are welcome! To contribute to ns-go/di, fork the repository and submit a pull request.

## License

NS-GO/DI is licensed under the [MIT License](https://github.com/ns-go/di/blob/main/LICENSE).