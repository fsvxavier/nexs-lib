-- Additional setup for nexs-lib examples
-- This script provides additional configurations and data specifically for examples

-- ===========================================
-- EXAMPLE-SPECIFIC CONFIGURATIONS
-- ===========================================

-- Set up row-level security for tenant isolation
-- This is disabled by default, examples will enable it as needed
ALTER TABLE shared_users DISABLE ROW LEVEL SECURITY;

-- Create additional functions for examples
-- ===========================================

-- Function to generate test data for batch operations
CREATE OR REPLACE FUNCTION generate_batch_test_data(batch_size INTEGER DEFAULT 100)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
BEGIN
    FOR i IN 1..batch_size LOOP
        INSERT INTO products (name, price, category) VALUES 
        (
            'Batch Product ' || i,
            ROUND((random() * 1000 + 10)::numeric, 2),
            CASE (i % 5)
                WHEN 0 THEN 'Electronics'
                WHEN 1 THEN 'Office'
                WHEN 2 THEN 'Furniture'
                WHEN 3 THEN 'Books'
                ELSE 'Accessories'
            END
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to generate test data for copy operations
CREATE OR REPLACE FUNCTION generate_copy_test_data(record_count INTEGER DEFAULT 1000)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    departments TEXT[] := ARRAY['Engineering', 'Marketing', 'Sales', 'HR', 'Finance', 'Operations'];
BEGIN
    FOR i IN 1..record_count LOOP
        INSERT INTO copy_test (name, email, age, salary, department, hire_date, active) VALUES 
        (
            'Employee ' || i,
            'employee' || i || '@company.com',
            25 + (i % 20), -- Age between 25-44
            40000 + (i % 100) * 100, -- Salary between 40k-50k
            departments[1 + (i % array_length(departments, 1))],
            CURRENT_DATE - (i % 1000) * INTERVAL '1 day',
            (i % 10) != 0 -- 90% active
        );
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to simulate account transactions for examples
CREATE OR REPLACE FUNCTION simulate_account_transactions(transaction_count INTEGER DEFAULT 100)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    from_account_id INTEGER;
    to_account_id INTEGER;
    amount DECIMAL(10,2);
    from_balance DECIMAL(10,2);
BEGIN
    FOR i IN 1..transaction_count LOOP
        -- Select random accounts
        SELECT id INTO from_account_id FROM accounts ORDER BY random() LIMIT 1;
        SELECT id INTO to_account_id FROM accounts WHERE id != from_account_id ORDER BY random() LIMIT 1;
        
        -- Get current balance
        SELECT balance INTO from_balance FROM accounts WHERE id = from_account_id;
        
        -- Generate random amount (up to 50% of balance)
        amount := ROUND((random() * LEAST(from_balance * 0.5, 100))::numeric, 2);
        
        -- Only proceed if amount is positive and we have sufficient balance
        IF amount > 0 AND from_balance >= amount THEN
            -- Debit from source account
            UPDATE accounts SET balance = balance - amount WHERE id = from_account_id;
            
            -- Credit to destination account
            UPDATE accounts SET balance = balance + amount WHERE id = to_account_id;
            
            -- Log the transaction
            INSERT INTO audit_log (table_name, operation, new_values, user_id) VALUES
            (
                'accounts',
                'TRANSFER',
                json_build_object(
                    'from_account', from_account_id,
                    'to_account', to_account_id,
                    'amount', amount,
                    'transaction_id', i
                ),
                'system'
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to populate chat channels for LISTEN/NOTIFY examples
CREATE OR REPLACE FUNCTION populate_chat_channels()
RETURNS VOID AS $$
DECLARE
    channels TEXT[] := ARRAY['general', 'tech', 'support', 'random', 'notifications', 'alerts'];
    users TEXT[] := ARRAY['admin', 'user1', 'user2', 'developer1', 'developer2', 'support1', 'customer1'];
    messages TEXT[] := ARRAY[
        'Hello everyone!',
        'How is everyone doing?',
        'Any updates?',
        'Working on new features',
        'Code review completed',
        'Need help with something',
        'System is running smoothly',
        'Performance looks good',
        'Tests are passing',
        'Deployment successful'
    ];
    i INTEGER;
    channel TEXT;
    username TEXT;
    message TEXT;
BEGIN
    FOR i IN 1..50 LOOP
        channel := channels[1 + (i % array_length(channels, 1))];
        username := users[1 + (i % array_length(users, 1))];
        message := messages[1 + (i % array_length(messages, 1))];
        
        INSERT INTO chat_messages (channel, username, message) VALUES (channel, username, message);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to create tenant-specific test data
CREATE OR REPLACE FUNCTION setup_tenant_test_data()
RETURNS VOID AS $$
DECLARE
    tenant_record RECORD;
    i INTEGER;
BEGIN
    -- Add more users to each tenant schema
    FOR tenant_record IN SELECT schema_name FROM tenants WHERE active = true LOOP
        FOR i IN 1..5 LOOP
            EXECUTE format('
                INSERT INTO %I.users (name, email) VALUES 
                (''Test User %s'', ''testuser%s@%s.com'')
                ON CONFLICT (email) DO NOTHING',
                tenant_record.schema_name,
                i,
                i,
                replace(tenant_record.schema_name, 'tenant_', '')
            );
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- EXAMPLE UTILITY FUNCTIONS
-- ===========================================

-- Function to reset all example data
CREATE OR REPLACE FUNCTION reset_example_data()
RETURNS VOID AS $$
BEGIN
    -- Clear all tables
    TRUNCATE TABLE audit_log RESTART IDENTITY CASCADE;
    TRUNCATE TABLE chat_messages RESTART IDENTITY CASCADE;
    TRUNCATE TABLE monitored_table RESTART IDENTITY CASCADE;
    TRUNCATE TABLE replica_test RESTART IDENTITY CASCADE;
    TRUNCATE TABLE performance_test RESTART IDENTITY CASCADE;
    TRUNCATE TABLE test_concurrent RESTART IDENTITY CASCADE;
    TRUNCATE TABLE test_transactions RESTART IDENTITY CASCADE;
    TRUNCATE TABLE shared_users RESTART IDENTITY CASCADE;
    TRUNCATE TABLE copy_test RESTART IDENTITY CASCADE;
    TRUNCATE TABLE products RESTART IDENTITY CASCADE;
    TRUNCATE TABLE accounts RESTART IDENTITY CASCADE;
    
    -- Clear tenant schema data
    TRUNCATE TABLE tenant_empresa_a.users RESTART IDENTITY CASCADE;
    TRUNCATE TABLE tenant_empresa_b.users RESTART IDENTITY CASCADE;
    TRUNCATE TABLE tenant_empresa_c.users RESTART IDENTITY CASCADE;
    
    -- Clear tenants (this will be recreated by sample data)
    TRUNCATE TABLE tenants RESTART IDENTITY CASCADE;
    
    RAISE NOTICE 'All example data has been reset';
END;
$$ LANGUAGE plpgsql;

-- Function to get table statistics for examples
CREATE OR REPLACE FUNCTION get_example_table_stats()
RETURNS TABLE(
    table_name TEXT,
    record_count BIGINT,
    table_size TEXT,
    last_updated TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.table_name::TEXT,
        t.record_count,
        pg_size_pretty(pg_total_relation_size(t.table_name::regclass))::TEXT as table_size,
        CURRENT_TIMESTAMP as last_updated
    FROM (
        SELECT 'products' as table_name, COUNT(*) as record_count FROM products
        UNION ALL
        SELECT 'accounts' as table_name, COUNT(*) as record_count FROM accounts
        UNION ALL
        SELECT 'copy_test' as table_name, COUNT(*) as record_count FROM copy_test
        UNION ALL
        SELECT 'tenants' as table_name, COUNT(*) as record_count FROM tenants
        UNION ALL
        SELECT 'shared_users' as table_name, COUNT(*) as record_count FROM shared_users
        UNION ALL
        SELECT 'chat_messages' as table_name, COUNT(*) as record_count FROM chat_messages
        UNION ALL
        SELECT 'monitored_table' as table_name, COUNT(*) as record_count FROM monitored_table
        UNION ALL
        SELECT 'replica_test' as table_name, COUNT(*) as record_count FROM replica_test
        UNION ALL
        SELECT 'performance_test' as table_name, COUNT(*) as record_count FROM performance_test
        UNION ALL
        SELECT 'audit_log' as table_name, COUNT(*) as record_count FROM audit_log
    ) t
    ORDER BY t.table_name;
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- EXAMPLE-SPECIFIC INDEXES
-- ===========================================

-- Additional indexes for performance testing
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_category ON products(name, category);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_copy_test_dept_salary ON copy_test(department, salary);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chat_messages_composite ON chat_messages(channel, created_at, username);

-- Partial indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_active_products ON products(category) WHERE name IS NOT NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_active_copy_test ON copy_test(department) WHERE active = true;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_active_tenants ON tenants(name) WHERE active = true;

-- ===========================================
-- HELPER VIEWS FOR EXAMPLES
-- ===========================================

-- View for batch operation examples
CREATE OR REPLACE VIEW batch_operation_summary AS
SELECT 
    category,
    COUNT(*) as product_count,
    AVG(price) as avg_price,
    SUM(price) as total_value,
    MIN(created_at) as first_created,
    MAX(created_at) as last_created
FROM products
GROUP BY category
ORDER BY product_count DESC;

-- View for multi-tenant examples
CREATE OR REPLACE VIEW multi_tenant_summary AS
SELECT 
    t.name as tenant_name,
    t.schema_name,
    t.active,
    COUNT(su.id) as shared_users_count,
    (
        SELECT COUNT(*)
        FROM information_schema.tables ist
        WHERE ist.table_schema = t.schema_name
        AND ist.table_name = 'users'
    ) as has_schema_table
FROM tenants t
LEFT JOIN shared_users su ON t.id = su.tenant_id
GROUP BY t.id, t.name, t.schema_name, t.active
ORDER BY t.name;

-- View for performance testing
CREATE OR REPLACE VIEW performance_metrics AS
SELECT 
    schemaname,
    relname as table_name,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    n_tup_hot_upd as hot_updates,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables
WHERE schemaname NOT IN ('information_schema', 'pg_catalog')
ORDER BY schemaname, relname;

-- ===========================================
-- FINALIZATION MESSAGE
-- ===========================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'NEXS-LIB Examples Setup Complete!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Available utility functions:';
    RAISE NOTICE '- generate_batch_test_data(batch_size)';
    RAISE NOTICE '- generate_copy_test_data(record_count)';
    RAISE NOTICE '- simulate_account_transactions(count)';
    RAISE NOTICE '- populate_chat_channels()';
    RAISE NOTICE '- setup_tenant_test_data()';
    RAISE NOTICE '- reset_example_data()';
    RAISE NOTICE '- get_example_table_stats()';
    RAISE NOTICE '========================================';
END $$;
