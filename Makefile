# zentao-mini Makefile (纯网页版)

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
	@echo "Running backend server..."
	@cd backend && go run cmd/server/main.go

backend-build:
	@echo "Building backend server..."
	@cd backend && go build -o server cmd/server/main.go

backend-app-run:
	@echo "Running embedded app (server + frontend static files)..."
	@cd backend && go run cmd/app/main.go

backend-app-build:
	@echo "Building embedded app (server + frontend static files)..."
	@cd backend && go build -o app cmd/app/main.go

# 组合命令
install:
	@echo "Installing all dependencies..."
	@make frontend-install
	@go mod tidy

build:
	@echo "Building web app (frontend + embedded binary)..."
	@make frontend-build
	@./scripts/copy-static.sh
	@make backend-app-build
	@echo "Build completed: backend/app"

release: build
	@echo "Release build completed"

# 清理命令
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf frontend/dist
	@rm -f backend/server backend/app

# 帮助命令
help:
	@echo "zentao-mini Makefile (纯网页版)"
	@echo ""
	@echo "Frontend commands:"
	@echo "  make frontend-install   - Install frontend dependencies"
	@echo "  make frontend-dev       - Run frontend dev server (6100)"
	@echo "  make frontend-build     - Build frontend"
	@echo ""
	@echo "Backend commands:"
	@echo "  make backend-install    - Tidy backend dependencies"
	@echo "  make backend-run        - Run HTTP server (12345)"
	@echo "  make backend-build      - Build HTTP server binary"
	@echo "  make backend-app-run    - Run embedded app (server + frontend)"
	@echo "  make backend-app-build  - Build embedded app binary"
	@echo ""
	@echo "Combined commands:"
	@echo "  make install            - Install all dependencies"
	@echo "  make build              - Frontend + embedded binary"
	@echo "  make release            - Same as build"
	@echo ""
	@echo "Other commands:"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make help               - Show this help"
