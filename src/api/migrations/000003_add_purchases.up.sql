CREATE TABLE purchases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    purchase_date DATE NOT NULL,
    amount_cents INTEGER NOT NULL,
    installment_count INTEGER NOT NULL CHECK (installment_count BETWEEN 2 AND 24),
    category_id INTEGER NOT NULL REFERENCES categories (id),
    category_name TEXT NOT NULL,
    deleted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_purchases_category_id ON purchases (category_id);

ALTER TABLE entries ADD COLUMN purchase_id INTEGER REFERENCES purchases (id);
ALTER TABLE entries ADD COLUMN installment_number INTEGER;

CREATE INDEX idx_entries_purchase_id ON entries (purchase_id);
