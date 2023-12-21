build:
	go build -o bin/linker-client ./client/client.go
	go build -o bin/linker-server ./server/server.go