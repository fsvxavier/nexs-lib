#!/bin/bash

# Super Simple PostgreSQL Replica Setup Script
set -e

PRIMARY_HOST=${PRIMARY_HOST:-"postgres-primary"}
REPLICATION_USER=${REPLICATION_USER:-"replicator"}
REPLICATION_PASSWORD=${REPLICATION_PASSWORD:-"replicator_password"}
REPLICA_NAME=${REPLICA_NAME:-"replica"}

echo "[$(date)] [$REPLICA_NAME] Starting replica setup..."

# Wait for primary to be ready
echo "[$(date)] [$REPLICA_NAME] Waiting for primary..."
until pg_isready -h "$PRIMARY_HOST" -p 5432 -U nexs_user -d nexs_testdb; do
    echo "[$(date)] [$REPLICA_NAME] Primary not ready, waiting..."
    sleep 3
done

echo "[$(date)] [$REPLICA_NAME] Primary is ready!"

# Create fresh base backup
echo "[$(date)] [$REPLICA_NAME] Creating fresh base backup..."
rm -rf /var/lib/postgresql/data/*

export PGPASSWORD="$REPLICATION_PASSWORD"
pg_basebackup -h "$PRIMARY_HOST" -p 5432 -D /var/lib/postgresql/data -U "$REPLICATION_USER" -v -P -R

chown -R postgres:postgres /var/lib/postgresql/data
chmod 700 /var/lib/postgresql/data

echo "[$(date)] [$REPLICA_NAME] Base backup completed successfully"

# Start PostgreSQL without custom config files
echo "[$(date)] [$REPLICA_NAME] Starting PostgreSQL..."
exec postgres
