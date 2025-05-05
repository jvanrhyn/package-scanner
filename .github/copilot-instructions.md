# Project Coding Standards

## Go (Golang) Guidelines

* Use Go 1.21+ for all development
* Follow idiomatic Go conventions (gofmt, go vet, golint)
* Keep functions small, with a single responsibility
* Prefer composition over inheritance
* Avoid global state unless explicitly justified

## Package Structure

* Organize code into clear packages (e.g., `api`, `service`, `model`, `util`)
* Use internal packages to enforce encapsulation where appropriate
* Avoid circular dependencies; refactor into interfaces or shared packages

## Interfaces and Structs

* Define interfaces based on consumer needs, not implementers
* Use struct embedding for composition
* Keep struct fields private unless used across packages
* Add validation logic to constructors where applicable

## Concurrency

* Use goroutines and channels thoughtfully—avoid overcomplicating with unnecessary concurrency
* Protect shared resources with sync primitives (e.g., `sync.Mutex`)
* Use context (`context.Context`) for cancellation and timeouts

## Naming Conventions

* Use CamelCase for exported names and mixedCaps for unexported names
* Keep names short but descriptive (e.g., `getUser`, `configPath`)
* Avoid acronyms in uppercase (`httpServer`, not `HTTPServer`)
* Interface names should describe behavior and end in `-er` (e.g., `Reader`, `Formatter`)

## Error Handling

* Always check and handle errors explicitly
* Wrap errors with context using `fmt.Errorf("context: %w", err)`
* Avoid panics except in truly exceptional circumstances
* Return early when checking for errors

## Testing

* Write table-driven tests using Go’s `testing` package
* Use `*_test.go` files colocated with the code being tested
* Mock dependencies using interfaces
* Aim for high test coverage and meaningful assertions

## Documentation

* Use GoDoc comments for public types and functions
* Keep comments up to date with code changes
* Keep the README.md file updated with project setup and usage instructions
* Use examples in comments to demonstrate usage of functions and packages
* Use `godoc` to generate documentation from comments
* Keep a changes log (CHANGELOG.md) for tracking changes and versions