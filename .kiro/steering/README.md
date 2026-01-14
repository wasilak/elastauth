# Steering Rules

This directory contains steering rules that guide development of this project.

## Active Steering Rules

1. **00-serena-bootstrap.md** - Serena MCP tool initialization (always)
2. **01-task-completion-requirements.md** - Task completion criteria (always)
3. **02-naming-conventions.md** - Brand-agnostic naming and Go conventions (always)
4. **03-go-best-practices.md** - Go language best practices (always)
5. **04-phase-gate-discipline.md** - Incremental development discipline (always)
6. **05-context7-research-tools.md** - Context7 MCP tools for research and development (always)
7. **06-starlight-documentation.md** - Starlight documentation best practices (always)
8. **07-docker-compose-modern-syntax.md** - Docker Compose V2 modern syntax (always)
9. **08-deployment-examples-documentation.md** - Minimal example READMEs with Starlight docs (always)

## Rule Priority

All rules marked `inclusion: always` are active for every conversation.

## Key Principles

- **Incremental Development**: Build in phases with mandatory gates
- **Brand Agnostic Code**: No project name in function/struct names
- **Go Idioms**: Follow standard Go conventions
- **Working Code**: Functional at every step, not theoretical architecture
- **Simple First**: Avoid premature optimization and abstraction
- **Research-Driven**: Use Context7 MCP tools for current best practices
- **Starlight Documentation**: Use proper .mdx format with components and Mermaid diagrams
- **Modern Docker**: Use `docker compose` (V2) not `docker-compose` (V1)
- **Minimal Examples**: Deployment example READMEs are minimal, Starlight docs are comprehensive

## For AI Assistants

Read these rules at the start of each conversation to understand:
- How to use Serena tools effectively
- What constitutes task completion
- Naming conventions to follow
- Go best practices to apply
- Phase gate discipline to maintain
- When and how to use Context7 research tools
- How to create proper Starlight documentation with components and Mermaid diagrams
- Modern Docker Compose V2 syntax
- How to structure deployment examples with minimal READMEs

## For Developers

These rules ensure:
- Consistent code style
- Proper task completion
- Incremental progress
- Quality gates at each phase
- Maintainable, idiomatic Go code
- Current best practices through research
- Professional Starlight documentation with proper components
- Modern Docker Compose V2 syntax in all examples
- Single source of truth with minimal example READMEs
