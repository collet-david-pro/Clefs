#!/bin/bash
# Compile l'application Clefs en bundle .app pour macOS (Apple Silicon ou Intel
# selon la machine courante). La base de données étant créée à côté de
# l'exécutable, le .app est surtout destiné aux tests locaux ; pour un dossier
# portable (clé USB, partage réseau), préférez le binaire nu :
#   go build -ldflags="-s -w" -o clefs ./cmd/main.go

set -e

# Version lue depuis la constante centralisée (internal/gui/version.go)
VERSION=$(grep -o 'AppVersion = "[^"]*"' internal/gui/version.go | cut -d'"' -f2)
echo "--- Compilation de Clefs ${VERSION} ---"

echo "--- Nettoyage des anciennes versions ---"
rm -rf Clefs.app

echo "--- Création de la structure du bundle .app ---"
mkdir -p "Clefs.app/Contents/MacOS"

echo "--- Création du fichier Info.plist ---"
cat > "Clefs.app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>clefs</string>
  <key>CFBundleIdentifier</key>
  <string>com.github.david35.clefs</string>
  <key>CFBundleName</key>
  <string>Clefs</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>${VERSION}</string>
  <key>CFBundleVersion</key>
  <string>${VERSION}</string>
  <key>LSMinimumSystemVersion</key>
  <string>10.12</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
PLIST

echo "--- Compilation du binaire Go ---"
go build -ldflags="-s -w" -o "Clefs.app/Contents/MacOS/clefs" ./cmd/main.go

echo ""
echo "✅ Compilation terminée : Clefs.app (version ${VERSION})"
