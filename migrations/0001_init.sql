-- Operators table
CREATE TABLE operators (
   id SERIAL PRIMARY KEY,
   name TEXT NOT NULL,
   api_key TEXT UNIQUE NOT NULL,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Players table
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    external_player_id TEXT NOT NULL,
    jurisdiction TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(operator_id, external_player_id)
);

-- Sessions table
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    operator_id INT NOT NULL REFERENCES operators(id),
    player_id INT NOT NULL REFERENCES players(id),
    launch_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);