build:
	go build -o ./tinymedia ./cmd/tinymedia/

test:
	go test ./... | grep -v 'no test files'

