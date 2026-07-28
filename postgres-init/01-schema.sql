-- ============================================================================
-- Schema for the simplified FIN database (users, transactions only)
-- Run with: psql -U <user> -d oteldemo -f 01-schema.sql
--
-- NOTE: reconstructed from the column structure shown via `psql \d`/`SELECT *`
-- earlier in this conversation, not your original DDL file. Column types
-- (NUMERIC precision, VARCHAR lengths, the `type` CHECK constraint) are
-- reasonable best-effort guesses — adjust anything that doesn't match what
-- you actually had.
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS transactions (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    amount       NUMERIC(12, 2) NOT NULL,
    currency     CHAR(3) NOT NULL,
    type         VARCHAR(20) NOT NULL CHECK (type IN ('credit', 'debit')),
    merchant     VARCHAR(255),
    description  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions (created_at);

-- updated_at auto-touch trigger (users only — transactions has no updated_at column)
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
