CREATE TABLE audit_events (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    prev_hash TEXT,
    current_hash TEXT NOT NULL
);
