---
inclusion: always
---

# Docker Compose Modern Syntax

## Use `docker compose` Not `docker-compose`

### Modern Docker CLI Plugin

Docker Compose V2 is now integrated into the Docker CLI as a plugin. Always use the modern syntax:

✅ **Correct**: `docker compose`
❌ **Deprecated**: `docker-compose`

### Command Examples

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# View logs
docker compose logs -f

# Check status
docker compose ps

# Execute commands
docker compose exec service-name command

# Build images
docker compose build

# Pull images
docker compose pull
```

### Why This Matters

1. **Modern Standard**: Docker Compose V2 is the current standard
2. **Better Integration**: Integrated with Docker CLI
3. **Consistent Experience**: Same CLI patterns as other Docker commands
4. **Future-Proof**: V1 (`docker-compose`) is deprecated
5. **Better Performance**: V2 is written in Go, faster than V1 (Python)

### Documentation and Examples

When writing documentation or examples:

- Always use `docker compose` in code blocks
- Update any legacy `docker-compose` references
- Mention V2 in prerequisites if relevant

### Example Documentation Pattern

```markdown
## Prerequisites

- Docker Engine 20.10 or later
- Docker Compose V2 (integrated with Docker CLI)

## Quick Start

```bash
# Start the stack
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

## Cleanup

```bash
docker compose down -v
```
```

### Checking Docker Compose Version

```bash
# Check if V2 is installed
docker compose version

# Should output something like:
# Docker Compose version v2.x.x
```

### Migration from V1 to V2

If you encounter `docker-compose` in existing code:

1. Replace `docker-compose` with `docker compose` (note the space)
2. Test all commands work correctly
3. Update documentation
4. No changes needed to `docker-compose.yml` files

### File Naming

Both V1 and V2 use the same file names:
- `docker-compose.yml` (primary file name)
- `docker-compose.yaml` (alternative)
- `compose.yml` (V2 also supports this shorter name)

### elastauth-Specific Usage

For elastauth deployment examples:

```bash
# Development
docker compose -f docker-compose.yml up -d

# Production
docker compose -f docker-compose.prod.yml up -d

# With environment file
docker compose --env-file .env.production up -d
```

## Remember

Always use `docker compose` (two words, space-separated) in:
- Documentation
- README files
- Deployment guides
- Scripts
- Examples
- Comments

This ensures elastauth documentation stays current with modern Docker practices.
