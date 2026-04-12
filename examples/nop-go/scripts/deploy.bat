@echo off
setlocal enabledelayedexpansion

:: nop-go Docker Deploy Script for Windows
:: Usage: deploy.bat [command]

set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
cd /d "%PROJECT_DIR%"

echo.
echo ====================================================
echo   nop-go Microservices Docker Deploy Script
echo ====================================================
echo.

if "%1"=="" goto :help
if "%1"=="help" goto :help
if "%1"=="--help" goto :help
if "%1"=="-h" goto :help

:: Check Docker
:check_docker
echo Checking dependencies...
where docker >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker not installed
    exit /b 1
)
echo [OK] Docker installed

:: Execute command
if "%1"=="build" goto :build
if "%1"=="up" goto :up
if "%1"=="down" goto :down
if "%1"=="restart" goto :restart
if "%1"=="logs" goto :logs
if "%1"=="status" goto :status
if "%1"=="clean" goto :clean
if "%1"=="all" goto :all
if "%1"=="infra" goto :infra
if "%1"=="swagger" goto :swagger

echo ERROR: Unknown command: %1
goto :help

:help
echo Usage: %~nx0 [command]
echo.
echo Commands:
echo   build       Build all service images
echo   up          Start all services
echo   down        Stop all services
echo   restart     Restart all services
echo   logs        View logs
echo   status      View service status
echo   clean       Clean all containers and images
echo   all         Build + Start (full deploy)
echo   infra       Start only infrastructure (MySQL, Redis)
echo   swagger     Start only Swagger UI
echo.
echo Examples:
echo   %~nx0 all      # Full deploy
echo   %~nx0 status   # View status
exit /b 0

:build
echo Building service images...
docker-compose -f docker-compose.yml build --parallel
echo [OK] Build complete
goto :end

:up
echo Starting all services...
docker-compose -f docker-compose.yml up -d
echo [OK] Services started
goto :status

:down
echo Stopping all services...
docker-compose -f docker-compose.yml down
echo [OK] Services stopped
goto :end

:restart
call :down
call :up
goto :end

:logs
if "%2"=="" (
    docker-compose -f docker-compose.yml logs -f
) else (
    docker-compose -f docker-compose.yml logs -f %2
)
goto :end

:status
echo.
echo Service Status:
echo.
docker-compose -f docker-compose.yml ps
echo.
echo Access URLs:
echo   API Gateway:   http://localhost:8000
echo   Swagger UI:    http://localhost:8080
echo   Admin Service: http://localhost:8001/swagger/index.html
echo   Customer:      http://localhost:8002/swagger/index.html
echo   Catalog:       http://localhost:8003/swagger/index.html
echo.
echo Database:
echo   MySQL: localhost:13306 (root / nop123456)
echo   Redis: localhost:16379
goto :end

:clean
echo Cleaning all containers and images...
docker-compose -f docker-compose.yml down -v --rmi local
echo [OK] Clean complete
goto :end

:all
call :check_docker
echo Generating Swagger docs...
call make swagger-gen >nul 2>&1
echo [OK] Swagger docs generated
echo Building service images...
docker-compose -f docker-compose.yml build --parallel
echo [OK] Images built
echo Starting infrastructure...
docker-compose -f docker-compose.yml up -d mysql redis
timeout /t 15 /nobreak >nul
echo Starting all services...
docker-compose -f docker-compose.yml up -d
echo.
echo ====================================================
echo   Deploy Complete!
echo ====================================================
echo.
echo Swagger UI: http://localhost:8080
call :status
goto :end

:infra
call :check_docker
echo Starting infrastructure (MySQL, Redis)...
docker-compose -f docker-compose.yml up -d mysql redis
echo [OK] Infrastructure started
echo.
echo Database:
echo   MySQL: localhost:13306 (root / nop123456)
echo   Redis: localhost:16379
goto :end

:swagger
echo Starting Swagger UI...
docker-compose -f docker-compose.yml up -d swagger-ui
echo [OK] Swagger UI started
echo.
echo Access URL: http://localhost:8080
goto :end

:end
exit /b 0