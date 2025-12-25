#!/bin/bash

# Converso CLI Build Script
# Cross-platform build automation for Converso CLI

set -e

# Configuration
VERSION=${VERSION:-$(git describe --tags --always 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "none")}
DATE=${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}
BUILD_DIR="dist"
PLATFORMS="darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        log_warning "Git is not installed, using default version info"
    fi
    
    log_success "Prerequisites check passed"
}

# Clean build directory
clean_build_dir() {
    log_info "Cleaning build directory..."
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    log_success "Build directory cleaned"
}

# Build for each platform
build_platforms() {
    log_info "Building for platforms: $PLATFORMS"
    
    for platform in $PLATFORMS; do
        GOOS=${platform%/*}
        GOARCH=${platform#*/}
        
        log_info "Building for $GOOS/$GOARCH..."
        
        # Set output filename
        if [ "$GOOS" = "windows" ]; then
            OUTPUT="$BUILD_DIR/converso-$GOOS-$GOARCH.exe"
        else
            OUTPUT="$BUILD_DIR/converso-$GOOS-$GOARCH"
        fi
        
        # Build command with ldflags
        GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "
            -X main.version=$VERSION
            -X main.commit=$COMMIT
            -X main.date=$DATE
            -s -w
        " -o "$OUTPUT" ./cmd/converso/
        
        # Make executable
        if [ "$GOOS" != "windows" ]; then
            chmod +x "$OUTPUT"
        fi
        
        log_success "Built: $OUTPUT"
    done
}

# Create archives
create_archives() {
    log_info "Creating distribution archives..."
    
    cd "$BUILD_DIR"
    
    for file in converso-*; do
        if [ -f "$file" ]; then
            if [[ "$file" == *.exe ]]; then
                # Windows executable
                zip "${file%.*}.zip" "$file"
                log_success "Created: ${file%.*}.zip"
            else
                # Unix executable
                tar -czf "${file}.tar.gz" "$file"
                log_success "Created: ${file}.tar.gz"
            fi
        fi
    done
    
    cd ..
}

# Create checksums
create_checksums() {
    log_info "Creating checksums..."
    
    cd "$BUILD_DIR"
    
    if command -v sha256sum &> /dev/null; then
        sha256sum * > SHA256SUMS
    elif command -v shasum &> /dev/null; then
        shasum -a 256 * > SHA256SUMS
    else
        log_warning "No checksum tool available"
    fi
    
    if [ -f SHA256SUMS ]; then
        log_success "Created: SHA256SUMS"
    fi
    
    cd ..
}

# Create release notes
create_release_notes() {
    log_info "Creating release notes..."
    
    cat > "$BUILD_DIR/RELEASE_NOTES.md" << EOF
# Converso CLI Release $VERSION

## Build Information
- **Version**: $VERSION
- **Commit**: $COMMIT
- **Build Date**: $DATE
- **Go Version**: $(go version)

## Files
EOF

    cd "$BUILD_DIR"
    for file in *; do
        if [ -f "$file" ] && [ "$file" != "RELEASE_NOTES.md" ] && [ "$file" != "SHA256SUMS" ]; then
            size=$(ls -lh "$file" | awk '{print $5}')
            echo "- \`$file\` ($size)" >> RELEASE_NOTES.md
        fi
    done
    
    cat >> "$BUILD_DIR/RELEASE_NOTES.md" << EOF

## Installation

### Linux/macOS
\`\`\`bash
# Download and extract
curl -L -o converso.tar.gz https://example.com/converso-linux-amd64.tar.gz
tar -xzf converso.tar.gz
sudo mv converso /usr/local/bin/
\`\`\`

### Windows
\`\`\`powershell
# Download and extract
Invoke-WebRequest -Uri "https://example.com/converso-windows-amd64.zip" -OutFile "converso.zip"
Expand-Archive -Path "converso.zip" -DestinationPath "$env:ProgramFiles\Converso"
\`\`\`

### Verification
\`\`\`bash
# Verify checksums
sha256sum -c SHA256SUMS
\`\`\`

## Usage
\`\`\`bash
converso --help
converso setup
converso login
converso youtube list-formats <url>
\`\`\`
EOF

    cd ..
    log_success "Created: $BUILD_DIR/RELEASE_NOTES.md"
}

# Main build function
main() {
    echo "🚀 Converso CLI Build Script"
    echo "=========================="
    echo "Version: $VERSION"
    echo "Commit: $COMMIT"
    echo "Date: $DATE"
    echo ""
    
    check_prerequisites
    clean_build_dir
    build_platforms
    create_archives
    create_checksums
    create_release_notes
    
    echo ""
    log_success "Build completed successfully!"
    log_info "Artifacts available in: $BUILD_DIR"
    
    # List created files
    echo ""
    echo "📁 Created files:"
    ls -la "$BUILD_DIR"
}

# Run main function
main "$@"
