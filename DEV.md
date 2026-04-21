# Developer Guide

Complete guide for developers contributing to Culturae.

## Prerequisites

- **Go**
- **Node.js**
- **Docker & Docker Compose**
- **golangci-lint**

---

For the api you have the [OpenAPI file](./backend/docs/admin-openapi.yaml) and for the wesocket section you have the [AsyncAPI file](./backend/docs/asyncapi.yaml).

## Infrastructure Setup

The first step is to start the required services: PostgreSQL, Redis, and MinIO.

```bash
cd backend

# Copy environment configuration
cp .env.example .env

# Start db
docker compose up -d
```

See `.env.example` for all available options.

---

## Using Task

[Task](https://taskfile.dev/) automates common development workflows. Run from the repository root:

```bash
task --list  # See all available tasks

# Infrastructure
task dev:infra      # Start postgres, redis, minio
task docker:down    # Stop all services
task docker:logs    # Follow backend logs

# Backend
task dev            # Start backend (requires infra running)
task test           # Run tests + linter
task build:backend  # Compile Go binary

# Frontend
task dev:dashboard  # Start admin dashboard (hot reload)
task build:dashboard

# Full build
task build          # Dashboard → Embed → Go binary
```

---

## Backend Development

Culturae uses **dependency injection**. All wiring happens in [/backend/internal/app/app.go](./backend/internal/app/app.go):

**Architecture Layers:**
```
HTTP Request
    ↓
Handler (routes, input validation)
    ↓
UseCase (business logic orchestration)
    ↓
Service + Repository (data & external calls)
    ↓
Model (domain entities)
    ↓
Database / Cache / Storage
```

> [!NOTE]  
> There is no test file for the moment (WIP)

**Run linter:**
```bash
cd backend
golangci-lint run

# Fix common issues
golangci-lint run --fix
```

---

## Frontend Development

```bash
# Install dependencies
pnpm install

# Start dev server with hot reload
task dev:dashboard
# or: pnpm run dev

# Linting and type checking
pnpm lint
pnpm check

# Build for production
task build:dashboard
# or: pnpm run build

#  embed it into the backend after build
task build:embed
```

## Nix

You can use the flake.nix config file for setup all dependances (Go, NodeJS, golangci and others).

```bash
nix shell

or 

direnv allow
```