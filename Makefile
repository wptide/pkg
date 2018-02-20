GO=`which go`
GOTEST=${GO} list -f '{{if len .TestGoFiles}}"go test -cover {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

# Show available make subcommands
default:
	@echo "Please supply one of:"
	@echo "\tdeps\t- Install dependencies."
	@echo "\ttest\t- Run test suite"

# Install dependencies.
deps:
	@echo "Installing dependencies ..."
	@glide install

# Run the go test suite.
test:
	@echo "Running tests ..."
	@${GOTEST}