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

# Configuration generation functions
CONFIGS_DIR := configs

# Generate Authelia configuration file
generate-authelia-config:
	@echo "$(BLUE)Generating Authelia configuration...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating Authelia configuration...$(NC)"; \
		rm -f $(CONFIGS_DIR)/authelia/configuration.yml; \
	fi
	@if [ -f "$(CONFIGS_DIR)/authelia/configuration.yml" ]; then \
		echo "$(YELLOW)Authelia configuration already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		mkdir -p $(CONFIGS_DIR)/authelia; \
		echo "$(GREEN)Generating random session secret...$(NC)"; \
		SESSION_SECRET=$$(openssl rand -base64 32); \
		echo "$(GREEN)Creating configuration.yml...$(NC)"; \
		printf '%s\n' \
			'---' \
			'# Authelia Configuration for elastauth Demo' \
			'' \
			'server:' \
			'  address: "tcp://0.0.0.0:9091"' \
			'' \
			'log:' \
			'  level: "info"' \
			'' \
			'# File-based authentication backend' \
			'authentication_backend:' \
			'  file:' \
			'    path: "/config/users_database.yml"' \
			'    password:' \
			'      algorithm: "argon2id"' \
			'      iterations: 3' \
			'      key_length: 32' \
			'      salt_length: 16' \
			'      memory: 65536' \
			'      parallelism: 4' \
			'' \
			'# Session configuration with Redis storage' \
			'session:' \
			"  secret: \"$$SESSION_SECRET\"" \
			'  cookies:' \
			'    - domain: "localhost"' \
			'      authelia_url: "http://localhost:9091"' \
			'  redis:' \
			'    host: "redis"' \
			'    port: 6379' \
			'' \
			'# Local SQLite storage for Authelia data' \
			'storage:' \
			'  local:' \
			'    path: "/config/db.sqlite3"' \
			'' \
			'# Access control - require one factor authentication' \
			'access_control:' \
			'  default_policy: "one_factor"' \
			'' \
			'# Filesystem notifier for demo purposes' \
			'notifier:' \
			'  filesystem:' \
			'    filename: "/config/notification.txt"' \
			> $(CONFIGS_DIR)/authelia/configuration.yml; \
		echo "$(GREEN)✓ Authelia configuration generated successfully$(NC)"; \
	fi

# Generate Authelia users database with hashed password
generate-authelia-users:
	@echo "$(BLUE)Generating Authelia users database...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating Authelia users database...$(NC)"; \
		rm -f $(CONFIGS_DIR)/authelia/users_database.yml; \
	fi
	@if [ -f "$(CONFIGS_DIR)/authelia/users_database.yml" ]; then \
		echo "$(YELLOW)Authelia users database already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		mkdir -p $(CONFIGS_DIR)/authelia; \
		echo "$(GREEN)Generating argon2id password hash for admin user...$(NC)"; \
		echo "$(YELLOW)Using Docker to generate password hash (password: demo-password)...$(NC)"; \
		PASSWORD_HASH=$$(docker run --rm authelia/authelia:latest authelia crypto hash generate argon2 --password 'demo-password' 2>&1 | grep 'Digest:' | sed 's/Digest: //'); \
		if [ -z "$$PASSWORD_HASH" ]; then \
			echo "$(RED)✗ Failed to generate password hash$(NC)"; \
			echo "$(RED)Please ensure Docker is running and authelia/authelia:latest image is available$(NC)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)Creating users_database.yml...$(NC)"; \
		printf '%s\n' \
			'---' \
			'# Authelia Users Database for elastauth Demo' \
			'# Default credentials:' \
			'#   Username: admin' \
			'#   Password: demo-password' \
			'# WARNING: These are demo credentials only. Change for production use!' \
			'' \
			'users:' \
			'  admin:' \
			'    displayname: "Admin User"' \
			"    password: \"$$PASSWORD_HASH\"" \
			'    email: "admin@example.com"' \
			'    groups:' \
			'      - admins' \
			> $(CONFIGS_DIR)/authelia/users_database.yml; \
		echo "$(GREEN)✓ Authelia users database generated successfully$(NC)"; \
		echo "$(YELLOW)Default credentials: admin / demo-password$(NC)"; \
	fi

# Generate elastauth configuration file
generate-elastauth-config:
	@echo "$(BLUE)Generating elastauth configuration...$(NC)"
	@if [ "$(FORCE)" = "1" ]; then \
		echo "$(YELLOW)Force flag set, regenerating elastauth configuration...$(NC)"; \
		rm -f $(CONFIGS_DIR)/elastauth/config.yml; \
	fi
	@if [ -f "$(CONFIGS_DIR)/elastauth/config.yml" ]; then \
		echo "$(YELLOW)elastauth configuration already exists, skipping generation$(NC)"; \
		echo "$(YELLOW)Use 'make init-force' or 'make FORCE=1 init' to regenerate$(NC)"; \
	else \
		mkdir -p $(CONFIGS_DIR)/elastauth; \
		echo "$(GREEN)Creating config.yml...$(NC)"; \
		printf '%s\n' \
			'# elastauth Configuration for Demo Environment' \
			'# This configuration demonstrates elastauth with Authelia authentication' \
			'' \
			'# Server configuration' \
			'server:' \
			'  port: 3000' \
			'  host: "0.0.0.0"' \
			'' \
			'# Authentication provider configuration' \
			'auth_provider:' \
			'  type: "authelia"' \
			'  authelia:' \
			'    # Header mappings from Authelia' \
			'    user_header: "Remote-User"' \
			'    groups_header: "Remote-Groups"' \
			'    email_header: "Remote-Email"' \
			'    name_header: "Remote-Name"' \
			'' \
			'# Elasticsearch backend configuration' \
			'elasticsearch:' \
			'  url: "https://elasticsearch:9200"' \
			'  username: "elastic"' \
			'  password: "demo-password-change-me"' \
			'  tls:' \
			'    enabled: true' \
			'    ca_cert: "/certs/ca.crt"' \
			'    skip_verify: false' \
			'' \
			'# Redis cache configuration' \
			'cache:' \
			'  type: "redis"' \
			'  redis:' \
			'    address: "redis:6379"' \
			'    db: 0' \
			'' \
			'# Logging configuration' \
			'logging:' \
			'  level: "info"' \
			'  format: "json"' \
			> $(CONFIGS_DIR)/elastauth/config.yml; \
		echo "$(GREEN)✓ elastauth configuration generated successfully$(NC)"; \
	fi

# Initialize the demo environment
init: ## Bootstrap the demo environment (generate certs and configs)
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)Initializing elastauth Demo Environment$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@mkdir -p $(CERTS_DIR) $(CONFIGS_DIR)/authelia $(CONFIGS_DIR)/elastauth
	@echo "$(GREEN)Step 1: Certificate Generation$(NC)"
	@$(MAKE) generate-ca
	@$(MAKE) generate-elasticsearch-cert
	@$(MAKE) generate-authelia-cert
	@echo ""
	@echo "$(GREEN)Step 2: Configuration Generation$(NC)"
	@$(MAKE) generate-authelia-config
	@$(MAKE) generate-authelia-users
	@$(MAKE) generate-elastauth-config
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ Initialization Complete!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Review generated configurations in $(CONFIGS_DIR)/"
	@echo "  2. Start services with: $(BLUE)make up$(NC)"
	@echo "  3. Check service health: $(BLUE)make test$(NC)"
	@echo "  4. View connection info: $(BLUE)make info$(NC)"
	@echo ""
	@echo "$(YELLOW)Default credentials:$(NC)"
	@echo "  Elasticsearch: elastic / demo-password-change-me"
	@echo "  Authelia: admin / demo-password"
	@echo ""

# Force initialization - regenerate all certificates and configs
init-force: ## Force regenerate all certificates and configs
	@$(MAKE) FORCE=1 init
