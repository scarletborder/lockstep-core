.PHONY: gen

# Directory for proto files
PROTO_DIR=proto/

# Output paths
PROTO_BACKEND_PATH=src/messages/pb/

# Default target, execute all generation tasks
all: gen

# Generate all proto files
gen:
	@echo "Generating Go and TypeScript code from .proto files..."
	
	# --- Generate Go code ---
	# --proto_path specifies the search path for .proto files
	# --go_out specifies the output directory for Go code
	protoc --proto_path=$(PROTO_DIR) --go_out=$(PROTO_BACKEND_PATH) $(PROTO_DIR)/*.proto
	
	@echo "Code generation complete."

# Clean generated files
clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_BACKEND_PATH)/*
	rm -rf $(PROTO_FRONTEND_PATH)/*
	@echo "Cleanup complete." 