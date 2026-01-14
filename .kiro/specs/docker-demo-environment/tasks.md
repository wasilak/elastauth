# Implementation Plan: Docker Demo Environment

## Overview

This implementation plan breaks down the creation of the Docker Compose demo environment into discrete, manageable tasks. Each task builds on previous work, ensuring incremental progress toward a fully functional demo environment. The implementation follows a logical sequence: directory structure → certificate generation → configuration files → Docker Compose → Makefile automation → documentation.

## Tasks

- [x] 1. Create directory structure and initial files
  - Create `deployment/example/` directory
  - Create subdirectories: `certs/`, `configs/authelia/`, `configs/elastauth/`
  - Create `.gitignore` to exclude generated files
  - Create `.env.example` with placeholder values
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [-] 2. Implement certificate generation in Makefile
  - [x] 2.1 Create Makefile with basic structure and help target
    - Define PHONY targets
    - Implement color output variables
    - Create help target with descriptions
    - _Requirements: 10.1, 10.6_

  - [x] 2.2 Implement CA certificate generation
    - Create function to generate CA private key (4096-bit RSA)
    - Create function to generate self-signed CA certificate (10-year validity)
    - Add proper subject DN for CA
    - _Requirements: 2.1, 2.4_

  - [x] 2.3 Implement Elasticsearch certificate generation
    - Create function to generate Elasticsearch private key (2048-bit RSA)
    - Create CSR with proper subject
    - Generate SAN configuration file with DNS:elasticsearch, DNS:localhost, IP:127.0.0.1
    - Sign certificate with CA (365-day validity)
    - _Requirements: 2.2, 2.4, 2.5_

  - [x] 2.4 Implement Authelia certificate generation
    - Create function to generate Authelia private key (2048-bit RSA)
    - Create CSR with proper subject
    - Generate SAN configuration file with DNS:authelia, DNS:localhost, IP:127.0.0.1
    - Sign certificate with CA (365-day validity)
    - _Requirements: 2.3, 2.4, 2.5_

  - [x] 2.5 Implement certificate storage and idempotency
    - Store all certificates in `certs/` directory
    - Add checks to skip generation if certificates exist
    - Add force flag to regenerate certificates
    - _Requirements: 2.6, 2.7_

  - [ ]* 2.6 Write property test for certificate validity period
    - **Property 1: Certificate Validity Period**
    - **Validates: Requirements 2.4**

  - [ ]* 2.7 Write property test for certificate SANs
    - **Property 2: Certificate Subject Alternative Names**
    - **Validates: Requirements 2.5**

- [x] 3. Implement configuration file generation in Makefile
  - [x] 3.1 Implement Authelia configuration generation
    - Generate `configs/authelia/configuration.yml` with heredoc
    - Include server, authentication_backend, session, storage, access_control, notifier sections
    - Generate random session secret
    - Configure Redis connection
    - _Requirements: 12.1, 12.4, 12.5, 13.3_

  - [x] 3.2 Implement Authelia users database generation
    - Generate `configs/authelia/users_database.yml`
    - Create admin user with argon2id hashed password
    - Use Docker to run authelia hash generation
    - Include displayname, email, and groups
    - _Requirements: 12.2, 13.2_

  - [x] 3.3 Implement elastauth configuration generation
    - Generate `configs/elastauth/config.yml` with heredoc
    - Configure Authelia provider with header mappings
    - Configure Elasticsearch connection with TLS
    - Configure Redis cache connection
    - Include logging configuration
    - _Requirements: 12.3, 12.6, 12.7_

  - [ ]* 3.4 Write property test for configuration completeness
    - **Property 4: Configuration File Completeness**
    - **Validates: Requirements 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7**

- [x] 4. Create Docker Compose configuration
  - [x] 4.1 Define networks and volumes
    - Create custom bridge network `elastauth-demo`
    - Define volumes for Elasticsearch and Redis data
    - _Requirements: 11.1_

  - [x] 4.2 Define Elasticsearch service
    - Use Elasticsearch 9.3.0 image
    - Configure single-node cluster with master and data roles
    - Enable xpack security with TLS
    - Mount certificates directory
    - Set memory limits (2GB heap)
    - Configure health check
    - Expose port 9200
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10, 3.11_

  - [x] 4.3 Define Redis service
    - Use redis:alpine image
    - Configure data volume
    - Configure health check
    - Expose port 6379
    - Set restart policy
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

  - [x] 4.4 Define Authelia service
    - Use latest Authelia image
    - Mount configuration and certificates
    - Configure environment variables
    - Set dependency on Redis with health check
    - Configure health check
    - Expose port 9091
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8, 4.9, 4.10_

  - [x] 4.5 Define elastauth service
    - Build from project Dockerfile
    - Mount configuration and certificates
    - Set dependencies on Redis, Elasticsearch, and Authelia with health checks
    - Expose port 3000
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

  - [ ]* 4.6 Write property test for service dependency order
    - **Property 5: Service Dependency Order**
    - **Validates: Requirements 11.3, 11.4**

- [x] 5. Implement Makefile service management targets
  - [x] 5.1 Implement init target
    - Create certificates directory
    - Call certificate generation functions
    - Call configuration generation functions
    - Display success message with next steps
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8_

  - [x] 5.2 Implement up target
    - Run docker-compose up in detached mode
    - Display service startup messages
    - _Requirements: 8.1_

  - [x] 5.3 Implement down target
    - Run docker-compose down
    - Display shutdown messages
    - _Requirements: 8.2_

  - [x] 5.4 Implement restart target
    - Call down target
    - Call up target
    - _Requirements: 8.3_

  - [x] 5.5 Implement logs target
    - Run docker-compose logs with timestamps
    - Support optional service name parameter
    - _Requirements: 8.4_

  - [x] 5.6 Implement status and ps targets
    - Run docker-compose ps
    - Display formatted service status
    - _Requirements: 8.5, 8.6_

- [-] 6. Implement Makefile cleanup targets
  - [x] 6.1 Implement clean target with confirmation
    - Prompt user for confirmation
    - Stop and remove containers
    - Remove volumes
    - Remove generated certificates
    - Remove generated configurations
    - Display cleanup messages
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6_

  - [x] 6.2 Implement clean-all target
    - Skip confirmation prompt
    - Call same cleanup operations as clean
    - _Requirements: 9.7_

- [x] 7. Implement Makefile information targets
  - [x] 7.1 Implement info target
    - Display connection URLs for all services
    - Display default credentials
    - Display example curl commands
    - Display authentication flow diagram
    - _Requirements: 10.2, 10.3, 10.4, 10.5_

  - [x] 7.2 Enhance help target
    - List all targets with descriptions
    - Group targets by category (init, service management, cleanup, info)
    - Display usage examples
    - _Requirements: 10.1_

- [x] 8. Implement Makefile testing target
  - [x] 8.1 Implement test target
    - Check Elasticsearch health endpoint with curl
    - Check Authelia health endpoint with curl
    - Check Redis connectivity with redis-cli
    - Check elastauth health endpoint with curl
    - Display pass/fail status for each service
    - Provide example curl commands for manual testing
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7_

- [ ] 9. Create comprehensive documentation
  - [ ] 9.1 Create README.md
    - Write introduction explaining demo purpose
    - Document quick start instructions
    - Document all Makefile targets
    - Explain authentication flow with diagram
    - Document all service ports and URLs
    - Include troubleshooting section
    - Document how to customize configuration
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7, 15.8, 15.9_

  - [ ] 9.2 Create .env.example file
    - Document all environment variables
    - Provide example values
    - Include security warnings for production
    - _Requirements: 13.5, 13.6_

- [ ] 10. Checkpoint - Verify complete demo environment
  - Run `make init` from clean state
  - Verify all certificates generated correctly
  - Verify all configuration files created
  - Run `make up` and verify all services start
  - Run `make test` and verify all health checks pass
  - Test authentication flow manually with curl
  - Run `make info` and verify output is correct
  - Run `make clean` and verify complete cleanup
  - Review README for completeness
  - Ensure all tests pass, ask the user if questions arise

- [ ]* 11. Write property test for bootstrap idempotency
  - **Property 3: Bootstrap Idempotency**
  - **Validates: Requirements 2.7**

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoint ensures incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- The implementation uses shell scripting in Makefile for simplicity
- OpenSSL is used for certificate generation (available on most systems)
- Docker and Docker Compose are prerequisites
