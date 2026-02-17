# Data Persistence Architecture

## Overview

Stock-tool adopts a lakehouse approach based on the Medallion Architecture to manage data across four layers:

| Layer | Purpose | Format | Example |
|---|---|---|---|
| Landing | Byte-for-byte preservation of source data | Original (JSON, CSV, etc.) | API response files, downloaded datasets |
| Bronze | Source data converted to a queryable columnar format | Iceberg (Parquet) | API responses as Iceberg tables |
| Silver (Curated) | Cleansed, normalized, analysis-ready data | Iceberg (Parquet) | Daily stock prices, brand master |
| Gold (Analytics) | Computed analysis results | Iceberg (Parquet) | Technical indicators, ML features, predictions |

Landing stores raw files on Ceph. Bronze–Gold are Apache Iceberg tables on the same Ceph storage. PostgreSQL: application state only (task scheduling, execution tracking).

## Infrastructure

- Kubernetes on a Proxmox cluster — all components run as k8s workloads
- Ceph (RadosGW) — S3-compatible object storage serving as the lakehouse storage layer
- AWS — minimal use (log aggregation); not the primary data store
- PostgreSQL — application metadata only (extract tasks, scheduling state)

### Language Roles

| Language | Responsibility |
|---|---|
| Go | Automated data extraction, Landing/Bronze/Silver writes, notifications, scheduled jobs |
| Python | Interactive analysis (Marimo), Gold layer generation, ML pipeline, experiment tracking |

## Technology Stack

| Component | Technology | Role |
|---|---|---|
| Table format | Apache Iceberg | Open table format with ACID transactions, time travel, schema evolution |
| Analytical engine | DuckDB | Shared SQL engine for both Go and Python via Iceberg extension |
| Go bindings | duckdb-go (official) | Go access to DuckDB (CGo required, pre-built static libraries available) |
| DataFrame | Polars | High-performance Arrow-native DataFrame library for Python analysis |
| Python Iceberg | PyIceberg | JVM-free Python access to Iceberg tables |
| Notebook | Marimo | Reactive notebook with pure Python files (Git-friendly, Jupyter alternative) |
| ML tracking | MLflow | Experiment tracking and model registry |
| Iceberg catalog | Apache Polaris | REST Catalog for multi-engine access (Phase 2) |

## Data Layers

### Landing (Raw Files)

- Writers: Go (k8s CronJob) | Readers: Go (Bronze conversion)
- Key property: Byte-for-byte fidelity — re-ingest from Landing without re-fetch
- Retention: Long-term system of record
- Path: `s3://locatw-{env}-stocktool-lakehouse/landing/{source}/{data_type}/{yyyy}/{mm}/{dd}/{timestamp}.{ext}`

### Bronze (Queryable Raw)

- Writers: Go (from Landing) | Readers: Go (Silver transform), Python (exploration)
- Key property: SQL-queryable (DuckDB) with original data semantics; format change only, no business logic
- Path: `s3://locatw-{env}-stocktool-lakehouse/bronze/{source}_{data_type}/`

### Silver (Curated)

- Writers: Go (from Bronze) | Readers: Python (analysis, feature engineering)
- Transformations: Type casting, null handling, dedup, field renaming
- Path: `s3://locatw-{env}-stocktool-lakehouse/silver/{entity}/`

### Gold (Analytics)

- Writers: Python (analysis, ML pipeline) | Readers: Python (model training, reporting)
- Content: Technical indicators, prediction features, model outputs
- Path: `s3://locatw-{env}-stocktool-lakehouse/gold/{analysis_type}/`

## Architecture

### Phase 1: DuckDB + Iceberg on Ceph

DuckDB serves as the shared analytical engine. Both Go and Python access the same Iceberg tables on Ceph through DuckDB's Iceberg extension and httpfs.

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                          Ceph (RadosGW / S3)                             │
│                                                                          │
│  s3://locatw-{env}-stocktool-lakehouse/                                  │
│  ├── landing/       Raw files (JSON, CSV, etc.)                          │
│  ├── bronze/        Iceberg tables (Queryable Raw)                       │
│  ├── silver/        Iceberg tables (Curated)                             │
│  └── gold/          Iceberg tables (Analytics)                           │
└──────────────────────────────────┬───────────────────────────────────────┘
                                   │
                            DuckDB (Iceberg extension + httpfs)
                                   │
                   ┌───────────────┼───────────────┐
                   │                               │
          Go (k8s CronJob/Pod)          Python (k8s Pod / local)
          - Data source -> Landing       - Marimo interactive analysis
          - Landing -> Bronze convert    - Gold layer generation
          - Bronze -> Silver transform   - ML pipeline + MLflow
          - Notifications
```

Catalog: Filesystem-based (Iceberg metadata JSON on Ceph). PostgreSQL: application state only.

Characteristics: Minimal components (DuckDB only, no catalog server), Iceberg time travel, Parquet columnar efficiency, fully local.

Limitations: Filesystem catalog limits concurrent writes; DuckDB Iceberg DML lacks UPDATE/DELETE on partitioned tables; go-duckdb requires CGo.

### Phase 2: + Apache Polaris Catalog

Adds Apache Polaris as an Iceberg REST Catalog server, deployed on k8s. This resolves the concurrent access limitation of Phase 1 and enables multi-engine interoperability.

```text
┌──────────────────────────────────────────────────────────────────────────┐
│                          Ceph (RadosGW / S3)                             │
│  s3://locatw-{env}-stocktool-lakehouse/{landing,bronze,silver,gold}/     │
└──────────────────────────────────┬───────────────────────────────────────┘
                                   │
                          Apache Polaris (k8s Pod)
                          - Iceberg REST Catalog
                          - PostgreSQL backend
                          - Centralized table metadata
                                   │
                   ┌───────────────┼───────────────┐
                   │               │               │
          Go (duckdb-go)     DuckDB CLI     Python (DuckDB/PyIceberg/Polars)
          - Extraction        - Ad-hoc       - Marimo analysis
          - Bronze/Silver      queries      - Gold generation
          - Notifications                   - ML + MLflow
```

What Polaris adds: Centralized table registry (single metadata source of truth), safe concurrent access (Go CronJobs + Python don't conflict), REST Catalog protocol (DuckDB, PyIceberg, Spark, Trino, Flink), PostgreSQL backend (reuses existing instance).

k8s deployment additions:

```yaml
# Polaris catalog server
polaris:
  image: apache/polaris
  env:
    CATALOG_BACKEND: postgresql

# Go extraction as CronJob
extract-job:
  schedule: "0 18 * * 1-5"

# MLflow for experiment tracking
mlflow:
  image: ghcr.io/mlflow/mlflow
```

Characteristics: All Phase 1 benefits + safe multi-process access, unified k8s management, engine extensibility (add Spark/Trino later).

Limitations: Polaris is JVM-based (higher memory), additional k8s manifests, more complex setup.

## Phased Adoption

```text
Phase 1: DuckDB + Iceberg on Ceph
  - Implement Go ingestion pipeline (data sources -> Landing -> Bronze -> Silver)
  - Set up Python + Marimo for interactive Silver/Gold analysis
  - Begin accumulating data on Ceph S3
      │
      ▼  When concurrent Go/Python access becomes necessary
Phase 2: + Apache Polaris Catalog
  - Deploy Polaris on k8s with PostgreSQL backend
  - Migrate from filesystem catalog to REST Catalog
  - Deploy MLflow on k8s for ML experiment tracking
```
