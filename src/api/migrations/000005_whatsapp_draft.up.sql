CREATE TABLE whatsapp_inbound_messages (
    message_id TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE whatsapp_outbound_messages (
    message_id TEXT PRIMARY KEY,
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE whatsapp_outbound_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_body TEXT NOT NULL DEFAULT '',
    match_text INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE whatsapp_draft (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    source_text TEXT NOT NULL,
    amount_cents INTEGER,
    entry_type TEXT,
    category_id INTEGER,
    entry_date TEXT,
    missing_amount INTEGER NOT NULL,
    missing_type INTEGER NOT NULL,
    missing_category INTEGER NOT NULL,
    missing_date INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
