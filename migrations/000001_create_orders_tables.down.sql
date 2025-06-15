-- Drop indexes
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_created_at_id;

-- Drop constraints (they will be dropped automatically when tables are dropped)
-- But we can be explicit for clarity
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_total_amount;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS chk_orders_status;

-- Drop tables (order matters due to foreign key constraints)
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
