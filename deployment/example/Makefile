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
	@echo "$(BLUE)╔════════════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(BLUE)║        elastauth Demo Environment - Available Commands        ║$(NC)"
	@echo "$(BLUE)╚════════════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(GREEN)  Initialization$(NC)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(BLUE)make init$(NC)          Bootstrap the demo environment"
	@echo "                     - Generate CA and service certificates"
	@echo "                     - Create Authelia configuration and users"
	@echo "                     - Create elastauth configuration"
	@echo ""
	@echo "  $(BLUE)make init-force$(NC)    Force regenerate all certificates and configs"
	@echo "                     - Useful when certificates expire"
	@echo "                     - Overwrites existing files"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(GREEN)  Service Management$(NC)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(BLUE)make up$(NC)            Start all services in detached mode"
	@echo "                     - Starts: Elasticsearch, Redis, Authelia, elastauth"
	@echo "                     - Services run in background"
	@echo ""
	@echo "  $(BLUE)make down$(NC)          Stop and remove all containers"
	@echo "                     - Preserves volumes (data persists)"
	@echo ""
	@echo "  $(BLUE)make restart$(NC)       Restart all services"
	@echo "                     - Equivalent to: make down && make up"
	@echo ""
	@echo "  $(BLUE)make ps$(NC)            Show running containers"
	@echo "                     - Displays container status and ports"
	@echo ""
	@echo "  $(BLUE)make status$(NC)        Display formatted service status"
	@echo "                     - Alias for 'make ps'"
	@echo ""
	@echo "  $(BLUE)make logs$(NC)          Display logs from all services"
	@echo "                     - Shows timestamped logs"
	@echo "                     - Use: make logs SERVICE=<name> for specific service"
	@echo ""
	@echo "  $(BLUE)make logs-follow$(NC)   Follow logs in real-time"
	@echo "                     - Continuously streams logs"
	@echo "                     - Press Ctrl+C to exit"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(GREEN)  Testing and Information$(NC)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(BLUE)make test$(NC)          Run health checks on all services"
	@echo "                     - Validates Elasticsearch, Authelia, Redis, elastauth"
	@echo "                     - Provides pass/fail status"
	@echo ""
	@echo "  $(BLUE)make info$(NC)          Display connection information and credentials"
	@echo "                     - Service URLs and ports"
	@echo "                     - Default credentials"
	@echo "                     - Example curl commands"
	@echo "                     - Authentication flow diagram"
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(GREEN)  Cleanup$(NC)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(BLUE)make clean$(NC)         Remove containers, volumes, certs, and configs"
	@echo "                     - Prompts for confirmation"
	@echo "                     - Removes all generated files"
	@echo "                     - Deletes Docker volumes (data loss!)"
	@echo ""
	@echo "  $(BLUE)make clean-all$(NC)     Force clean without confirmation"
	@echo "                     - Same as 'make clean' but no prompt"
	@echo "                     - Use with caution!"
	@echo ""
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(YELLOW)  Quick Start Guide$(NC)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(YELLOW)1.$(NC) $(BLUE)make init$(NC)       Initialize the environment"
	@echo "  $(YELLOW)2.$(NC) $(BLUE)make up$(NC)         Start all services"
	@echo "  $(YELLOW)3.$(NC) $(BLUE)make test$(NC)       Verify services are healthy"
	@echo "  $(YELLOW)4.$(NC) $(BLUE)make info$(NC)       Get connection details"
	@echo ""
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(YELLOW)  Usage Examples$(NC)"
	@echo "$(YELLOW)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  $(YELLOW)View logs for specific service:$(NC)"
	@echo "    make logs SERVICE=elasticsearch"
	@echo "    make logs SERVICE=authelia"
	@echo "    make logs SERVICE=elastauth"
	@echo ""
	@echo "  $(YELLOW)Follow logs for specific service:$(NC)"
	@echo "    make logs-follow SERVICE=elastauth"
	@echo ""
	@echo "  $(YELLOW)Force regenerate certificates:$(NC)"
	@echo "    make init-force"
	@echo "    make FORCE=1 init"
	@echo ""
	@echo "  $(YELLOW)Complete cleanup and restart:$(NC)"
	@echo "    make clean-all && make init && make up"
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(BLUE)  Service Ports$(NC)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  Elasticsearch:  https://localhost:9200"
	@echo "  Authelia:       http://localhost:9091"
	@echo "  elastauth:      http://localhost:3000"
	@echo "  Redis:          localhost:6379"
	@echo ""
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(BLUE)  Default Credentials (Demo Only!)$(NC)"
	@echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "  Elasticsearch: elastic / demo-password-change-me"
	@echo "  Authelia:      admin / demo-password"
	@echo ""
	@echo "$(RED)  ⚠️  WARNING: Change these passwords for production use!$(NC)"
	@echo ""
	@echo "$(GREEN)For detailed connection info and examples: $(BLUE)make info$(NC)"
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

# Start all services in detached mode
up: ## Start all services in detached mode
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)Starting elastauth Demo Services$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	@docker-compose up -d
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ Services Started!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Services are starting in the background...$(NC)"
	@echo "  - Elasticsearch: http://localhost:9200"
	@echo "  - Authelia: http://localhost:9091"
	@echo "  - elastauth: http://localhost:3000"
	@echo "  - Redis: localhost:6379"
	@echo ""
	@echo "$(YELLOW)Next steps:$(NC)"
	@echo "  1. Wait for services to become healthy (30-60 seconds)"
	@echo "  2. Check service status: $(BLUE)make ps$(NC)"
	@echo "  3. View logs: $(BLUE)make logs$(NC)"
	@echo "  4. Run health checks: $(BLUE)make test$(NC)"
	@echo ""

# Stop and remove all containers
down: ## Stop and remove all containers
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)Stopping elastauth Demo Services$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Stopping and removing containers...$(NC)"
	@docker-compose down
	@echo ""
	@echo "$(GREEN)✓ Services stopped and containers removed$(NC)"
	@echo ""

# Restart all services
restart: ## Restart all services
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)Restarting elastauth Demo Services$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@$(MAKE) down
	@echo ""
	@$(MAKE) up

# Display logs from all services (or specific service if SERVICE is set)
logs: ## Display logs from all services with timestamps
	@if [ -n "$(SERVICE)" ]; then \
		echo "$(BLUE)Displaying logs for service: $(SERVICE)$(NC)"; \
		docker-compose logs --timestamps $(SERVICE); \
	else \
		echo "$(BLUE)Displaying logs from all services$(NC)"; \
		docker-compose logs --timestamps; \
	fi
	@echo ""
	@echo "$(YELLOW)Tip: Use 'make logs SERVICE=<service-name>' to view specific service logs$(NC)"
	@echo "$(YELLOW)Tip: Use 'make logs-follow' to follow logs in real-time$(NC)"

# Follow logs in real-time
logs-follow: ## Follow logs in real-time
	@if [ -n "$(SERVICE)" ]; then \
		echo "$(BLUE)Following logs for service: $(SERVICE)$(NC)"; \
		docker-compose logs --timestamps --follow $(SERVICE); \
	else \
		echo "$(BLUE)Following logs from all services$(NC)"; \
		docker-compose logs --timestamps --follow; \
	fi

# Show running containers
ps: ## Show running containers
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)elastauth Demo Services Status$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@docker-compose ps
	@echo ""

# Alias for ps target
status: ps ## Display formatted service status (alias for ps)

# Test target - run health checks on all services
test: ## Run health checks on all services
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)elastauth Demo Environment - Health Checks$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)Running health checks on all services...$(NC)"
	@echo ""
	@PASS_COUNT=0; \
	FAIL_COUNT=0; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo "$(BLUE)  Service Health Checks$(NC)"; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo ""; \
	echo "$(YELLOW)1. Checking Elasticsearch...$(NC)"; \
	if curl -k -s -u elastic:demo-password-change-me https://localhost:9200/_cluster/health > /dev/null 2>&1; then \
		echo "   $(GREEN)✓ PASS$(NC) - Elasticsearch is healthy"; \
		echo "   Endpoint: https://localhost:9200/_cluster/health"; \
		PASS_COUNT=$$((PASS_COUNT + 1)); \
	else \
		echo "   $(RED)✗ FAIL$(NC) - Elasticsearch is not responding"; \
		echo "   Endpoint: https://localhost:9200/_cluster/health"; \
		echo "   $(YELLOW)Tip: Check if service is running with 'make ps'$(NC)"; \
		FAIL_COUNT=$$((FAIL_COUNT + 1)); \
	fi; \
	echo ""; \
	echo "$(YELLOW)2. Checking Authelia...$(NC)"; \
	if curl -s http://localhost:9091/api/health > /dev/null 2>&1; then \
		echo "   $(GREEN)✓ PASS$(NC) - Authelia is healthy"; \
		echo "   Endpoint: http://localhost:9091/api/health"; \
		PASS_COUNT=$$((PASS_COUNT + 1)); \
	else \
		echo "   $(RED)✗ FAIL$(NC) - Authelia is not responding"; \
		echo "   Endpoint: http://localhost:9091/api/health"; \
		echo "   $(YELLOW)Tip: Check if service is running with 'make ps'$(NC)"; \
		FAIL_COUNT=$$((FAIL_COUNT + 1)); \
	fi; \
	echo ""; \
	echo "$(YELLOW)3. Checking Redis...$(NC)"; \
	if docker exec redis-demo redis-cli ping > /dev/null 2>&1; then \
		echo "   $(GREEN)✓ PASS$(NC) - Redis is healthy"; \
		echo "   Connection: localhost:6379"; \
		PASS_COUNT=$$((PASS_COUNT + 1)); \
	else \
		echo "   $(RED)✗ FAIL$(NC) - Redis is not responding"; \
		echo "   Connection: localhost:6379"; \
		echo "   $(YELLOW)Tip: Check if service is running with 'make ps'$(NC)"; \
		FAIL_COUNT=$$((FAIL_COUNT + 1)); \
	fi; \
	echo ""; \
	echo "$(YELLOW)4. Checking elastauth...$(NC)"; \
	if curl -s http://localhost:3000/health > /dev/null 2>&1; then \
		echo "   $(GREEN)✓ PASS$(NC) - elastauth is healthy"; \
		echo "   Endpoint: http://localhost:3000/health"; \
		PASS_COUNT=$$((PASS_COUNT + 1)); \
	else \
		echo "   $(RED)✗ FAIL$(NC) - elastauth is not responding"; \
		echo "   Endpoint: http://localhost:3000/health"; \
		echo "   $(YELLOW)Tip: Check if service is running with 'make ps'$(NC)"; \
		FAIL_COUNT=$$((FAIL_COUNT + 1)); \
	fi; \
	echo ""; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo "$(BLUE)  Test Results$(NC)"; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo ""; \
	echo "  $(GREEN)Passed:$(NC) $$PASS_COUNT / 4"; \
	echo "  $(RED)Failed:$(NC) $$FAIL_COUNT / 4"; \
	echo ""; \
	if [ $$FAIL_COUNT -eq 0 ]; then \
		echo "$(GREEN)✓ All health checks passed!$(NC)"; \
		echo ""; \
		echo "$(YELLOW)The demo environment is ready to use.$(NC)"; \
	else \
		echo "$(RED)✗ Some health checks failed.$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Troubleshooting steps:$(NC)"; \
		echo "  1. Check service status: $(BLUE)make ps$(NC)"; \
		echo "  2. View service logs: $(BLUE)make logs$(NC)"; \
		echo "  3. Wait a bit longer for services to start (they may still be initializing)"; \
		echo "  4. Restart services: $(BLUE)make restart$(NC)"; \
	fi; \
	echo ""; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo "$(BLUE)  Manual Testing Examples$(NC)"; \
	echo "$(BLUE)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"; \
	echo ""; \
	echo "$(YELLOW)Test Elasticsearch directly:$(NC)"; \
	echo "  curl -k -u elastic:demo-password-change-me https://localhost:9200/_cluster/health"; \
	echo ""; \
	echo "$(YELLOW)Test Authelia health:$(NC)"; \
	echo "  curl http://localhost:9091/api/health"; \
	echo ""; \
	echo "$(YELLOW)Test Redis connectivity:$(NC)"; \
	echo "  docker exec redis-demo redis-cli ping"; \
	echo ""; \
	echo "$(YELLOW)Test elastauth health:$(NC)"; \
	echo "  curl http://localhost:3000/health"; \
	echo ""; \
	echo "$(YELLOW)Test authentication flow through elastauth:$(NC)"; \
	echo "  curl -H \"Remote-User: admin\" \\"; \
	echo "       -H \"Remote-Groups: admins\" \\"; \
	echo "       -H \"Remote-Email: admin@example.com\" \\"; \
	echo "       -H \"Remote-Name: Admin User\" \\"; \
	echo "       http://localhost:3000/"; \
	echo ""; \
	echo "$(YELLOW)Query Elasticsearch through elastauth:$(NC)"; \
	echo "  curl -H \"Remote-User: admin\" \\"; \
	echo "       -H \"Remote-Groups: admins\" \\"; \
	echo "       http://localhost:3000/_cluster/health"; \
	echo ""; \
	echo "$(YELLOW)For more information: $(BLUE)make info$(NC)"; \
	echo ""

# Info target - display connection information and credentials
info: ## Display connection information and credentials
	@echo "$(BLUE)========================================$(NC)"
	@echo "$(BLUE)elastauth Demo Environment - Connection Info$(NC)"
	@echo "$(BLUE)========================================$(NC)"
	@echo ""
	@echo "$(GREEN)Service URLs:$(NC)"
	@echo "  Elasticsearch:  https://localhost:9200"
	@echo "  Authelia:       http://localhost:9091"
	@echo "  elastauth:      http://localhost:3000"
	@echo "  Redis:          localhost:6379"
	@echo ""
	@echo "$(GREEN)Default Credentials:$(NC)"
	@echo "  $(YELLOW)Elasticsearch:$(NC)"
	@echo "    Username: elastic"
	@echo "    Password: demo-password-change-me"
	@echo ""
	@echo "  $(YELLOW)Authelia:$(NC)"
	@echo "    Username: admin"
	@echo "    Password: demo-password"
	@echo "    Email:    admin@example.com"
	@echo ""
	@echo "$(RED)⚠️  WARNING: These are demo credentials only!$(NC)"
	@echo "$(RED)   Change these passwords for production use!$(NC)"
	@echo ""
	@echo "$(GREEN)Example curl Commands:$(NC)"
	@echo ""
	@echo "  $(YELLOW)1. Check Elasticsearch health (direct):$(NC)"
	@echo "     curl -k -u elastic:demo-password-change-me https://localhost:9200/_cluster/health"
	@echo ""
	@echo "  $(YELLOW)2. Check Authelia health:$(NC)"
	@echo "     curl http://localhost:9091/api/health"
	@echo ""
	@echo "  $(YELLOW)3. Check elastauth health:$(NC)"
	@echo "     curl http://localhost:3000/health"
	@echo ""
	@echo "  $(YELLOW)4. Test authentication flow through elastauth:$(NC)"
	@echo "     curl -H \"Remote-User: admin\" \\"
	@echo "          -H \"Remote-Groups: admins\" \\"
	@echo "          -H \"Remote-Email: admin@example.com\" \\"
	@echo "          -H \"Remote-Name: Admin User\" \\"
	@echo "          http://localhost:3000/"
	@echo ""
	@echo "  $(YELLOW)5. Query Elasticsearch through elastauth:$(NC)"
	@echo "     curl -H \"Remote-User: admin\" \\"
	@echo "          -H \"Remote-Groups: admins\" \\"
	@echo "          http://localhost:3000/_cluster/health"
	@echo ""
	@echo "$(GREEN)Authentication Flow:$(NC)"
	@echo ""
	@echo "  ┌──────────┐"
	@echo "  │  Client  │"
	@echo "  └────┬─────┘"
	@echo "       │ 1. HTTP Request"
	@echo "       ▼"
	@echo "  ┌─────────────────┐"
	@echo "  │   Authelia      │ :9091"
	@echo "  │  (Auth Server)  │"
	@echo "  │                 │"
	@echo "  │ - Validates     │"
	@echo "  │   credentials   │"
	@echo "  │ - Sets headers  │"
	@echo "  └────┬────────────┘"
	@echo "       │ 2. Request + Auth Headers"
	@echo "       │    (Remote-User, Remote-Groups, etc.)"
	@echo "       ▼"
	@echo "  ┌─────────────────┐"
	@echo "  │   elastauth     │ :3000"
	@echo "  │  (Auth Proxy)   │"
	@echo "  │                 │"
	@echo "  │ - Validates     │"
	@echo "  │   headers       │"
	@echo "  │ - Maps to ES    │"
	@echo "  │   user          │"
	@echo "  └────┬────────────┘"
	@echo "       │ 3. Proxied Request + ES Auth"
	@echo "       ▼"
	@echo "  ┌─────────────────┐"
	@echo "  │ Elasticsearch   │ :9200"
	@echo "  │  (Backend)      │"
	@echo "  │                 │"
	@echo "  │ - Processes     │"
	@echo "  │   request       │"
	@echo "  │ - Returns data  │"
	@echo "  └─────────────────┘"
	@echo ""
	@echo "       ┌──────────┐"
	@echo "       │  Redis   │ :6379"
	@echo "       │ (Cache)  │"
	@echo "       └──────────┘"
	@echo "            ▲"
	@echo "            │ Session Storage"
	@echo "            │"
	@echo "       ┌────┴────┐"
	@echo "       │         │"
	@echo "    Authelia  elastauth"
	@echo ""
	@echo "$(YELLOW)For more information:$(NC)"
	@echo "  - View logs: $(BLUE)make logs$(NC)"
	@echo "  - Check status: $(BLUE)make ps$(NC)"
	@echo "  - Run health checks: $(BLUE)make test$(NC)"
	@echo ""

# Clean target with confirmation - removes containers, volumes, certs, and configs
clean: ## Remove containers, volumes, certs, and configs (with confirmation)
	@echo "$(RED)========================================$(NC)"
	@echo "$(RED)WARNING: Cleanup Operation$(NC)"
	@echo "$(RED)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)This will:$(NC)"
	@echo "  - Stop and remove all containers"
	@echo "  - Remove all Docker volumes (data will be lost)"
	@echo "  - Remove generated certificates in $(CERTS_DIR)/"
	@echo "  - Remove generated configurations in $(CONFIGS_DIR)/"
	@echo ""
	@echo "$(RED)This operation cannot be undone!$(NC)"
	@echo ""
	@read -p "Are you sure you want to continue? [y/N] " -n 1 -r; \
	echo ""; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		echo "$(BLUE)Starting cleanup...$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Step 1: Stopping and removing containers...$(NC)"; \
		docker-compose down -v 2>/dev/null || true; \
		echo "$(GREEN)✓ Containers and volumes removed$(NC)"; \
		echo ""; \
		echo "$(YELLOW)Step 2: Removing generated certificates...$(NC)"; \
		if [ -d "$(CERTS_DIR)" ]; then \
			rm -rf $(CERTS_DIR)/*; \
			echo "$(GREEN)✓ Certificates removed$(NC)"; \
		else \
			echo "$(YELLOW)No certificates directory found$(NC)"; \
		fi; \
		echo ""; \
		echo "$(YELLOW)Step 3: Removing generated configurations...$(NC)"; \
		if [ -d "$(CONFIGS_DIR)" ]; then \
			rm -rf $(CONFIGS_DIR)/*; \
			echo "$(GREEN)✓ Configurations removed$(NC)"; \
		else \
			echo "$(YELLOW)No configurations directory found$(NC)"; \
		fi; \
		echo ""; \
		echo "$(GREEN)========================================$(NC)"; \
		echo "$(GREEN)✓ Cleanup Complete!$(NC)"; \
		echo "$(GREEN)========================================$(NC)"; \
		echo ""; \
		echo "$(YELLOW)To start fresh:$(NC)"; \
		echo "  1. Run: $(BLUE)make init$(NC)"; \
		echo "  2. Run: $(BLUE)make up$(NC)"; \
		echo ""; \
	else \
		echo ""; \
		echo "$(YELLOW)Cleanup cancelled$(NC)"; \
		echo ""; \
	fi

# Clean-all target - force clean without confirmation
clean-all: ## Force clean without confirmation
	@echo "$(RED)========================================$(NC)"
	@echo "$(RED)Force Cleanup (No Confirmation)$(NC)"
	@echo "$(RED)========================================$(NC)"
	@echo ""
	@echo "$(BLUE)Starting cleanup...$(NC)"
	@echo ""
	@echo "$(YELLOW)Step 1: Stopping and removing containers...$(NC)"
	@docker-compose down -v 2>/dev/null || true
	@echo "$(GREEN)✓ Containers and volumes removed$(NC)"
	@echo ""
	@echo "$(YELLOW)Step 2: Removing generated certificates...$(NC)"
	@if [ -d "$(CERTS_DIR)" ]; then \
		rm -rf $(CERTS_DIR)/*; \
		echo "$(GREEN)✓ Certificates removed$(NC)"; \
	else \
		echo "$(YELLOW)No certificates directory found$(NC)"; \
	fi
	@echo ""
	@echo "$(YELLOW)Step 3: Removing generated configurations...$(NC)"
	@if [ -d "$(CONFIGS_DIR)" ]; then \
		rm -rf $(CONFIGS_DIR)/*; \
		echo "$(GREEN)✓ Configurations removed$(NC)"; \
	else \
		echo "$(YELLOW)No configurations directory found$(NC)"; \
	fi
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ Cleanup Complete!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(YELLOW)To start fresh:$(NC)"
	@echo "  1. Run: $(BLUE)make init$(NC)"
	@echo "  2. Run: $(BLUE)make up$(NC)"
	@echo ""
