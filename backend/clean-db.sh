#!/bin/bash

# Clean Database Script Runner
# This script runs the clean-db.sql script to remove user data while preserving reference data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get database connection details from environment or use defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-carbuyer}"
DB_USER="${DB_USER:-calvinkorver}"

# Check if DATABASE_URL is set (common in production)
if [ -n "$DATABASE_URL" ]; then
    echo -e "${YELLOW}Using DATABASE_URL from environment${NC}"
    psql "$DATABASE_URL" -f clean-db.sql
else
    # Prompt for password if not in environment
    if [ -z "$DB_PASSWORD" ]; then
        echo -e "${YELLOW}Database password not set in DB_PASSWORD environment variable${NC}"
        echo -e "${YELLOW}You will be prompted for the password${NC}"
    fi
    
    # Construct connection string
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    echo -e "${GREEN}Cleaning database: ${DB_NAME} on ${DB_HOST}:${DB_PORT}${NC}"
    echo -e "${YELLOW}This will delete conversation data (threads, messages, offers, dealers)${NC}"
    echo -e "${YELLOW}User accounts and preferences will be preserved${NC}"
    echo -e "${YELLOW}Reference data (makes, models, message_types, vehicle_trims) will be preserved${NC}"
    echo ""
    read -p "Are you sure you want to continue? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        echo -e "${RED}Aborted${NC}"
        exit 1
    fi
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f clean-db.sql
    
    # Unset password if we set it
    unset PGPASSWORD
fi

echo -e "${GREEN}Database cleanup completed!${NC}"

