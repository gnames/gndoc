# Run the tests of the project
test:
    go test -count=1 -p 1 -shuffle=on -coverprofile=coverage.txt -covermode=atomic ./...

# List available recipes
list:
    @just --list
