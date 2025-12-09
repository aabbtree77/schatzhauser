BIN_DIR := bin

.PHONY: all server god clean

all: server god

server:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/server ./cmd/server

god:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/god ./cmd/god

clean:
	rm -rf $(BIN_DIR)
