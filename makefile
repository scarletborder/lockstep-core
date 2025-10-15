.PHONY: gen wire

# Directory for proto files
PROTO_DIR=proto/

# Output paths
PROTO_BACKEND_PATH=src/messages/pb/
OUTPUT_BIN_PATH=out


# wire tool path
WIRE_BIN=wire

# Default target, execute all generation tasks
all: gen wire build

# Generate all proto files
gen:
	@echo "Generating Go and TypeScript code from .proto files..."

	# make sure output directories exist
	mkdir -p ./src/messages/pb

	# --- Generate Go code ---
	# --proto_path specifies the search path for .proto files
	# --go_out specifies the output directory for Go code
	protoc --proto_path=$(PROTO_DIR) --go_out=$(PROTO_BACKEND_PATH) $(PROTO_DIR)/*.proto
	
	@echo "Code generation complete."

# Generate wire dependency injection code
wire:
	@echo "Generating wire dependency injection code..."
	${WIRE_BIN} ./src/app
	@echo "Wire generation complete."

# build to binary
build:
	@echo "Building the Go application..."
	go build -o $(OUTPUT_BIN_PATH)/ ./...
	@echo "Build complete. Binary is located at $(OUTPUT_BIN_PATH)/"

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_BACKEND_PATH)/*
	rm -rf $(PROTO_FRONTEND_PATH)/*
	@echo "Cleanup complete." 