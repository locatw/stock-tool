# J-Quants Data Ingestion

## Overview

J-Quants-specific requirements extending the general framework in [data-ingestion-requirements.md](data-ingestion-requirements.md). API base: `https://api.jquants.com/v1`.

## Data Types and Update Timing

J-Quants provides multiple data types, each with its own update schedule. The authoritative reference for update times is the [J-Quants data update page](https://jpx.gitbook.io/j-quants-ja/outline/data-update).

### Daily Data Types

| Data Type | Expected Update Time (JST) | Notes |
|---|---|---|
| listed_info | ~17:30 | Listed issue information |
| daily_quotes | ~16:30 | Daily OHLCV quotes |
| financial_statements | ~18:00 / ~24:30 | Two windows: preliminary and final |
| statements | ~18:00 / ~24:30 | Two windows: preliminary and final |
| index_option | ~27:00 | Next-day early morning |
| prices_am | ~12:00 | Morning session quotes |
| trades_spec | ~16:30 / ~17:30 | Short selling data, two windows |
| margin_trading | ~16:30 | Margin interest data |
| breakdown | ~18:00 | Trading breakdown |
| dividend | 12:00-19:00 (hourly) | Dividend information, updated hourly within the window |
| indices (TOPIX) | ~16:30 | TOPIX index data |
| futures | ~27:00 | Futures data |
| options | ~27:00 | Options data |

### Weekly Data Types

| Data Type | Expected Update Time (JST) | Notes |
|---|---|---|
| weekly_margin_trading | 2nd business day ~16:30 | Weekly margin interest |
| trading_by_investor_type | 4th business day ~18:00 | Investor type trading data |

### Irregular Data Types

| Data Type | Expected Update Time (JST) | Notes |
|---|---|---|
| earnings_calendar | ~19:00 when page updates | Earnings announcement calendar |
| trading_calendar | Yearly, ~end of March | TSE trading calendar |

### Types with Multiple Daily Windows

`financial_statements` and `statements`: preliminary (~18:00 JST, partial) + final (~24:30 JST, complete). System must fetch at both windows (each entry in `update_times`).

## Plan-Based Historical Limits

J-Quants subscription plan determines the maximum lookback period for historical data:

| Plan | Historical Limit | Additional Constraints |
|---|---|---|
| Free | ~2 years | 12-week delay on recent data |
| Light | ~5 years | — |
| Standard | ~10 years | — |
| Premium | All data (~2008-) | — |

Current subscription: Free plan (upgrade planned).

### Plan Configuration

Subscription plan is a source-level DB config item. Changing it adjusts historical limit and plan-specific constraints without code changes.

Free plan constraint: 12-week delay on recent data — dates within the delay window must be excluded from the expected-dates set for gap detection. Constraint lifts automatically when plan setting changes.

## Trading Calendar Dependency

Gap detection requires TSE business days (excludes weekends and Japanese holidays).

- Source: `trading_calendar` endpoint — updated yearly (~end of March)
- Ingestion: Treated as a data type in the same extraction pipeline; serves as gap detection reference for all other J-Quants data types
- Bootstrap: Must be fetched first on initial setup (no calendar = no accurate gap detection)

## J-Quants Configuration Items

Maps to [general configuration framework](data-ingestion-requirements.md#extensible-per-data-source-configuration).

### Source-Level Settings

| Setting | Description | Example Value |
|---|---|---|
| plan | Subscription plan — determines historical limit and constraints | `free` |

### Data-Type-Level Settings

| Setting | Description | Example Value |
|---|---|---|
| update_times | Expected update time(s) in JST | `["16:30"]` or `["18:00", "24:30"]` |
| update_frequency | How often data is updated | `daily`, `weekly`, `irregular` |
| backfill_enabled | Whether to backfill missing dates | `true` |
| enabled | Whether to extract this data type | `true` |

`update_times` is an array to support multiple daily windows (e.g., `financial_statements`).
