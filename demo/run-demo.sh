#!/bin/bash

# R2S Demo Runner Script
# This script starts all demo services and opens the frontend

echo "Starting R2S Demo Environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f "../.env" ]; then
    echo -e "${YELLOW}Warning: .env file not found. Creating default configuration...${NC}"
    cat > ../.env << EOF
DATABASE_URL=postgresql://postgres:password@localhost:5432/r2s_dev?sslmode=disable
DEMO_PORT=3008
TX_HELPER_PORT=3006
BLOCKCHAIN_RPC_URL=https://public-en.node.kaia.io
CAMPAIGN_FACTORY_ADDRESS=0x1234567890123456789012345678901234567890
USDT_ADDRESS=0x0987654321098765432109876543210987654321
EOF
    echo -e "${GREEN}Created default .env file${NC}"
fi

# Function to check if port is in use
check_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null ; then
        echo -e "${RED}Port $1 is already in use${NC}"
        return 1
    fi
    return 0
}

# Function to kill process on port
kill_port() {
    if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null ; then
        echo "Killing process on port $1..."
        kill -9 $(lsof -t -i:$1)
        sleep 1
    fi
}

# Function to start a service
start_service() {
    local name=$1
    local port=$2
    local cmd=$3
    local dir=$4
    
    echo -e "${YELLOW}Starting $name on port $port...${NC}"
    
    # Kill existing process if any
    kill_port $port
    
    # Start the service
    cd $dir
    $cmd > /tmp/${name}.log 2>&1 &
    local pid=$!
    cd - > /dev/null
    
    # Wait for service to start
    sleep 2
    
    # Check if service started successfully
    if ps -p $pid > /dev/null; then
        echo -e "${GREEN}$name started (PID: $pid)${NC}"
        echo $pid >> /tmp/r2s_demo.pids
    else
        echo -e "${RED}Failed to start $name${NC}"
        echo "Check logs at /tmp/${name}.log"
        return 1
    fi
}

# Clean up function
cleanup() {
    echo -e "\n${YELLOW}Shutting down demo services...${NC}"
    
    if [ -f /tmp/r2s_demo.pids ]; then
        while read pid; do
            if ps -p $pid > /dev/null 2>&1; then
                kill -9 $pid 2>/dev/null
            fi
        done < /tmp/r2s_demo.pids
        rm /tmp/r2s_demo.pids
    fi
    
    echo -e "${GREEN}Demo services stopped${NC}"
    exit 0
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Clear previous PIDs file
rm -f /tmp/r2s_demo.pids

# Step 1: Check PostgreSQL
echo "Checking PostgreSQL connection..."
if ! PGPASSWORD=password psql -h localhost -U postgres -d r2s_dev -c '\q' 2>/dev/null; then
    echo -e "${RED}PostgreSQL is not running or database r2s_dev doesn't exist${NC}"
    echo "Please ensure PostgreSQL is running and create the database:"
    echo "  createdb r2s_dev"
    exit 1
fi
echo -e "${GREEN}PostgreSQL connected${NC}"

# Step 2: Seed demo data
echo -e "\n${YELLOW}Seeding demo data...${NC}"
cd demo
if go run seed.go; then
    echo -e "${GREEN}Demo data seeded${NC}"
else
    echo -e "${RED}Failed to seed demo data${NC}"
    exit 1
fi
cd ..

# Step 3: Start TX Helper
start_service "TX-Helper" 3006 "go run main.go" "../tx-helper"

# Step 4: Start Demo API
start_service "Demo-API" 3008 "go run demo_api.go" "demo"

# Step 5: Display status
echo -e "\n${GREEN}═══════════════════════════════════════════════════${NC}"
echo -e "${GREEN}       R2S Demo Environment Started Successfully!    ${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════${NC}"
echo ""
echo "Services Running:"
echo "   • Demo API:    http://localhost:3008"
echo "   • TX Helper:   http://localhost:3006"
echo ""
echo "Frontend Demo:"
echo "   Open demo/frontend-integration.html in your browser"
echo ""
echo "Quick Test Commands:"
echo "   curl http://localhost:3008/demo/users"
echo "   curl http://localhost:3008/demo/campaigns"
echo "   curl http://localhost:3008/demo/stats"
echo ""
echo "Logs:"
echo "   • Demo API:  tail -f /tmp/Demo-API.log"
echo "   • TX Helper: tail -f /tmp/TX-Helper.log"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"

# Step 6: Open browser (optional)
if command -v open &> /dev/null; then
    echo -e "\n${YELLOW}Opening frontend demo in browser...${NC}"
    sleep 2
    open "frontend-integration.html"
elif command -v xdg-open &> /dev/null; then
    echo -e "\n${YELLOW}Opening frontend demo in browser...${NC}"
    sleep 2
    xdg-open "frontend-integration.html"
fi

# Keep script running
echo -e "\n${GREEN}Demo environment is running. Press Ctrl+C to stop.${NC}"
while true; do
    sleep 1
done