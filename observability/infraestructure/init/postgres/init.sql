-- Nexs Observability Test Database initialization

-- Create schemas for different test scenarios
CREATE SCHEMA IF NOT EXISTS nexs_tracer_tests;
CREATE SCHEMA IF NOT EXISTS nexs_logger_tests;

-- Create tables for tracer tests
CREATE TABLE IF NOT EXISTS nexs_tracer_tests.test_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS nexs_tracer_tests.test_orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES nexs_tracer_tests.test_users(id),
    order_id VARCHAR(50) UNIQUE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create tables for logger tests
CREATE TABLE IF NOT EXISTS nexs_logger_tests.test_logs (
    id SERIAL PRIMARY KEY,
    level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL,
    trace_id VARCHAR(64),
    span_id VARCHAR(32),
    service_name VARCHAR(100),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_test_users_email ON nexs_tracer_tests.test_users(email);
CREATE INDEX IF NOT EXISTS idx_test_orders_user_id ON nexs_tracer_tests.test_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_test_orders_order_id ON nexs_tracer_tests.test_orders(order_id);
CREATE INDEX IF NOT EXISTS idx_test_logs_level ON nexs_logger_tests.test_logs(level);
CREATE INDEX IF NOT EXISTS idx_test_logs_trace_id ON nexs_logger_tests.test_logs(trace_id);
CREATE INDEX IF NOT EXISTS idx_test_logs_timestamp ON nexs_logger_tests.test_logs(timestamp);

-- Insert sample data for testing
INSERT INTO nexs_tracer_tests.test_users (name, email) VALUES 
    ('João Silva', 'joao.silva@example.com'),
    ('Maria Santos', 'maria.santos@example.com'),
    ('Pedro Oliveira', 'pedro.oliveira@example.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO nexs_tracer_tests.test_orders (user_id, order_id, amount, payment_method) VALUES 
    (1, 'ORD-001', 149.99, 'credit_card'),
    (1, 'ORD-002', 299.50, 'pix'),
    (2, 'ORD-003', 89.90, 'debit_card'),
    (3, 'ORD-004', 199.99, 'credit_card')
ON CONFLICT (order_id) DO NOTHING;

-- Grant permissions
GRANT ALL PRIVILEGES ON SCHEMA nexs_tracer_tests TO nexs;
GRANT ALL PRIVILEGES ON SCHEMA nexs_logger_tests TO nexs;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA nexs_tracer_tests TO nexs;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA nexs_logger_tests TO nexs;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA nexs_tracer_tests TO nexs;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA nexs_logger_tests TO nexs;
