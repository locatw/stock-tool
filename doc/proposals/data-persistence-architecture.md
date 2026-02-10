# Data Persistence Architecture

## Overview

Stock-tool adopts a lakehouse approach based on the Medallion Architecture to manage data across four layers:

| Layer | Purpose | Format | Example |
|---|---|---|---|
| **Landing** | Byte-for-byte preservation of source data | Original (JSON, CSV, etc.) | API response files, downloaded datasets |
| **Bronze** | Source data converted to a queryable columnar format | Iceberg (Parquet) | API responses as Iceberg tables |
| **Silver (Curated)** | Cleansed, normalized, analysis-ready data | Iceberg (Parquet) | Daily stock prices, brand master |
| **Gold (Analytics)** | Computed analysis results | Iceberg (Parquet) | Technical indicators, ML features, predictions |

Landing stores raw files as-is on Ceph. Bronze through Gold are stored as Apache Iceberg tables on the same Ceph storage. PostgreSQL remains solely for application state management (task scheduling, execution tracking).

## Infrastructure

- **Kubernetes** on a Proxmox cluster — all components run as k8s workloads
- **Ceph (RadosGW)** — S3-compatible object storage serving as the lakehouse storage layer
- **AWS** — minimal use (log aggregation); not the primary data store
- **PostgreSQL** — application metadata only (extract tasks, scheduling state)

### Language Roles

| Language | Responsibility |
|---|---|
| **Go** | Automated data extraction, Landing/Bronze/Silver writes, notifications, scheduled jobs |
| **Python** | Interactive analysis (Marimo), Gold layer generation, ML pipeline, experiment tracking |

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

Source data preserved byte-for-byte in its original format (JSON, CSV, XML, etc.) on Ceph. No parsing or transformation is applied — files are stored exactly as received from external sources.

- **Writers:** Go (automated extraction via k8s CronJob)
- **Readers:** Go (for Bronze conversion)
- **Key property:** Complete fidelity — if Bronze conversion logic changes or a bug is discovered, data can be re-ingested from Landing without fetching from the source again
- **Retention:** Long-term; serves as the system of record for all ingested data

Storage path convention:

```text
s3://locatw-{env}-stocktool-lakehouse/landing/{source}/{data_type}/{yyyy}/{mm}/{dd}/{timestamp}.{ext}
```

### Bronze (Queryable Raw)

Landing data converted to Iceberg tables (Parquet columnar format) for efficient querying. The conversion is purely a format change — column names, values, and structure mirror the source data with no business logic applied.

- **Writers:** Go (automated conversion from Landing)
- **Readers:** Go (for Silver transformation), Python (for exploratory analysis)
- **Key property:** Queryable via SQL (DuckDB) while preserving the original data semantics
- **Source-agnostic:** Each data source gets its own Bronze table(s); new sources are added without affecting existing ones

Storage path convention:

```text
s3://locatw-{env}-stocktool-lakehouse/bronze/{source}_{data_type}/
```

### Silver (Curated)

Cleansed, normalized, and type-converted data ready for analysis.

- **Writers:** Go (automated transformation from Bronze)
- **Readers:** Python (interactive analysis, feature engineering)
- **Transformations:** Data type casting, null handling, deduplication, field renaming to consistent conventions

Storage path convention:

```text
s3://locatw-{env}-stocktool-lakehouse/silver/{entity}/
```

### Gold (Analytics)

Computed results including technical indicators, ML features, and predictions.

- **Writers:** Python (interactive analysis, ML pipeline)
- **Readers:** Python (further analysis, model training, reporting)
- **Content:** Technical indicators (moving averages, RSI, etc.), prediction features, model outputs

Storage path convention:

```text
s3://locatw-{env}-stocktool-lakehouse/gold/{analysis_type}/
```

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

**Catalog:** Filesystem-based Iceberg catalog. Metadata is managed via Iceberg's metadata JSON files stored alongside data on Ceph.

**PostgreSQL role:** Application state only — extract task definitions, execution history, scheduling metadata. No analytical data.

**Characteristics:**

- Minimal components — DuckDB is the only engine, no catalog server needed
- Iceberg time travel enables querying historical data snapshots
- Parquet columnar format provides storage efficiency and fast analytical queries
- Fully local — no external service dependencies beyond Ceph

**Limitations:**

- Filesystem catalog has constraints on concurrent writes from multiple processes
- DuckDB Iceberg DML does not yet support UPDATE/DELETE on partitioned tables
- go-duckdb requires CGo (slightly more complex builds)

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

**What Polaris adds:**

- Centralized table registry — single source of truth for all Iceberg table metadata
- Safe concurrent access — Go CronJobs and Python analysis sessions do not conflict
- REST Catalog protocol — any Iceberg-compatible engine (DuckDB, PyIceberg, Spark, Trino, Flink) can connect
- PostgreSQL backend — reuses the existing PostgreSQL instance

**k8s deployment additions:**

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

**Characteristics:**

- All Phase 1 benefits plus safe multi-process access
- Unified k8s container management
- Engine extensibility — add Spark on k8s or Trino later without changing the storage layer

**Limitations:**

- Polaris is JVM-based — higher memory footprint than Phase 1
- Additional k8s manifests to manage
- More complex initial setup

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
