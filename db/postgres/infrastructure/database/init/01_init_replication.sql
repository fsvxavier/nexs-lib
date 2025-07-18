-- Database initialization script for nexs-lib
-- This script creates the necessary database structure for testing and examples

-- Create replication user
CREATE USER replicator WITH REPLICATION PASSWORD 'replicator_password';

-- Create application database if it doesn't exist
-- (This is handled by POSTGRES_DB environment variable in docker-compose)

-- Connect to the application database
\c nexs_testdb;
