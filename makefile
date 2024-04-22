build:
	@make build-windows
	@make build-linux

build-windows:
	@env GOOS=windows GOARCH=amd64 go build -o dist/

build-linux:
	@env GOOS=linux GOARCH=amd64 go build -o dist/

hello:
	@echo "Hello, World"