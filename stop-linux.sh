#!/bin/bash

#############################################
# Pack Calculator - Linux Stop Script
# Stops backend and frontend processes
#############################################

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PID_DIR="./pids"
BACKEND_PID="$PID_DIR/backend.pid"
FRONTEND_PID="$PID_DIR/frontend.pid"

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo -e "${BLUE}"
echo "================================================"
echo "  Pack Calculator - Stopping Services"
echo "================================================"
echo -e "${NC}"

# Stop backend
if [ -f "$BACKEND_PID" ]; then
    BACKEND_PID_VALUE=$(cat $BACKEND_PID)
    if ps -p $BACKEND_PID_VALUE > /dev/null 2>&1; then
        log_info "Stopping backend (PID: $BACKEND_PID_VALUE)..."
        kill $BACKEND_PID_VALUE
        sleep 2
        
        # Force kill if still running
        if ps -p $BACKEND_PID_VALUE > /dev/null 2>&1; then
            log_warning "Backend did not stop gracefully. Force killing..."
            kill -9 $BACKEND_PID_VALUE
        fi
        
        log_info "✓ Backend stopped"
    else
        log_warning "Backend process not running"
    fi
    rm -f $BACKEND_PID
else
    log_warning "Backend PID file not found"
fi

# Stop frontend
if [ -f "$FRONTEND_PID" ]; then
    FRONTEND_PID_VALUE=$(cat $FRONTEND_PID)
    if ps -p $FRONTEND_PID_VALUE > /dev/null 2>&1; then
        log_info "Stopping frontend (PID: $FRONTEND_PID_VALUE)..."
        kill $FRONTEND_PID_VALUE
        sleep 2
        
        # Force kill if still running
        if ps -p $FRONTEND_PID_VALUE > /dev/null 2>&1; then
            log_warning "Frontend did not stop gracefully. Force killing..."
            kill -9 $FRONTEND_PID_VALUE
        fi
        
        log_info "✓ Frontend stopped"
    else
        log_warning "Frontend process not running"
    fi
    rm -f $FRONTEND_PID
else
    log_warning "Frontend PID file not found"
fi

# Kill any remaining processes on the ports
log_info "Checking for remaining processes..."

# Check backend port
BACKEND_PORT=8080
if lsof -Pi :$BACKEND_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    log_warning "Found process on port $BACKEND_PORT. Killing..."
    lsof -ti:$BACKEND_PORT | xargs kill -9 2>/dev/null
fi

# Check frontend port
FRONTEND_PORT=3000
if lsof -Pi :$FRONTEND_PORT -sTCP:LISTEN -t >/dev/null 2>&1; then
    log_warning "Found process on port $FRONTEND_PORT. Killing..."
    lsof -ti:$FRONTEND_PORT | xargs kill -9 2>/dev/null
fi

echo ""
echo -e "${GREEN}================================================"
echo "  ✓ All services stopped"
echo "================================================${NC}"
echo ""

