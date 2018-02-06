GO=`which go`
GOTEST=${GO} list -f '{{if len .TestGoFiles}}"go test -cover {{.ImportPath}}"{{end}}' ./... | xargs -L 1 sh -c

# Show available make subcommands
default:
	@echo "Please supply one of:"
	@echo "\ttest\t- Run test suite"

# Run the go test suite.
test:
	@echo "Running tests ..."
	@${GOTEST}