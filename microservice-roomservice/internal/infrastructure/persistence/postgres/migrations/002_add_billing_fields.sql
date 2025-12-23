-- Migration: Add billing fields and update status constraints
-- This migration adds support for the saga orchestration flow

-- Add billing_folder_id column
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS billing_folder_id VARCHAR(255);

-- Add billing_ticket_id column
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS billing_ticket_id VARCHAR(255);

-- Drop existing status constraint
ALTER TABLE orders 
DROP CONSTRAINT IF EXISTS orders_status_check;

-- Add updated status constraint with new saga statuses
ALTER TABLE orders 
ADD CONSTRAINT orders_status_check 
CHECK (status IN ('pending', 'pending_kitchen', 'pending_billing', 'preparing', 'completed', 'cancelled'));

-- Add indexes for billing fields (optional, for performance)
CREATE INDEX IF NOT EXISTS idx_orders_billing_folder_id ON orders(billing_folder_id);
CREATE INDEX IF NOT EXISTS idx_orders_billing_ticket_id ON orders(billing_ticket_id);

-- Comments for documentation
COMMENT ON COLUMN orders.billing_folder_id IS 'Billing folder ID from billing service (saga orchestration)';
COMMENT ON COLUMN orders.billing_ticket_id IS 'Billing ticket ID from billing service (saga orchestration)';
COMMENT ON COLUMN orders.status IS 'Order status: pending (legacy), pending_kitchen (waiting for kitchen), pending_billing (waiting for billing), preparing (in kitchen), completed, cancelled';
