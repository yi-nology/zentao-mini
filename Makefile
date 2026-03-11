# Chandao Mini Makefile

# 前端相关命令
frontend-install:
	@echo "Installing frontend dependencies..."
	@cd frontend && npm install

frontend-dev:
	@echo "Running frontend in development mode..."
	@cd frontend && npm run dev

frontend-build:
	@echo "Building frontend..."
	@cd frontend && npm run build

# 后端相关命令
backend-install:
	@echo "Installing backend dependencies..."
	@cd backend && go mod tidy

backend-run:
	@echo "Running backend..."
	@cd backend && go run main.go

backend-build:
	@echo "Building backend..."
	@cd backend && go build -o server main.go

# Wails 相关命令
wails-build:
	@echo "Building Wails application..."
	@wails build

wails-run:
	@echo "Running Wails application..."
	@wails dev

# 组合命令
install:
	@echo "Installing all dependencies..."
	@make frontend-install
	@go mod tidy

build:
	@echo "Building all components..."
	@make frontend-build
	@make wails-build

# 清理命令
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf frontend/dist
	@rm -rf build

# 状态命令
status:
	@echo "Project status:"
	@echo "- Frontend: $(shell ls -la frontend/ | grep package.json | wc -l) package.json found"
	@echo "- Backend: $(shell ls -la | grep main.go | wc -l) main.go found"
	@echo "- Wails: $(shell ls -la | grep wails.json | wc -l) wails.json found"

# 帮助命令
help:
	@echo "Chandao Mini Makefile"
	@echo ""
	@echo "Frontend commands:"
	@echo "  make frontend-install   - Install frontend dependencies"
	@echo "  make frontend-dev       - Run frontend in development mode"
	@echo "  make frontend-build     - Build frontend"
	@echo ""
	@echo "Backend commands:"
	@echo "  make backend-install    - Install backend dependencies"
	@echo "  make backend-run        - Run backend"
	@echo "  make backend-build      - Build backend"
	@echo ""
	@echo "Wails commands:"
	@echo "  make wails-build        - Build Wails application"
	@echo "  make wails-run          - Run Wails application"
	@echo ""
	@echo "Combined commands:"
	@echo "  make install            - Install all dependencies"
	@echo "  make build              - Build all components"
	@echo ""
	@echo "Other commands:"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make status             - Check project status"
	@echo "  make help               - Show this help"
