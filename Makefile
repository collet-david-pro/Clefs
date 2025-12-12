# Makefile pour Gestionnaire de Clés

.PHONY: all build run clean install-deps build-all help

# Variables
APP_NAME=clefs
BUILD_DIR=build
CMD_DIR=./cmd

# Couleurs pour les messages
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

all: build

help:
	@echo "$(GREEN)Gestionnaire de Clés - Commandes disponibles:$(NC)"
	@echo "  make run           - Exécuter l'application en mode développement"
	@echo "  make build         - Compiler pour votre système"
	@echo "  make build-all     - Compiler pour tous les systèmes (Windows, Mac, Linux)"
	@echo "  make install-deps  - Installer les dépendances Go"
	@echo "  make clean         - Nettoyer les fichiers de build"
	@echo "  make test          - Lancer les tests (si disponibles)"

install-deps:
	@echo "$(YELLOW)Installation des dépendances...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dépendances installées$(NC)"

run:
	@echo "$(YELLOW)Démarrage de l'application...$(NC)"
	go run $(CMD_DIR)/main.go

build:
	@echo "$(YELLOW)Compilation pour votre système...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Compilation terminée: $(BUILD_DIR)/$(APP_NAME)$(NC)"

build-windows:
	@echo "$(YELLOW)Compilation pour Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows.exe $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Windows: $(BUILD_DIR)/$(APP_NAME)-windows.exe$(NC)"

build-mac:
	@echo "$(YELLOW)Compilation pour macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-mac $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ macOS: $(BUILD_DIR)/$(APP_NAME)-mac$(NC)"

build-mac-arm:
	@echo "$(YELLOW)Compilation pour macOS (Apple Silicon)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-mac-arm64 $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ macOS ARM: $(BUILD_DIR)/$(APP_NAME)-mac-arm64$(NC)"

build-linux:
	@echo "$(YELLOW)Compilation pour Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(CMD_DIR)/main.go
	@echo "$(GREEN)✓ Linux: $(BUILD_DIR)/$(APP_NAME)-linux$(NC)"

build-all: build-windows build-mac build-mac-arm build-linux
	@echo "$(GREEN)✓ Toutes les versions compilées dans $(BUILD_DIR)/$(NC)"
	@ls -lh $(BUILD_DIR)/

clean:
	@echo "$(YELLOW)Nettoyage...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean -cache
	@echo "$(GREEN)✓ Nettoyage terminé$(NC)"

test:
	@echo "$(YELLOW)Lancement des tests...$(NC)"
	go test ./...
	@echo "$(GREEN)✓ Tests terminés$(NC)"

# Commandes de développement
dev: install-deps run

# Vérification du code
lint:
	@echo "$(YELLOW)Vérification du code...$(NC)"
	go fmt ./...
	go vet ./...
	@echo "$(GREEN)✓ Vérification terminée$(NC)"

# Mise à jour des dépendances
update-deps:
	@echo "$(YELLOW)Mise à jour des dépendances...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)✓ Dépendances mises à jour$(NC)"
