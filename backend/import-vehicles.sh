#!/bin/bash

# Script to import vehicle data into the database
# Usage: ./import-vehicles.sh <path-to-sql-file> [database-url]

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Vehicle Data Import Script${NC}\n"

# Check if SQL file path is provided
if [ -z "$1" ]; then
    echo -e "${RED}✗ Error: SQL file path is required${NC}"
    echo -e "${YELLOW}Usage: ./import-vehicles.sh <path-to-sql-file> [database-url]${NC}"
    echo -e "${YELLOW}Example: ./import-vehicles.sh ../data/vehicles.sql${NC}"
    echo -e "${YELLOW}Or with explicit DB URL: ./import-vehicles.sh ../data/vehicles.sql postgres://user:pass@host/db${NC}"
    exit 1
fi

SQL_FILE="$1"
DATABASE_URL="${2:-$DATABASE_URL}"

# If SQL file is relative and not found, check container location
if [ ! -f "$SQL_FILE" ] && [ ! -f "/root/$SQL_FILE" ]; then
    # Try container location
    if [ -f "/root/car-db.sql" ] && [ "$SQL_FILE" = "car-db.sql" ]; then
        SQL_FILE="/root/car-db.sql"
    else
        echo -e "${RED}✗ Error: SQL file not found: $SQL_FILE${NC}"
        exit 1
    fi
elif [ ! -f "$SQL_FILE" ] && [ -f "/root/$SQL_FILE" ]; then
    SQL_FILE="/root/$SQL_FILE"
fi

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}✗ Error: DATABASE_URL environment variable is not set${NC}"
    echo -e "${YELLOW}Either set DATABASE_URL environment variable or pass it as second argument${NC}"
    echo -e "${YELLOW}Example: DATABASE_URL=postgres://... ./import-vehicles.sh $SQL_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}✓ SQL file found: $SQL_FILE${NC}"
echo -e "${GREEN}✓ Database URL configured${NC}"
echo ""

# Check if binary exists (container) or needs to be built (local)
IMPORT_BINARY=""
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Check for binary in container location first
if [ -f "/root/bin/import-vehicles" ]; then
    IMPORT_BINARY="/root/bin/import-vehicles"
    echo -e "${GREEN}✓ Using pre-built binary from container${NC}"
elif [ -f "$SCRIPT_DIR/bin/import-vehicles" ]; then
    IMPORT_BINARY="$SCRIPT_DIR/bin/import-vehicles"
    echo -e "${GREEN}✓ Using existing binary${NC}"
else
    # Build the import-vehicles binary
    echo -e "${BLUE}Building import-vehicles binary...${NC}"
    cd "$SCRIPT_DIR"
    
    if ! go build -o bin/import-vehicles cmd/import-vehicles/main.go; then
        echo -e "${RED}✗ Failed to build import-vehicles${NC}"
        exit 1
    fi
    
    IMPORT_BINARY="$SCRIPT_DIR/bin/import-vehicles"
    echo -e "${GREEN}✓ Build successful${NC}"
fi

echo ""

# Run the import
echo -e "${BLUE}Starting import...${NC}"
echo -e "${YELLOW}This may take a few minutes depending on the size of your SQL file${NC}"
echo ""

if "$IMPORT_BINARY" "$SQL_FILE" "$DATABASE_URL"; then
    echo ""
    echo -e "${GREEN}✓ Import completed successfully!${NC}"
    echo -e "${GREEN}Vehicle makes, models, and trims are now available in the database${NC}"
else
    echo ""
    echo -e "${RED}✗ Import failed${NC}"
    exit 1
fi

