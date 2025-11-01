# Configuration Guide

This document provides detailed information about configuring and using SecuredContainer's image rebuilding functionality.

## Image Rebuilding Configuration

### Registry Authentication

SecuredContainer needs appropriate credentials to push rebuilt images. You can configure this using Kubernetes secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: registry-credentials
  namespace: default
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: <base64-encoded-docker-config>
```

To create this secret from your existing Docker configuration:
```bash
kubectl create secret generic registry-credentials \
  --from-file=.dockerconfigjson=$HOME/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson
```

### Rebuild Strategy Configuration

You can configure how images are rebuilt using annotations on your ContainerSecurity resource:

```yaml
apiVersion: security.securedcontainer.io/v1alpha1
kind: ContainerSecurity
metadata:
  name: example-security
  annotations:
    securedcontainer.io/rebuild-strategy: "aggressive" # or "conservative"
    securedcontainer.io/preserve-labels: "true"
    securedcontainer.io/additional-packages: "ca-certificates,curl"
spec:
  # ... rest of the configuration
```

Available annotations:
- `securedcontainer.io/rebuild-strategy`: 
  - `aggressive`: Updates all packages (default)
  - `conservative`: Updates only packages with known vulnerabilities
- `securedcontainer.io/preserve-labels`: Maintains original image labels
- `securedcontainer.io/additional-packages`: Comma-separated list of additional packages to install

## Local Development and Testing

To test image rebuilding locally:

1. Install required tools:
   ```bash
   # Install Trivy
   curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.44.0

   # Verify Docker installation
   docker --version
   ```

2. Test image scanning:
   ```bash
   # Scan an image
   trivy image nginx:latest

   # Get JSON output for processing
   trivy image -f json nginx:latest > scan_results.json
   ```

3. Test image rebuilding:
   ```bash
   # Create a test Dockerfile
   cat << EOF > Dockerfile.secure
   FROM nginx:latest
   RUN apt-get update && \
       DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
   EOF

   # Build the secure image
   docker build -t nginx:secure -f Dockerfile.secure .

   # Verify the new image
   trivy image nginx:secure
   ```

4. Compare results:
   ```bash
   # Compare vulnerability counts
   trivy image --format json nginx:latest | jq '.Vulnerabilities | length'
   trivy image --format json nginx:secure | jq '.Vulnerabilities | length'
   ```

## Troubleshooting

Common issues and solutions:

1. **Registry Push Failures**
   - Verify registry credentials
   - Check network connectivity
   - Ensure proper image naming convention

2. **Build Failures**
   - Check available disk space
   - Verify Docker daemon status
   - Review build logs for package conflicts

3. **Scanning Issues**
   - Update Trivy vulnerability database
   - Check network access to vulnerability feeds
   - Verify Trivy cache directory permissions

4. **Performance Optimization**
   - Use BuildKit for faster builds
   - Enable Trivy cache
   - Configure resource limits appropriately