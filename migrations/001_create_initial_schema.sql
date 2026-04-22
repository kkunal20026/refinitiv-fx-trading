# Migration: 001_create_initial_schema

## Up

```sql
CREATE TABLE IF NOT EXISTS rates (
    id UUID PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    bid DECIMAL(18, 8) NOT NULL,
    ask DECIMAL(18, 8) NOT NULL,
    mid DECIMAL(18, 8) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    source VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(from_currency, to_currency, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_rates_pair ON rates(from_currency, to_currency);
CREATE INDEX IF NOT EXISTS idx_rates_timestamp ON rates(timestamp DESC);

CREATE TABLE IF NOT EXISTS deals (
    id UUID PRIMARY KEY,
    client_id VARCHAR(100) NOT NULL,
    trade_id VARCHAR(100) NOT NULL UNIQUE,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    amount DECIMAL(18, 2) NOT NULL,
    rate DECIMAL(18, 8) NOT NULL,
    status VARCHAR(20) NOT NULL,
    direction VARCHAR(10) NOT NULL,
    value_date TIMESTAMP NOT NULL,
    settlement_date TIMESTAMP,
    reference VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deals_client_id ON deals(client_id);
CREATE INDEX IF NOT EXISTS idx_deals_trade_id ON deals(trade_id);
CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);
CREATE INDEX IF NOT EXISTS idx_deals_created_at ON deals(created_at DESC);
```

## Down

```sql
DROP INDEX IF EXISTS idx_deals_created_at;
DROP INDEX IF EXISTS idx_deals_status;
DROP INDEX IF EXISTS idx_deals_trade_id;
DROP INDEX IF EXISTS idx_deals_client_id;
DROP TABLE IF EXISTS deals;

DROP INDEX IF EXISTS idx_rates_timestamp;
DROP INDEX IF EXISTS idx_rates_pair;
DROP TABLE IF EXISTS rates;
```
