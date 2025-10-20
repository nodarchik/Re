#!/bin/bash

#############################################
# Pack Calculator - Linux Startup Script
# Automatically installs prerequisites and starts app
#############################################

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_PORT=8080
FRONTEND_PORT=3000
DB_PORT=5432
DB_NAME="packcalculator"
DB_USER="postgres"
DB_PASSWORD="postgres"

# Log file locations
LOG_DIR="./logs"
BACKEND_LOG="$LOG_DIR/backend.log"
FRONTEND_LOG="$LOG_DIR/frontend.log"
PID_DIR="./pids"
BACKEND_PID="$PID_DIR/backend.pid"
FRONTEND_PID="$PID_DIR/frontend.pid"

#############################################
# Functions
#############################################

print_header() {
    echo -e "${BLUE}"
    echo "================================================"
    echo "  Pack Calculator - Linux Startup"
    echo "================================================"
    echo -e "${NC}"
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    else
        log_error "Cannot detect Linux distribution"
        exit 1
    fi
}

install_postgresql() {
    log_info "Installing PostgreSQL..."
    
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y postgresql postgresql-contrib
        sudo systemctl start postgresql
        sudo systemctl enable postgresql
    elif [ "$OS" = "fedora" ] || [ "$OS" = "rhel" ] || [ "$OS" = "centos" ]; then
        sudo dnf install -y postgresql-server postgresql-contrib
        sudo postgresql-setup --initdb 2>/dev/null || true
        sudo systemctl start postgresql
        sudo systemctl enable postgresql
    elif [ "$OS" = "arch" ] || [ "$OS" = "manjaro" ]; then
        sudo pacman -Sy --noconfirm postgresql
        sudo -u postgres initdb -D /var/lib/postgres/data 2>/dev/null || true
        sudo systemctl start postgresql
        sudo systemctl enable postgresql
    else
        log_error "Unsupported distribution: $OS"
        log_error "Please install PostgreSQL manually"
        exit 1
    fi
    
    log_info "✓ PostgreSQL installed"
}

install_go() {
    log_info "Installing Go..."
    
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y golang-go
    elif [ "$OS" = "fedora" ] || [ "$OS" = "rhel" ] || [ "$OS" = "centos" ]; then
        sudo dnf install -y golang
    elif [ "$OS" = "arch" ] || [ "$OS" = "manjaro" ]; then
        sudo pacman -S --noconfirm go
    else
        # Generic installation from official source
        log_info "Installing Go from official source..."
        GO_VERSION="1.21.5"
        wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
        rm go${GO_VERSION}.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    fi
    
    log_info "✓ Go installed"
}

install_nodejs() {
    log_info "Installing Node.js..."
    
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get update
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
    elif [ "$OS" = "fedora" ] || [ "$OS" = "rhel" ] || [ "$OS" = "centos" ]; then
        curl -fsSL https://rpm.nodesource.com/setup_18.x | sudo bash -
        sudo dnf install -y nodejs
    elif [ "$OS" = "arch" ] || [ "$OS" = "manjaro" ]; then
        sudo pacman -S --noconfirm nodejs npm
    else
        # Generic installation using NodeSource
        log_info "Installing Node.js from NodeSource..."
        curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
    
    log_info "✓ Node.js installed"
}

install_build_tools() {
    log_info "Installing build tools..."
    
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get install -y curl wget git build-essential
    elif [ "$OS" = "fedora" ] || [ "$OS" = "rhel" ] || [ "$OS" = "centos" ]; then
        sudo dnf install -y curl wget git gcc-c++ make
    elif [ "$OS" = "arch" ] || [ "$OS" = "manjaro" ]; then
        sudo pacman -S --noconfirm curl wget git base-devel
    fi
    
    log_info "✓ Build tools installed"
}

check_and_install() {
    local cmd=$1
    local install_func=$2
    
    if ! command -v $cmd &> /dev/null; then
        log_warning "$cmd not found. Installing..."
        $install_func
    else
        log_info "✓ $cmd is available"
    fi
}

check_port() {
    if command -v lsof &> /dev/null; then
        if lsof -Pi :$1 -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_error "Port $1 is already in use"
            lsof -i :$1
            exit 1
        fi
    elif command -v ss &> /dev/null; then
        if ss -ltn | grep -q ":$1 "; then
            log_error "Port $1 is already in use"
            ss -ltn | grep ":$1 "
            exit 1
        fi
    fi
}

#############################################
# Start Script
#############################################

print_header

# Detect OS
log_info "Detecting operating system..."
detect_os
log_info "Detected: $OS $VERSION"

#############################################
# Install Prerequisites
#############################################

log_info "Checking and installing prerequisites..."

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    log_warning "Running as root. This is not recommended."
    log_warning "Consider running as a regular user with sudo access."
fi

# Install curl/wget first (needed for other installations)
if ! command -v curl &> /dev/null; then
    log_info "Installing curl..."
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        sudo apt-get update && sudo apt-get install -y curl
    elif [ "$OS" = "fedora" ] || [ "$OS" = "rhel" ] || [ "$OS" = "centos" ]; then
        sudo dnf install -y curl
    elif [ "$OS" = "arch" ] || [ "$OS" = "manjaro" ]; then
        sudo pacman -Sy --noconfirm curl
    fi
fi

# Check and install PostgreSQL
check_and_install "psql" "install_postgresql"

# Check and install Go
check_and_install "go" "install_go"

# Check and install Node.js
check_and_install "node" "install_nodejs"

# Install build tools if needed
if ! command -v gcc &> /dev/null || ! command -v make &> /dev/null; then
    install_build_tools
fi

log_info "✓ All prerequisites installed"

# Show versions
echo ""
log_info "Installed versions:"
go version 2>/dev/null || echo "  Go: Not found in PATH (restart shell may be needed)"
node --version 2>/dev/null || echo "  Node: Not found"
npm --version 2>/dev/null || echo "  npm: Not found"
psql --version 2>/dev/null || echo "  PostgreSQL: Not found"

# If Go is not in PATH, add it
if ! command -v go &> /dev/null; then
    if [ -d "/usr/local/go/bin" ]; then
        export PATH=$PATH:/usr/local/go/bin
        log_info "Added Go to PATH for this session"
    fi
fi

echo ""

#############################################
# Check Ports
#############################################

log_info "Checking ports..."
check_port $BACKEND_PORT
check_port $FRONTEND_PORT
log_info "✓ Ports $BACKEND_PORT and $FRONTEND_PORT are available"

#############################################
# Setup Directories
#############################################

log_info "Creating directories..."
mkdir -p $LOG_DIR
mkdir -p $PID_DIR
mkdir -p bin

#############################################
# PostgreSQL Setup
#############################################

log_info "Setting up PostgreSQL..."

# Check if PostgreSQL is running
if ! pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
    log_warning "PostgreSQL is not running. Attempting to start..."
    
    # Try different methods to start PostgreSQL
    if command -v systemctl &> /dev/null; then
        sudo systemctl start postgresql
    elif command -v service &> /dev/null; then
        sudo service postgresql start
    else
        log_error "Cannot start PostgreSQL automatically."
        log_error "Please start PostgreSQL manually and run this script again."
        exit 1
    fi
    
    # Wait for PostgreSQL to start
    log_info "Waiting for PostgreSQL to start..."
    for i in {1..30}; do
        if pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
fi

if ! pg_isready -h localhost -p $DB_PORT > /dev/null 2>&1; then
    log_error "PostgreSQL is not responding after 30 seconds"
    log_error "Try manually: sudo systemctl start postgresql"
    exit 1
fi

log_info "✓ PostgreSQL is running"

# Configure PostgreSQL authentication if needed
log_info "Configuring PostgreSQL authentication..."
PG_HBA=$(sudo -u postgres psql -t -P format=unaligned -c 'SHOW hba_file;' 2>/dev/null || echo "")
if [ -n "$PG_HBA" ] && [ -f "$PG_HBA" ]; then
    if ! sudo grep -q "local.*all.*all.*trust" "$PG_HBA" 2>/dev/null; then
        log_info "Setting up local trust authentication..."
        sudo cp "$PG_HBA" "$PG_HBA.backup"
        sudo sed -i.bak 's/local.*all.*all.*peer/local   all             all                                     trust/' "$PG_HBA"
        sudo sed -i.bak 's/host.*all.*all.*127.0.0.1\/32.*ident/host    all             all             127.0.0.1\/32            trust/' "$PG_HBA"
        sudo sed -i.bak 's/host.*all.*all.*127.0.0.1\/32.*md5/host    all             all             127.0.0.1\/32            trust/' "$PG_HBA"
        
        # Restart PostgreSQL
        if command -v systemctl &> /dev/null; then
            sudo systemctl restart postgresql
        elif command -v service &> /dev/null; then
            sudo service postgresql restart
        fi
        
        sleep 3
        log_info "✓ PostgreSQL authentication configured"
    fi
fi

# Check if database exists
if psql -U $DB_USER -h localhost -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw $DB_NAME; then
    log_info "✓ Database '$DB_NAME' already exists"
else
    log_info "Creating database '$DB_NAME'..."
    if createdb -U $DB_USER -h localhost $DB_NAME 2>/dev/null; then
        log_info "✓ Database created successfully"
    else
        log_warning "Failed to create database as current user. Trying with postgres user..."
        sudo -u postgres createdb $DB_NAME 2>/dev/null || {
            log_error "Failed to create database"
            log_error "Try manually: sudo -u postgres createdb $DB_NAME"
            exit 1
        }
        log_info "✓ Database created successfully"
    fi
fi

#############################################
# Backend Setup
#############################################

log_info "Setting up backend..."

cd backend

# Install Go dependencies with memory optimization
log_info "Installing Go dependencies..."
export GOGC=50
export GOMEMLIMIT=256MiB
go mod download

# Build backend with memory optimization
log_info "Building backend..."
# Set Go build flags to reduce memory usage
export GOGC=50  # Reduce garbage collection threshold
export GOMEMLIMIT=256MiB  # Limit memory usage
go build -ldflags="-s -w" -o ../bin/pack-api ./cmd/api

cd ..

# Set environment variables
export PORT=$BACKEND_PORT
export DB_HOST=localhost
export DB_PORT=$DB_PORT
export DB_USER=$DB_USER
export DB_PASSWORD=$DB_PASSWORD
export DB_NAME=$DB_NAME
export CACHE_SIZE=1000

# Start backend
log_info "Starting backend on port $BACKEND_PORT..."
nohup ./bin/pack-api > $BACKEND_LOG 2>&1 &
BACKEND_PID_VALUE=$!
echo $BACKEND_PID_VALUE > $BACKEND_PID

# Wait for backend to start
log_info "Waiting for backend to be ready..."
for i in {1..30}; do
    if curl -s http://localhost:$BACKEND_PORT/health > /dev/null 2>&1; then
        log_info "✓ Backend is running (PID: $BACKEND_PID_VALUE)"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Backend failed to start within 30 seconds"
        log_error "Check logs: cat $BACKEND_LOG"
        cat $BACKEND_LOG
        exit 1
    fi
    sleep 1
done

#############################################
# Frontend Setup
#############################################

log_info "Setting up frontend..."

cd frontend

# Install npm dependencies
if [ ! -d "node_modules" ]; then
    log_info "Installing npm dependencies (this may take a few minutes)..."
    npm install
else
    log_info "✓ npm dependencies already installed"
fi

# Set environment variable
export REACT_APP_API_URL=http://localhost:$BACKEND_PORT

# Create .env file
echo "REACT_APP_API_URL=http://localhost:$BACKEND_PORT" > .env

# Start frontend (development mode)
log_info "Starting frontend on port $FRONTEND_PORT..."
nohup npm start > ../$FRONTEND_LOG 2>&1 &
FRONTEND_PID_VALUE=$!
echo $FRONTEND_PID_VALUE > ../$FRONTEND_PID

cd ..

# Wait for frontend to start
log_info "Waiting for frontend to be ready..."
for i in {1..90}; do
    if curl -s http://localhost:$FRONTEND_PORT > /dev/null 2>&1; then
        log_info "✓ Frontend is running (PID: $FRONTEND_PID_VALUE)"
        break
    fi
    if [ $i -eq 90 ]; then
        log_error "Frontend failed to start within 90 seconds"
        log_error "Check logs: cat $FRONTEND_LOG"
        tail -50 $FRONTEND_LOG
        exit 1
    fi
    sleep 1
done

#############################################
# Success
#############################################

echo ""
echo -e "${GREEN}================================================"
echo "  🚀 Pack Calculator is running!"
echo "================================================${NC}"
echo ""
echo -e "${BLUE}Backend:${NC}  http://localhost:$BACKEND_PORT"
echo -e "${BLUE}Frontend:${NC} http://localhost:$FRONTEND_PORT"
echo -e "${BLUE}Health:${NC}   http://localhost:$BACKEND_PORT/health"
echo ""
echo -e "${BLUE}Logs:${NC}"
echo "  Backend:  tail -f $BACKEND_LOG"
echo "  Frontend: tail -f $FRONTEND_LOG"
echo ""
echo -e "${BLUE}PIDs:${NC}"
echo "  Backend:  $BACKEND_PID_VALUE"
echo "  Frontend: $FRONTEND_PID_VALUE"
echo ""
echo -e "${YELLOW}To stop the application:${NC}"
echo "  ./stop-linux.sh"
echo "  OR"
echo "  kill $BACKEND_PID_VALUE $FRONTEND_PID_VALUE"
echo ""

# Test backend
log_info "Testing backend..."
HEALTH=$(curl -s http://localhost:$BACKEND_PORT/health)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Backend health check passed${NC}"
    if command -v jq &> /dev/null; then
        echo "$HEALTH" | jq '.'
    else
        echo "$HEALTH"
    fi
else
    log_warning "Backend health check failed"
fi

echo ""
log_info "Opening browser..."
sleep 2

# Try to open browser
if command -v xdg-open &> /dev/null; then
    xdg-open http://localhost:$FRONTEND_PORT 2>/dev/null &
elif command -v gnome-open &> /dev/null; then
    gnome-open http://localhost:$FRONTEND_PORT 2>/dev/null &
elif command -v firefox &> /dev/null; then
    firefox http://localhost:$FRONTEND_PORT 2>/dev/null &
elif command -v google-chrome &> /dev/null; then
    google-chrome http://localhost:$FRONTEND_PORT 2>/dev/null &
elif command -v chromium &> /dev/null; then
    chromium http://localhost:$FRONTEND_PORT 2>/dev/null &
else
    log_info "Please open your browser and go to: http://localhost:$FRONTEND_PORT"
fi

echo ""
log_info "Application started successfully!"
echo ""
log_info "If this is your first time:"
log_info "  1. The frontend may take 30-60 seconds to compile"
log_info "  2. Enter an amount (e.g., 501) and click Calculate"
log_info "  3. Try the edge case: Change packs to 23,31,53 and calculate 500000"
echo ""
