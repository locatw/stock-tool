# J-Quants

Source profile extending [data-acquisition.md](../data-acquisition.md) with J-Quants-specific constraints. API base: `https://api.jquants.com/v1`. Authoritative reference for update times: [J-Quants data update page](https://jpx.gitbook.io/j-quants-ja/outline/data-update).

## Source-Level Configuration

Maps FR-11 source-level items to J-Quants values.

| # | Item | Value |
|---|---|---|
| S1 | Timezone | `Asia/Tokyo` (JST) |
| S2 | Rate limits and mitigation | TBD — document API rate limits and strategy once measured |
| S3 | Max concurrent executions | TBD — determine safe concurrency level |
| — | `plan` (J-Quants-specific) | Subscription plan (`free`, `light`, `standard`, `premium`) — determines historical limit and constraints |

## Plan-Based Historical Limits

| Plan | Historical Limit | Additional Constraints |
|---|---|---|
| Free | ~2 years | 12-week delay on recent data |
| Light | ~5 years | — |
| Standard | ~10 years | — |
| Premium | All data (~2008-) | — |

- Subscription plan is a source-level DB config item; changing it adjusts historical limit and constraints without code changes
- Free plan 12-week delay: dates within the delay window must be excluded from the expected-dates set for gap detection (FR-4)
- Constraint lifts automatically when plan setting changes

## Trading Calendar Dependency

- `trading_calendar` endpoint provides TSE business days (excludes weekends and Japanese holidays)
- Serves as gap detection reference for all J-Quants data types
- Bootstrap: must be fetched first on initial setup — no calendar means no accurate gap detection
- Updated yearly (~end of March)

## Data Types

### Daily

| Data Type | Update Time (JST) | Notes |
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

### Weekly

| Data Type | Update Time (JST) | Notes |
|---|---|---|
| weekly_margin_trading | 2nd business day ~16:30 | Weekly margin interest |
| trading_by_investor_type | 4th business day ~18:00 | Investor type trading data |

### Irregular

| Data Type | Update Time (JST) | Notes |
|---|---|---|
| earnings_calendar | ~19:00 when page updates | Earnings announcement calendar |
| trading_calendar | Yearly, ~end of March | TSE trading calendar |

### Multiple Update Windows

`financial_statements` and `statements` have two daily windows: preliminary (~18:00 JST, partial) + final (~24:30 JST, complete). Each window is an entry in the `update_times` array. System must fetch at both windows to capture the complete dataset.

## Data-Type-Level Configuration

Maps FR-11 data-type-level items to J-Quants defaults/values.

| # | Item | J-Quants Default | Notes |
|---|---|---|---|
| D1 | Update time | Per data type tables above | Array for multi-window types |
| D2 | Update frequency | Per data type tables above | `daily`, `weekly`, or `irregular` |
| D3 | Processing range per execution | TBD | Determine per data type based on API response size |
| D4 | Backfill behavior | TBD | Plan-based historical limit bounds backfill range; Free plan delay excludes recent dates |
| D5 | Backfill target | `true` | All types subject to gap detection by default |
| D6 | Re-run strategy | Per FR-10 decision | — |
| D7 | Retry policy | 3 retries, exponential backoff, retry on 429/5xx/timeout | — |
| D8 | Empty response handling | `success` | — |
| D9 | Dependencies | `trading_calendar` for all gap-detected types | Calendar must exist before gap detection runs |
| D10 | Stale execution timeout | Source-level default | — |

## Constraints

- Plan-based historical limit bounds all backfill and gap detection ranges
- Free plan 12-week delay: recent dates excluded from expected-dates set for gap detection
- Trading calendar must be bootstrapped before gap detection is accurate for any data type
- Multi-window types (`financial_statements`, `statements`) require fetches at both update windows
