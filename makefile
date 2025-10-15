.PHONY: gen wire

# Directory for proto files
PROTO_DIR=proto/

# Output paths
PROTO_BACKEND_PATH=src/messages/pb/

# wire tool path
WIRE_BIN=~/go/bin/wire

# Default target, execute all generation tasks
all: gen wire

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

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_BACKEND_PATH)/*
	rm -rf $(PROTO_FRONTEND_PATH)/*
	@echo "Cleanup complete." 