build: build-wolf-hook build-wolf-input

build-wolf-hook:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o _output/wolf-hook

build-wolf-input:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o _output/wolf-input cmd/wolf-input/main.go
