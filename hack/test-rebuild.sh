#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to print status messages
print_status() {
    echo -e "${YELLOW}[*]${NC} $1"
}

# Function to print success messages
print_success() {
    echo -e "${GREEN}[+]${NC} $1"
}

# Function to print error messages
print_error() {
    echo -e "${RED}[-]${NC} $1"
}

# Check required tools
check_requirements() {
    print_status "Checking required tools..."
    
    if ! command -v buildah &> /dev/null; then
        print_error "Buildah is not installed. Please install Buildah first."
        exit 1
    fi
    
    if ! command -v trivy &> /dev/null; then
        print_error "Trivy is not installed. Please install Trivy first."
        exit 1
    }
    
    print_success "All required tools are available"
}

# Scan an image with Trivy
scan_image() {
    local image=$1
    local output_file=$2
    
    print_status "Scanning image: $image"
    trivy image -f json "$image" > "$output_file"
    
    # Get vulnerability count
    local vuln_count=$(jq '.Vulnerabilities | length' "$output_file")
    print_success "Found $vuln_count vulnerabilities"
    return "$vuln_count"
}

# Rebuild an image with security patches
rebuild_image() {
    local source_image=$1
    local target_image=$2
    
    print_status "Creating secure Dockerfile for $source_image"
    
    # Create temporary Dockerfile
    cat << EOF > Dockerfile.secure
FROM $source_image

# Try different package managers (only the appropriate one will succeed)
RUN set -ex && \
    # Alpine
    (apk update && apk upgrade --no-cache) || \
    # Debian/Ubuntu
    (apt-get update && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y) || \
    # RHEL/CentOS
    (yum update -y && yum upgrade -y) || \
    true
EOF
    
    print_status "Building secure image: $target_image"
    buildah bud --format docker -t "$target_image" -f Dockerfile.secure .
    
    # Cleanup
    rm Dockerfile.secure
    
    print_success "Successfully built secure image"
}

# Compare vulnerability counts
compare_results() {
    local original_count=$1
    local new_count=$2
    
    local difference=$((original_count - new_count))
    
    if [ "$difference" -gt 0 ]; then
        print_success "Reduced vulnerabilities by $difference"
    elif [ "$difference" -eq 0 ]; then
        print_status "No vulnerability reduction achieved"
    else
        print_error "Vulnerability count increased by $((-difference))"
    fi
}

# Main function
main() {
    if [ -z "$1" ]; then
        print_error "Please provide an image to secure"
        echo "Usage: $0 <image-name>"
        exit 1
    fi
    
    local source_image=$1
    local target_image="${source_image%:*}:secure"
    
    check_requirements
    
    # Create temporary directory for scan results
    local tmp_dir=$(mktemp -d)
    local original_scan="$tmp_dir/original_scan.json"
    local secure_scan="$tmp_dir/secure_scan.json"
    
    # Scan original image
    scan_image "$source_image" "$original_scan"
    local original_count=$?
    
    # Rebuild image
    rebuild_image "$source_image" "$target_image"
    
    # Scan secure image
    scan_image "$target_image" "$secure_scan"
    local new_count=$?
    
    # Compare results
    compare_results "$original_count" "$new_count"
    
    # Cleanup
    rm -rf "$tmp_dir"
}

# Run main function with provided arguments
main "$@"