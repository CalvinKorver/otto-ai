#!/bin/bash

# Development startup script for car-buyer
# Starts both frontend and backend servers

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting car-buyer development environment...${NC}\n"

# Function to cleanup background processes on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down servers...${NC}"
    kill $(jobs -p) 2>/dev/null
    exit
}

trap cleanup SIGINT SIGTERM

# Start backend
echo -e "${GREEN}Starting backend server...${NC}"
cd backend

# Test backend compilation first
if ! go build -o /dev/null cmd/server/main.go 2>&1; then
    echo -e "${RED}✗ Backend compilation failed!${NC}"
    echo -e "${YELLOW}Fix the errors above and try again.${NC}"
    cd ..
    exit 1
fi

# If compilation succeeds, run the server
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# Give backend a moment to start
sleep 1

# Check if backend process is still running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}✗ Backend server failed to start!${NC}"
    exit 1
fi

# Start frontend
echo -e "${GREEN}Starting frontend server...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

echo -e "\n${GREEN}✓ Servers started!${NC}"
echo -e "${BLUE}Frontend:${NC} http://localhost:3000"
echo -e "${BLUE}Backend:${NC}  http://localhost:8080"
echo -e "\n${YELLOW}Press Ctrl+C to stop all servers${NC}\n"

# Wait for all background processes
wait
