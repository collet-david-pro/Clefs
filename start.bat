@echo off
REM Script de démarrage pour Windows

echo ================================
echo Gestionnaire de Cles - Demarrage
echo ================================
echo.

REM Vérifier si Go est installé
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERREUR] Go n'est pas installe sur ce systeme
    echo.
    echo Pour installer Go:
    echo   Telechargez depuis https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo [OK] Go est installe
go version
echo.

REM Vérifier si les dépendances sont installées
if not exist "go.sum" (
    echo Installation des dependances...
    go mod download
    go mod tidy
    echo [OK] Dependances installees
    echo.
)

REM Lancer l'application
echo Lancement de l'application...
echo.
go run ./cmd/main.go

pause
