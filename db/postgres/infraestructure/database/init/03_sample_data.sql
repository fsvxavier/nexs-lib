-- Sample data for nexs-lib testing and examples
-- This script populates the database with test data based on the examples

-- ===========================================
-- PRODUCTS DATA (for batch, copy, transaction examples)
-- ===========================================

INSERT INTO products (name, price, category) VALUES
('Laptop Gaming', 1299.99, 'Electronics'),
('Mouse Wireless', 29.99, 'Electronics'),
('Keyboard Mechanical', 149.99, 'Electronics'),
('Monitor 4K', 399.99, 'Electronics'),
('Headset Gaming', 79.99, 'Electronics'),
('Smartphone Pro', 899.99, 'Electronics'),
('Tablet 10"', 329.99, 'Electronics'),
('Smartwatch', 249.99, 'Electronics'),
('Camera DSLR', 699.99, 'Electronics'),
('Speaker Bluetooth', 59.99, 'Electronics'),
('Notebook Premium', 12.99, 'Office'),
('Pen Professional', 24.99, 'Office'),
('Desk Organizer', 19.99, 'Office'),
('Coffee Mug', 9.99, 'Office'),
('Lamp LED', 45.99, 'Office'),
('Chair Ergonomic', 299.99, 'Furniture'),
('Desk Standing', 449.99, 'Furniture'),
('Bookshelf', 129.99, 'Furniture'),
('Filing Cabinet', 89.99, 'Furniture'),
('Whiteboard', 69.99, 'Office')
ON CONFLICT DO NOTHING;

-- ===========================================
-- ACCOUNTS DATA (for transaction examples)
-- ===========================================

INSERT INTO accounts (name, balance) VALUES
('Alice Johnson', 1000.00),
('Bob Smith', 500.00),
('Charlie Brown', 750.00),
('Diana Prince', 1250.00),
('Eve Wilson', 300.00),
('Frank Miller', 2000.00),
('Grace Lee', 850.00),
('Henry Davis', 1500.00),
('Ivy Chen', 425.00),
('Jack Taylor', 975.00)
ON CONFLICT DO NOTHING;

-- ===========================================
-- COPY_TEST DATA (for COPY operations examples)
-- ===========================================

INSERT INTO copy_test (name, email, age, salary, department, hire_date, active) VALUES
('John Doe', 'john.doe@company.com', 28, 5500.00, 'Engineering', '2023-01-15', true),
('Jane Smith', 'jane.smith@company.com', 32, 6200.00, 'Marketing', '2022-03-22', true),
('Mike Johnson', 'mike.johnson@company.com', 35, 7500.00, 'Engineering', '2021-07-10', true),
('Sarah Wilson', 'sarah.wilson@company.com', 29, 5800.00, 'HR', '2023-02-01', true),
('Tom Brown', 'tom.brown@company.com', 41, 8200.00, 'Sales', '2020-05-15', true),
('Lisa Davis', 'lisa.davis@company.com', 27, 5200.00, 'Marketing', '2023-04-12', true),
('David Miller', 'david.miller@company.com', 33, 6800.00, 'Engineering', '2022-08-30', true),
('Emma Taylor', 'emma.taylor@company.com', 26, 4900.00, 'HR', '2023-06-05', true),
('Chris Anderson', 'chris.anderson@company.com', 38, 7200.00, 'Sales', '2021-11-20', true),
('Amy White', 'amy.white@company.com', 31, 6500.00, 'Marketing', '2022-12-08', true),
('Robert Lee', 'robert.lee@company.com', 45, 9500.00, 'Engineering', '2019-03-14', true),
('Jennifer Garcia', 'jennifer.garcia@company.com', 34, 7800.00, 'Sales', '2021-09-25', true),
('Kevin Martinez', 'kevin.martinez@company.com', 29, 5600.00, 'HR', '2023-01-30', true),
('Michelle Rodriguez', 'michelle.rodriguez@company.com', 36, 8900.00, 'Engineering', '2020-10-12', true),
('Daniel Kim', 'daniel.kim@company.com', 28, 5400.00, 'Marketing', '2023-03-18', true)
ON CONFLICT DO NOTHING;

-- ===========================================
-- TENANTS DATA (for multi-tenancy examples)
-- ===========================================

INSERT INTO tenants (name, schema_name, active) VALUES
('Empresa A', 'tenant_empresa_a', true),
('Empresa B', 'tenant_empresa_b', true),
('Empresa C', 'tenant_empresa_c', true),
('Test Company', 'tenant_test_company', true),
('Demo Corp', 'tenant_demo_corp', false)
ON CONFLICT (name) DO NOTHING;

-- ===========================================
-- SHARED_USERS DATA (for row-level multi-tenancy)
-- ===========================================

INSERT INTO shared_users (tenant_id, name, email) VALUES
-- Empresa A users
(1, 'João Silva', 'joao@empresaa.com'),
(1, 'Maria Santos', 'maria@empresaa.com'),
(1, 'Pedro Costa', 'pedro@empresaa.com'),
-- Empresa B users
(2, 'Ana Oliveira', 'ana@empresab.com'),
(2, 'Carlos Lima', 'carlos@empresab.com'),
(2, 'Fernanda Rocha', 'fernanda@empresab.com'),
-- Empresa C users
(3, 'Ricardo Alves', 'ricardo@empresac.com'),
(3, 'Lucia Mendes', 'lucia@empresac.com'),
(3, 'Bruno Pereira', 'bruno@empresac.com'),
-- Test Company users
(4, 'Test User 1', 'test1@testcompany.com'),
(4, 'Test User 2', 'test2@testcompany.com')
ON CONFLICT (tenant_id, email) DO NOTHING;

-- ===========================================
-- TENANT SCHEMA USERS (for schema-based multi-tenancy)
-- ===========================================

INSERT INTO tenant_empresa_a.users (name, email) VALUES
('João Silva', 'joao@empresaa.com'),
('Maria Santos', 'maria@empresaa.com'),
('Pedro Costa', 'pedro@empresaa.com'),
('Ana Ferreira', 'ana@empresaa.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO tenant_empresa_b.users (name, email) VALUES
('Ana Oliveira', 'ana@empresab.com'),
('Carlos Lima', 'carlos@empresab.com'),
('Fernanda Rocha', 'fernanda@empresab.com'),
('Roberto Silva', 'roberto@empresab.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO tenant_empresa_c.users (name, email) VALUES
('Ricardo Alves', 'ricardo@empresac.com'),
('Lucia Mendes', 'lucia@empresac.com'),
('Bruno Pereira', 'bruno@empresac.com'),
('Carla Souza', 'carla@empresac.com')
ON CONFLICT (email) DO NOTHING;

-- ===========================================
-- CHAT MESSAGES DATA (for LISTEN/NOTIFY examples)
-- ===========================================

INSERT INTO chat_messages (channel, username, message) VALUES
('general', 'admin', 'Welcome to the chat system!'),
('general', 'user1', 'Hello everyone!'),
('general', 'user2', 'How is everyone doing?'),
('tech', 'developer1', 'Discussing new features'),
('tech', 'developer2', 'Code review completed'),
('support', 'support1', 'How can I help you?'),
('support', 'customer1', 'I need help with my account'),
('random', 'user3', 'Random message here'),
('random', 'user4', 'Another random message'),
('notifications', 'system', 'System maintenance scheduled for tonight')
ON CONFLICT DO NOTHING;

-- ===========================================
-- MONITORED TABLE DATA (for LISTEN/NOTIFY examples)
-- ===========================================

INSERT INTO monitored_table (data) VALUES
('Initial data 1'),
('Initial data 2'),
('Initial data 3'),
('Sample monitoring data'),
('Change detection test'),
('Notification trigger test'),
('Database activity log'),
('System status update'),
('Performance metrics'),
('Error tracking data')
ON CONFLICT DO NOTHING;

-- ===========================================
-- REPLICA TEST DATA (for read replica examples)
-- ===========================================

INSERT INTO replica_test (message) VALUES
('Test message 1 for replica'),
('Test message 2 for replica'),
('Test message 3 for replica'),
('Replication test data'),
('Read replica verification'),
('Data consistency check'),
('Load balancing test'),
('Failover scenario test'),
('Performance comparison'),
('Sync status verification')
ON CONFLICT DO NOTHING;

-- ===========================================
-- PERFORMANCE TEST DATA
-- ===========================================

INSERT INTO performance_test (test_data)
SELECT 'Performance test data batch ' || i || ' - ' || md5(random()::text)
FROM generate_series(1, 1000) i
ON CONFLICT DO NOTHING;

-- ===========================================
-- SAMPLE AUDIT DATA
-- ===========================================

INSERT INTO audit_log (table_name, operation, new_values, user_id) VALUES
('products', 'INSERT', '{"id": 1, "name": "Sample Product", "price": 99.99}', 'admin'),
('accounts', 'INSERT', '{"id": 1, "name": "Sample Account", "balance": 1000.00}', 'admin'),
('tenants', 'INSERT', '{"id": 1, "name": "Sample Tenant", "active": true}', 'admin')
ON CONFLICT DO NOTHING;

-- ===========================================
-- ADDITIONAL TEST DATA
-- ===========================================

-- Insert test data for concurrent operations
INSERT INTO test_concurrent (worker_id, task_id) VALUES
(1, 1), (1, 2), (1, 3),
(2, 1), (2, 2), (2, 3),
(3, 1), (3, 2), (3, 3)
ON CONFLICT DO NOTHING;

-- Insert test data for transaction operations
INSERT INTO test_transactions (name, amount) VALUES
('Transaction 1', 100.00),
('Transaction 2', 250.00),
('Transaction 3', 500.00),
('Transaction 4', 750.00),
('Transaction 5', 1000.00)
ON CONFLICT DO NOTHING;

-- ===========================================
-- FINALIZATION
-- ===========================================

-- Update statistics for better query performance
ANALYZE;

-- Display sample data summary
SELECT 
    'products' as table_name, 
    COUNT(*) as record_count 
FROM products
UNION ALL
SELECT 
    'accounts' as table_name, 
    COUNT(*) as record_count 
FROM accounts
UNION ALL
SELECT 
    'copy_test' as table_name, 
    COUNT(*) as record_count 
FROM copy_test
UNION ALL
SELECT 
    'tenants' as table_name, 
    COUNT(*) as record_count 
FROM tenants
UNION ALL
SELECT 
    'shared_users' as table_name, 
    COUNT(*) as record_count 
FROM shared_users
UNION ALL
SELECT 
    'chat_messages' as table_name, 
    COUNT(*) as record_count 
FROM chat_messages
UNION ALL
SELECT 
    'monitored_table' as table_name, 
    COUNT(*) as record_count 
FROM monitored_table
UNION ALL
SELECT 
    'replica_test' as table_name, 
    COUNT(*) as record_count 
FROM replica_test
UNION ALL
SELECT 
    'performance_test' as table_name, 
    COUNT(*) as record_count 
FROM performance_test
ORDER BY table_name;
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
