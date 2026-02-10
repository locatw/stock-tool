# Local Development Environment

## Overview

Stock-tool runs on Kubernetes (Proxmox cluster) in production.
This document describes the repository strategy for separating source code from deployment manifests and the phased approach to building a local development environment.

The local environment is established in two phases:

| Phase | Approach | Purpose |
|---|---|---|
| **Phase 1** | Docker Compose | Lightweight local environment using existing tooling |
| **Phase 2** | Kind + Kustomize | Local k8s cluster mirroring production topology |

## Repository Strategy

### Source Code and Deployment Manifest Separation

Following ArgoCD best practices, source code and Kubernetes manifests are maintained in separate repositories:

- **Application repository** (`stock-tool`) — source code, Dockerfiles, CI pipelines
- **Infrastructure repository** — all Kubernetes manifests, ArgoCD configuration, platform services

This separation provides:

- **Independent release cycles** — application code changes do not trigger GitOps reconciliation until a manifest update references the new image tag
- **Clean audit trail** — Git history in the infrastructure repository reflects only deployment changes
- **Access control** — different teams or automation can own each repository independently

### Infrastructure Repository Structure

Platform GitOps and application configuration are managed as a single monorepo. All manifests use Kustomize with base/overlay structure.

```text
infra-repo/
├── infrastructure/              # Platform services (Prometheus, Grafana, etc.)
│   ├── base/
│   └── overlays/
│       ├── production/
│       └── local/
├── apps/                        # Application k8s manifests
│   ├── base/
│   │   └── stock-tool/
│   └── overlays/
│       ├── production/
│       │   └── stock-tool/
│       └── local/
│           └── stock-tool/
└── clusters/
    └── production/              # ArgoCD Application definitions
```

### Application Repository Scope

The `stock-tool` repository contains only:

- Application source code (Go, Python)
- Dockerfiles for building container images
- CI pipeline definitions
- Development tooling (`compose.yml`, linters, test configuration)

Kubernetes manifests, Kustomize overlays, and ArgoCD Application definitions reside exclusively in the infrastructure repository.

### Manifest Management

All Kubernetes manifests — both infrastructure services and application workloads — are managed with Kustomize using a base/overlay pattern.
Each environment (production, local) has its own overlay directory that patches the shared base.

### Namespace Strategy

Namespaces are organized by functional area:

| Namespace pattern | Contents |
|---|---|
| `monitoring` | Prometheus, Grafana |
| `logging` | Log aggregation stack |
| `ingress` | Ingress controllers |
| `stocktool-{env}` | Application workloads per environment |

### GitOps

ArgoCD manages deployments using a combination of App-of-Apps and ApplicationSet patterns:

- **App-of-Apps** — a root Application that manages other Application definitions in `clusters/{env}/`
- **ApplicationSet** — generates Application resources dynamically based on directory structure or Git repository contents

## Local Development Environment

### Phase 1: Docker Compose

Docker Compose provides infrastructure services locally while application code runs natively on the host. This extends the existing `compose.yml` configuration.

**Architecture:**

```text
┌─────────────────────────────────────────────┐
│            Docker Compose                    │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │PostgreSQL │  │SeaweedFS │  │  Other   │  │
│  │  (db)     │  │ (s3)     │  │ services │  │
│  └──────────┘  └──────────┘  └──────────┘  │
└──────────────────┬──────────────────────────┘
                   │ localhost
    ┌──────────────┼──────────────┐
    │                             │
  Go (native)            Python (native)
  - go run / go test      - marimo / pytest
```

**Principles:**

- **Infrastructure in containers, application on host** — `compose.yml` runs PostgreSQL, SeaweedFS, and other infrastructure services; Go and Python execute natively for fast iteration and debugger access
- **Compose profiles** — group services by function so developers start only what they need (e.g., `docker compose --profile lakehouse up`)
- **Consistent naming** — service names and environment variable names align with the Kubernetes deployment to minimize configuration differences when transitioning to Phase 2
- **Environment variables** — managed via `.env` file, following the existing pattern in the repository

### Phase 2: Kind + Kustomize Overlay

Once the production Kubernetes cluster is operational in the infrastructure repository, the local environment transitions to Kind with Kustomize overlays.

**Architecture:**

```text
┌──────────────────────────────────────────────┐
│              Kind cluster                     │
│                                               │
│  ┌──────────────────────────────────────────┐ │
│  │  stocktool-local namespace               │ │
│  │  ┌──────────┐  ┌──────────┐             │ │
│  │  │ App Pod  │  │PostgreSQL│             │ │
│  │  └──────────┘  └──────────┘             │ │
│  └──────────────────────────────────────────┘ │
│  ┌──────────────────────────────────────────┐ │
│  │  infrastructure namespaces               │ │
│  │  SeaweedFS (replacing Ceph), etc.        │ │
│  └──────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

**Overlay strategy (Pattern B: overlays in infrastructure repository):**

Local overlays live alongside production overlays in the infrastructure repository. This approach ensures:

- Base manifests and all overlays are managed in a single repository, maintaining consistency
- All environments (production, local) are visible in one place
- Changes to base manifests are immediately reflected in all overlays

**Local overlay patches:**

| Production | Local override |
|---|---|
| Ceph (RadosGW) | SeaweedFS |
| Production resource requests/limits | Relaxed or removed |
| External ingress | NodePort or port-forward |
| Replicas for HA | Single replica |

**Image loading:**

Application container images are loaded into the Kind cluster directly from the local Docker daemon:

```bash
kind load docker-image stock-tool:latest --name stock-tool
```

## Phased Adoption

```text
Phase 1: Docker Compose
  - Extend compose.yml with profiles for lakehouse services (SeaweedFS, etc.)
  - Run Go/Python natively on host against containerized infrastructure
  - Align service names and env vars with future k8s resource names
      │
      ▼  Trigger: production k8s cluster becomes operational
Phase 2: Kind + Kustomize Overlay
  - Create local overlays in the infrastructure repository
  - Deploy to Kind cluster using kustomize build | kubectl apply
  - Load app images via kind load docker-image
  - Match production topology for integration testing
```
