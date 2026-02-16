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

Rationale: independent release cycles (app changes don't trigger GitOps until manifest update), clean audit trail (infra Git history = deployment changes only), separate access control.

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

`stock-tool` contains: source code (Go, Python), Dockerfiles, CI pipelines, dev tooling (`compose.yaml`, linters, tests). All k8s manifests and ArgoCD definitions live in the infrastructure repository.

### Manifest Management

All k8s manifests use Kustomize base/overlay pattern. Each environment (production, local) has its own overlay directory patching the shared base.

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

Infrastructure in containers, application on host. Extends existing `compose.yaml`.

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

- `compose.yaml` runs PostgreSQL, SeaweedFS, etc.; Go/Python execute natively for fast iteration
- Compose profiles group services by function (`docker compose --profile lakehouse up`)
- Service/env-var names align with k8s deployment for Phase 2 transition
- Environment variables via `.env` file

### Phase 2: Kind + Kustomize Overlay

Transitions to Kind with Kustomize overlays once the production k8s cluster is operational.

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

**Overlay strategy:** Local overlays live alongside production overlays in the infrastructure repository. Rationale: single repo for all environments ensures consistency; base changes reflect in all overlays immediately.

**Local overlay patches:**

| Production | Local override |
|---|---|
| Ceph (RadosGW) | SeaweedFS |
| Production resource requests/limits | Relaxed or removed |
| External ingress | NodePort or port-forward |
| Replicas for HA | Single replica |

**Image loading:** `kind load docker-image stock-tool:latest --name stock-tool`

## Phased Adoption

```text
Phase 1: Docker Compose
  - Extend compose.yaml with profiles for lakehouse services (SeaweedFS, etc.)
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
