DROP INDEX idx_entries_purchase_id;
ALTER TABLE entries DROP COLUMN installment_number;
ALTER TABLE entries DROP COLUMN purchase_id;
DROP TABLE purchases;
