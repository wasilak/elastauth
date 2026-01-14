.PHONY: help init init-force up down restart logs logs-follow ps status test info clean clean-all

# Color output variables
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

# Force flag for certificate regeneration
FORCE ?= 0

# Default target
.DEFAULT_GOAL := help

# Help target - displays all available targets with descriptions
help: ## Display this help message
	@echo "$(BLUE)elastauth Demo Environment - Available Commands$(NC)"
	@echo ""
	@echo "$(GREEN)Initialization:$(NC)"
	@echo "  make init          - Bootstrap the demo environment (generate certs and configs)"
	@echo "  make init-force    - Force regenerate all certificates and configs"
	@echo ""
	@echo "$(GREEN)Service Management:$(NC)"
	@echo "  make up            - Start all services in detached mode"
	@echo "  make down          - Stop and remove all containers"
	@echo "  make restart       - Restart all services"
	@echo "  make ps            - Show running containers"
	@echo "  make logs          - Display logs from all services"
	@echo "  make logs-follow   - Follow logs in real-time"
	@echo ""
	@echo "$(GREEN)Testing and Validation:$(NC)"
	@echo "  make test          - Run health checks on all services"
	@echo "  make info          - Display connection information and credentials"
	@echo ""
	@echo "$(GREEN)Cleanup:$(NC)"
	@echo "  make clean         - Remove containers, volumes, certs, and configs (with confirmation)"
	@echo "  make clean-all     - Force clean without confirmation"
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC)"
	@echo "  1. make init       - Initialize the environment"
	@echo "  2. make up         - Start all services"
	@echo "  3. make test       - Verify services are healthy"
	@echo "  4. make info       - Get connection details"
	@echo ""

# Certificate generation functions
CERTS_DIR := certs

# Generate CA certificate and private key
generate-ca:
	@echo "$(BLUE)Generating Certificate Authority...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating CA certificate...$(NC)"; \
		rm -f $(CERTS_DIR)/ca.crt $(CERTS_DIR)/ca.key $(CERTS_DIR)/ca.srl; \
	fi
	@if [ -f "$(CERTS_DIR)/ca.crt" ] && [ -f "$(CERTS_DIR)/ca.key" ]; then \
		echo "$(YELLOW)CA certificate already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		echo "$(GREEN)Creating CA private key (4096-bit RSA)...$(NC)"; \
		openssl genrsa -out $(CERTS_DIR)/ca.key 4096 2>/dev/null; \
		echo "$(GREEN)Creating self-signed CA certificate (10-year validity)...$(NC)"; \
		openssl req -new -x509 -days 3650 -key $(CERTS_DIR)/ca.key \
			-out $(CERTS_DIR)/ca.crt \
			-subj "/C=US/ST=Demo/L=Demo/O=elastauth-demo/CN=Demo-CA" 2>/dev/null; \
		echo "$(GREEN)✓ CA certificate generated successfully$(NC)"; \
	fi

# Generate Elasticsearch certificate
generate-elasticsearch-cert:
	@echo "$(BLUE)Generating Elasticsearch certificate...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating Elasticsearch certificate...$(NC)"; \
		rm -f $(CERTS_DIR)/elasticsearch.crt $(CERTS_DIR)/elasticsearch.key; \
	fi
	@if [ -f "$(CERTS_DIR)/elasticsearch.crt" ] && [ -f "$(CERTS_DIR)/elasticsearch.key" ]; then \
		echo "$(YELLOW)Elasticsearch certificate already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		echo "$(GREEN)Creating Elasticsearch private key (2048-bit RSA)...$(NC)"; \
		openssl genrsa -out $(CERTS_DIR)/elasticsearch.key 2048 2>/dev/null; \
		echo "$(GREEN)Creating Certificate Signing Request...$(NC)"; \
		openssl req -new -key $(CERTS_DIR)/elasticsearch.key \
			-out $(CERTS_DIR)/elasticsearch.csr \
			-subj "/C=US/ST=Demo/L=Demo/O=elastauth-demo/CN=elasticsearch" 2>/dev/null; \
		echo "$(GREEN)Creating SAN configuration...$(NC)"; \
		echo "subjectAltName = DNS:elasticsearch,DNS:localhost,IP:127.0.0.1" > $(CERTS_DIR)/elasticsearch.ext; \
		echo "$(GREEN)Signing certificate with CA (365-day validity)...$(NC)"; \
		openssl x509 -req -in $(CERTS_DIR)/elasticsearch.csr \
			-CA $(CERTS_DIR)/ca.crt -CAkey $(CERTS_DIR)/ca.key -CAcreateserial \
			-out $(CERTS_DIR)/elasticsearch.crt -days 365 \
			-extfile $(CERTS_DIR)/elasticsearch.ext 2>/dev/null; \
		rm -f $(CERTS_DIR)/elasticsearch.csr $(CERTS_DIR)/elasticsearch.ext; \
		echo "$(GREEN)✓ Elasticsearch certificate generated successfully$(NC)"; \
	fi

# Generate Authelia certificate
generate-authelia-cert:
	@echo "$(BLUE)Generating Authelia certificate...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating Authelia certificate...$(NC)"; \
		rm -f $(CERTS_DIR)/authelia.crt $(CERTS_DIR)/authelia.key; \
	fi
	@if [ -f "$(CERTS_DIR)/authelia.crt" ] && [ -f "$(CERTS_DIR)/authelia.key" ]; then \
		echo "$(YELLOW)Authelia certificate already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		echo "$(GREEN)Creating Authelia private key (2048-bit RSA)...$(NC)"; \
		openssl genrsa -out $(CERTS_DIR)/authelia.key 2048 2>/dev/null; \
		echo "$(GREEN)Creating Certificate Signing Request...$(NC)"; \
		openssl req -new -key $(CERTS_DIR)/authelia.key \
			-out $(CERTS_DIR)/authelia.csr \
			-subj "/C=US/ST=Demo/L=Demo/O=elastauth-demo/CN=authelia" 2>/dev/null; \
		echo "$(GREEN)Creating SAN configuration...$(NC)"; \
		echo "subjectAltName = DNS:authelia,DNS:localhost,IP:127.0.0.1" > $(CERTS_DIR)/authelia.ext; \
		echo "$(GREEN)Signing certificate with CA (365-day validity)...$(NC)"; \
		openssl x509 -req -in $(CERTS_DIR)/authelia.csr \
			-CA $(CERTS_DIR)/ca.crt -CAkey $(CERTS_DIR)/ca.key -CAcreateserial \
			-out $(CERTS_DIR)/authelia.crt -days 365 \
			-extfile $(CERTS_DIR)/authelia.ext 2>/dev/null; \
		rm -f $(CERTS_DIR)/authelia.csr $(CERTS_DIR)/authelia.ext; \
		echo "$(GREEN)✓ Authelia certificate generated successfully$(NC)"; \
	fi

# Force initialization - regenerate all certificates and configs
init-force: ## Force regenerate all certificates and configs
	@$(MAKE) FORCE=1 init
