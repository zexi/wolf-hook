build:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o _output/wolf-hook
