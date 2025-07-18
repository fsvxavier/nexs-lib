-- Database schema for nexs-lib testing and examples
-- This script creates tables, indexes, and sample data based on the examples

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto extension for password hashing
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ===========================================
-- CORE TABLES FOR EXAMPLES
-- ===========================================

-- Products table for batch operations, copy examples, and transactions
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Accounts table for transaction examples
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Copy test table for COPY operations examples
CREATE TABLE IF NOT EXISTS copy_test (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    age INTEGER NOT NULL,
    salary DECIMAL(10,2) NOT NULL,
    department VARCHAR(50) NOT NULL,
    hire_date DATE NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Replica test table for read replica examples
CREATE TABLE IF NOT EXISTS replica_test (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- MULTI-TENANCY TABLES
-- ===========================================

-- Tenants table for multi-tenancy examples
CREATE TABLE IF NOT EXISTS tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    schema_name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    active BOOLEAN DEFAULT true,
    settings JSONB DEFAULT '{}'
);

-- Shared users table for row-level multi-tenancy
CREATE TABLE IF NOT EXISTS shared_users (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

-- ===========================================
-- LISTEN/NOTIFY TABLES
-- ===========================================

-- Monitored table for LISTEN/NOTIFY examples
CREATE TABLE IF NOT EXISTS monitored_table (
    id SERIAL PRIMARY KEY,
    data VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Chat messages table for LISTEN/NOTIFY chat example
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    channel VARCHAR(50) NOT NULL,
    username VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ===========================================
-- ADDITIONAL TESTING TABLES
-- ===========================================

-- Test transactions table for advanced examples
CREATE TABLE IF NOT EXISTS test_transactions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    amount DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Test concurrent table for advanced examples
CREATE TABLE IF NOT EXISTS test_concurrent (
    id SERIAL PRIMARY KEY,
    worker_id INT,
    task_id INT,
    completed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Audit log table for hooks and monitoring examples
CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(100) NOT NULL,
    operation VARCHAR(10) NOT NULL, -- INSERT, UPDATE, DELETE
    old_values JSONB,
    new_values JSONB,
    user_id VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(255),
    ip_address INET
);

-- Performance test table for load testing
CREATE TABLE IF NOT EXISTS performance_test (
    id SERIAL PRIMARY KEY,
    test_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- MULTI-TENANT SCHEMAS
-- ===========================================

-- Create schemas for schema-based multi-tenancy examples
CREATE SCHEMA IF NOT EXISTS tenant_empresa_a;
CREATE SCHEMA IF NOT EXISTS tenant_empresa_b;
CREATE SCHEMA IF NOT EXISTS tenant_empresa_c;

-- Create users tables in tenant schemas
CREATE TABLE IF NOT EXISTS tenant_empresa_a.users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tenant_empresa_b.users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tenant_empresa_c.users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ===========================================
-- INDEXES FOR PERFORMANCE
-- ===========================================

-- Core table indexes
CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

CREATE INDEX IF NOT EXISTS idx_accounts_name ON accounts(name);
CREATE INDEX IF NOT EXISTS idx_accounts_balance ON accounts(balance);

CREATE INDEX IF NOT EXISTS idx_copy_test_email ON copy_test(email);
CREATE INDEX IF NOT EXISTS idx_copy_test_department ON copy_test(department);
CREATE INDEX IF NOT EXISTS idx_copy_test_hire_date ON copy_test(hire_date);

-- Multi-tenancy indexes
CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_schema_name ON tenants(schema_name);
CREATE INDEX IF NOT EXISTS idx_tenants_active ON tenants(active);

CREATE INDEX IF NOT EXISTS idx_shared_users_tenant_id ON shared_users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_shared_users_email ON shared_users(email);

-- Listen/Notify indexes
CREATE INDEX IF NOT EXISTS idx_monitored_table_created_at ON monitored_table(created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_channel ON chat_messages(channel);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);

-- Audit and testing indexes
CREATE INDEX IF NOT EXISTS idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_performance_test_created_at ON performance_test(created_at);

-- Composite indexes for complex queries
CREATE INDEX IF NOT EXISTS idx_shared_users_tenant_email ON shared_users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_chat_messages_channel_time ON chat_messages(channel, created_at);

-- ===========================================
-- FUNCTIONS AND TRIGGERS
-- ===========================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for automatic updated_at updates
CREATE TRIGGER update_accounts_updated_at 
    BEFORE UPDATE ON accounts 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function for audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, operation, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, operation, old_values, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, operation, old_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD));
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Audit triggers for main tables
CREATE TRIGGER audit_products_trigger 
    AFTER INSERT OR UPDATE OR DELETE ON products 
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

CREATE TRIGGER audit_accounts_trigger 
    AFTER INSERT OR UPDATE OR DELETE ON accounts 
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- ===========================================
-- LISTEN/NOTIFY FUNCTIONS
-- ===========================================

-- Function to notify data changes
CREATE OR REPLACE FUNCTION notify_data_change()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        PERFORM pg_notify('data_change', json_build_object(
            'table', TG_TABLE_NAME,
            'operation', TG_OP,
            'id', NEW.id,
            'timestamp', CURRENT_TIMESTAMP
        )::text);
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        PERFORM pg_notify('data_change', json_build_object(
            'table', TG_TABLE_NAME,
            'operation', TG_OP,
            'id', NEW.id,
            'timestamp', CURRENT_TIMESTAMP
        )::text);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        PERFORM pg_notify('data_change', json_build_object(
            'table', TG_TABLE_NAME,
            'operation', TG_OP,
            'id', OLD.id,
            'timestamp', CURRENT_TIMESTAMP
        )::text);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Triggers for LISTEN/NOTIFY examples
CREATE TRIGGER notify_monitored_table_trigger 
    AFTER INSERT OR UPDATE OR DELETE ON monitored_table 
    FOR EACH ROW EXECUTE FUNCTION notify_data_change();

-- Function to notify new chat messages
CREATE OR REPLACE FUNCTION notify_new_chat_message()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_message', json_build_object(
        'channel', NEW.channel,
        'username', NEW.username,
        'message', NEW.message,
        'timestamp', NEW.created_at
    )::text);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER notify_new_chat_message_trigger 
    AFTER INSERT ON chat_messages 
    FOR EACH ROW EXECUTE FUNCTION notify_new_chat_message();

-- ===========================================
-- ROW LEVEL SECURITY
-- ===========================================

-- Enable RLS on shared_users table for multi-tenancy
ALTER TABLE shared_users ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant isolation
CREATE POLICY tenant_isolation ON shared_users
    FOR ALL
    TO PUBLIC
    USING (tenant_id = COALESCE(current_setting('app.current_tenant_id', true)::integer, tenant_id));

-- ===========================================
-- VIEWS FOR TESTING
-- ===========================================

-- View for product statistics
CREATE OR REPLACE VIEW product_stats AS
SELECT 
    category,
    COUNT(*) as product_count,
    AVG(price) as avg_price,
    MIN(price) as min_price,
    MAX(price) as max_price,
    SUM(price) as total_value
FROM products
GROUP BY category;

-- View for account balances
CREATE OR REPLACE VIEW account_summary AS
SELECT 
    name,
    balance,
    CASE 
        WHEN balance < 0 THEN 'Negative'
        WHEN balance = 0 THEN 'Zero'
        ELSE 'Positive'
    END as balance_status,
    created_at
FROM accounts
ORDER BY balance DESC;

-- View for tenant statistics
CREATE OR REPLACE VIEW tenant_stats AS
SELECT 
    t.name as tenant_name,
    t.schema_name,
    t.active,
    COUNT(su.id) as user_count,
    t.created_at as tenant_created_at
FROM tenants t
LEFT JOIN shared_users su ON t.id = su.tenant_id
GROUP BY t.id, t.name, t.schema_name, t.active, t.created_at
ORDER BY t.name;

-- Performance test table for load testing
CREATE TABLE IF NOT EXISTS performance_test (
    id SERIAL PRIMARY KEY,
    test_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Multi-tenant tables in tenant schemas
CREATE TABLE IF NOT EXISTS tenant_1.tenant_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(50) NOT NULL DEFAULT 'tenant_1',
    data_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tenant_2.tenant_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(50) NOT NULL DEFAULT 'tenant_2',
    data_value TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_table_name ON audit_log(table_name);
CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_performance_test_created_at ON performance_test(created_at);

-- Composite indexes for complex queries
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_product ON order_items(order_id, product_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for automatic updated_at updates
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function for audit logging
CREATE OR REPLACE FUNCTION audit_trigger_function()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (table_name, operation, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (table_name, operation, old_values, new_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO audit_log (table_name, operation, old_values)
        VALUES (TG_TABLE_NAME, TG_OP, row_to_json(OLD));
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

-- Audit triggers for main tables
CREATE TRIGGER audit_users_trigger AFTER INSERT OR UPDATE OR DELETE ON users FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
CREATE TRIGGER audit_products_trigger AFTER INSERT OR UPDATE OR DELETE ON products FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();
CREATE TRIGGER audit_orders_trigger AFTER INSERT OR UPDATE OR DELETE ON orders FOR EACH ROW EXECUTE FUNCTION audit_trigger_function();

-- Notification function for LISTEN/NOTIFY examples
CREATE OR REPLACE FUNCTION notify_new_order()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('new_order', json_build_object(
        'order_id', NEW.id,
        'user_id', NEW.user_id,
        'total_amount', NEW.total_amount,
        'status', NEW.status
    )::text);
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER notify_new_order_trigger AFTER INSERT ON orders FOR EACH ROW EXECUTE FUNCTION notify_new_order();

-- Views for testing complex queries
CREATE OR REPLACE VIEW user_orders_summary AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    COUNT(o.id) as total_orders,
    COALESCE(SUM(o.total_amount), 0) as total_spent,
    MAX(o.created_at) as last_order_date
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.username, u.email;

CREATE OR REPLACE VIEW product_sales_summary AS
SELECT 
    p.id as product_id,
    p.name,
    p.category,
    COALESCE(SUM(oi.quantity), 0) as total_sold,
    COALESCE(SUM(oi.total_price), 0) as total_revenue,
    COUNT(DISTINCT oi.order_id) as total_orders
FROM products p
LEFT JOIN order_items oi ON p.id = oi.product_id
GROUP BY p.id, p.name, p.category;
