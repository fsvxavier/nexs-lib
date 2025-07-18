-- Sample data for nexs-lib testing and examples
-- This script populates the database with test data

-- Insert sample users
INSERT INTO users (username, email, password_hash, first_name, last_name) VALUES
('admin', 'admin@nexs.com', crypt('admin123', gen_salt('bf')), 'Admin', 'User'),
('john_doe', 'john@example.com', crypt('password123', gen_salt('bf')), 'John', 'Doe'),
('jane_smith', 'jane@example.com', crypt('password123', gen_salt('bf')), 'Jane', 'Smith'),
('bob_wilson', 'bob@example.com', crypt('password123', gen_salt('bf')), 'Bob', 'Wilson'),
('alice_brown', 'alice@example.com', crypt('password123', gen_salt('bf')), 'Alice', 'Brown'),
('test_user', 'test@nexs.com', crypt('test123', gen_salt('bf')), 'Test', 'User');

-- Insert sample products
INSERT INTO products (name, description, price, stock_quantity, category, sku) VALUES
('Laptop Pro 15"', 'High-performance laptop for developers', 1999.99, 25, 'Electronics', 'LAP-PRO-15'),
('Wireless Mouse', 'Ergonomic wireless mouse', 29.99, 100, 'Electronics', 'MOU-WIR-001'),
('Mechanical Keyboard', 'RGB mechanical keyboard', 149.99, 50, 'Electronics', 'KEY-MEC-RGB'),
('Monitor 4K 27"', '4K resolution monitor', 399.99, 30, 'Electronics', 'MON-4K-27'),
('USB-C Hub', 'Multi-port USB-C hub', 79.99, 75, 'Electronics', 'HUB-USBC-001'),
('Desk Lamp LED', 'Adjustable LED desk lamp', 45.99, 60, 'Office', 'LAM-LED-001'),
('Office Chair', 'Ergonomic office chair', 299.99, 20, 'Office', 'CHA-OFF-001'),
('Notebook A5', 'Premium notebook A5', 12.99, 200, 'Office', 'NOT-A5-001'),
('Pen Set', 'Professional pen set', 24.99, 150, 'Office', 'PEN-SET-001'),
('Coffee Mug', 'Ceramic coffee mug', 9.99, 300, 'Office', 'MUG-COF-001');

-- Insert sample orders
DO $$
DECLARE
    user_record RECORD;
    product_record RECORD;
    order_id UUID;
    i INTEGER;
BEGIN
    -- Create orders for each user
    FOR user_record IN SELECT id FROM users WHERE username != 'admin' LOOP
        FOR i IN 1..3 LOOP
            INSERT INTO orders (user_id, total_amount, status, shipping_address, billing_address)
            VALUES (
                user_record.id,
                0, -- Will be updated after adding items
                CASE 
                    WHEN i = 1 THEN 'completed'
                    WHEN i = 2 THEN 'processing'
                    ELSE 'pending'
                END,
                '123 Main St, City, State 12345',
                '123 Main St, City, State 12345'
            ) RETURNING id INTO order_id;
            
            -- Add 2-4 random items to each order
            FOR j IN 1..(1 + floor(random() * 4)) LOOP
                SELECT * INTO product_record FROM products ORDER BY random() LIMIT 1;
                
                INSERT INTO order_items (order_id, product_id, quantity, unit_price, total_price)
                VALUES (
                    order_id,
                    product_record.id,
                    1 + floor(random() * 3), -- 1-3 quantity
                    product_record.price,
                    product_record.price * (1 + floor(random() * 3))
                );
            END LOOP;
            
            -- Update order total
            UPDATE orders SET total_amount = (
                SELECT SUM(total_price) FROM order_items WHERE order_items.order_id = orders.id
            ) WHERE id = order_id;
        END LOOP;
    END LOOP;
END $$;

-- Insert sample notifications
INSERT INTO notifications (user_id, title, message, type) 
SELECT 
    u.id,
    'Welcome to NEXS!',
    'Thank you for joining our platform. Explore our features and enjoy your experience.',
    'welcome'
FROM users u WHERE u.username != 'admin';

INSERT INTO notifications (user_id, title, message, type, is_read) 
SELECT 
    o.user_id,
    'Order Confirmed',
    'Your order #' || o.id || ' has been confirmed and is being processed.',
    'order',
    CASE WHEN o.status = 'completed' THEN true ELSE false END
FROM orders o;

-- Insert sample user sessions
INSERT INTO user_sessions (user_id, session_token, expires_at)
SELECT 
    u.id,
    'sess_' || u.username || '_' || extract(epoch from now()),
    now() + interval '24 hours'
FROM users u WHERE u.username != 'admin';

-- Insert sample tenant data
INSERT INTO tenant_1.tenant_data (data_value) VALUES
('Tenant 1 Sample Data 1'),
('Tenant 1 Sample Data 2'),
('Tenant 1 Sample Data 3');

INSERT INTO tenant_2.tenant_data (data_value) VALUES
('Tenant 2 Sample Data 1'),
('Tenant 2 Sample Data 2'),
('Tenant 2 Sample Data 3');

-- Insert performance test data
INSERT INTO performance_test (test_data)
SELECT 'Performance test data ' || i
FROM generate_series(1, 1000) i;

-- Create some indexes for performance testing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_performance_test_id ON performance_test(id);

-- Update statistics
ANALYZE;

-- Show database information
SELECT 
    schemaname,
    relname as tablename,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes
FROM pg_stat_user_tables
ORDER BY schemaname, relname;
