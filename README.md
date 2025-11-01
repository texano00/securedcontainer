# SecuredContainer

SecuredContainer is an open-source Kubernetes operator that automatically secures running containers by scanning for vulnerabilities and applying security patches.

## Features

- Automatic container image vulnerability scanning using Trivy
- Automated security patching of container images
- Configurable scanning intervals and patch policies
- Selective workload targeting using label selectors
- Built-in metrics and monitoring through Prometheus/Grafana
- Kubernetes-native CustomResourceDefinitions (CRDs)

## Installation

### Prerequisites

- Kubernetes cluster 1.19+
- Helm 3.x
- Buildah (for image rebuilding)
- Container registry credentials

### Using Helm

```bash
# Add the Helm repository
helm repo add securedcontainer https://texano00.github.io/securedcontainer/charts
helm repo update

# Install the latest stable version
helm install securedcontainer securedcontainer/securedcontainer

# Or install a specific version
helm install securedcontainer securedcontainer/securedcontainer --version 1.0.0

# Or using kubectl
kubectl apply -f https://raw.githubusercontent.com/texano00/securedcontainer/main/config/install.yaml
```

## Technical Details

### Image Rebuilding Process

When SecuredContainer detects a vulnerable container image, it performs the following steps:

1. **Vulnerability Scanning**:
   - Uses Trivy to perform a deep scan of the container image
   - Identifies CVEs, vulnerabilities, and outdated packages
   - Generates a detailed vulnerability report

2. **Local Image Rebuilding**:
   - Creates a temporary Dockerfile based on the original image
   - Automatically detects the base OS (Alpine, Debian/Ubuntu, or RHEL/CentOS)
   - Applies appropriate update commands:
     ```dockerfile
     # For Alpine Linux
     RUN apk update && apk upgrade --no-cache

     # For Debian/Ubuntu
     RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y

     # For RHEL/CentOS
     RUN yum update -y && yum upgrade -y
     ```
   - Uses Docker's BuildKit for efficient, layer-optimized builds
   - Maintains original image metadata and labels

3. **Image Verification**:
   - Runs a second Trivy scan on the rebuilt image
   - Compares vulnerability counts before/after
   - Generates a patch report

4. **Registry Integration**:
   - Uses the configured `imagePushSecret` for authentication
   - Pushes the patched image with the configured suffix
   - Maintains a history of patched images for rollback

### Required Tools

The operator requires the following tools to be available in its environment:

1. **Trivy**: For vulnerability scanning
   - Version: Latest stable
   - Used for both initial and verification scans
   - Supports multiple vulnerability databases

2. **Docker**: For image rebuilding
   - Version: 20.10+
   - BuildKit enabled for efficient builds
   - Required permissions:
     - Access to Docker daemon
     - Read/Write to local image store
     - Network access for pulling/pushing images

3. **Container Registry Access**:
   - Push access to target registries
   - Authentication via:
     - Kubernetes secrets
     - Docker config.json
     - Registry-specific credentials

### Example Configuration

```yaml
# Example authentication secret for registry access
apiVersion: v1
kind: Secret
metadata:
  name: registry-credentials
  namespace: default
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: <base64-encoded-docker-config>
---
# SecuredContainer configuration
```

## Usage

1. Create a SecuredContainer resource:

```yaml
apiVersion: security.securedcontainer.io/v1alpha1
kind: ContainerSecurity
metadata:
  name: example-security
  namespace: default
spec:
  selector:
    matchLabels:
      secure: "true"
  scanInterval: 24
  autoPatch: true
  tagSuffix: "-sc"
  imagePushSecret: "registry-credentials"
```

2. Label your workloads:

```bash
kubectl label deployment/my-app secure=true
```

The operator will:
1. Detect workloads matching the selector
2. Scan container images for vulnerabilities using Trivy
3. Create patched versions of vulnerable images
4. Update workloads to use the secured images
5. Monitor for new vulnerabilities continuously

## Architecture

SecuredContainer operates as a Kubernetes operator that:
- Watches for ContainerSecurity resources
- Monitors labeled Deployments and StatefulSets
- Scans images using Trivy
- Creates and pushes patched images
- Updates workload specifications
- Exports metrics for monitoring

## Versioning

SecuredContainer follows [Semantic Versioning](https://semver.org/). Version numbers are in the format `MAJOR.MINOR.PATCH`:

- **MAJOR**: Incompatible API changes
- **MINOR**: New features (backward-compatible)
- **PATCH**: Bug fixes (backward-compatible)

### Version Tags

- Release versions: `v1.2.3`
- Development builds: `v1.2.3-dev.commit`
- Feature builds: `v1.2.3-develop.commit`

### Artifacts

All artifacts are published to GitHub Container Registry (ghcr.io):

- **Container Images**: `ghcr.io/texano00/securedcontainer:$VERSION`
- **Helm Charts**: `oci://ghcr.io/texano00/charts/securedcontainer:$VERSION`

## Development

Please refer to our [Contributing Guide](CONTRIBUTING.md) for:
- Development workflow
- Branch strategy
- Release process
- Coding standards
- Testing requirements

### Quick Start for Developers

```bash
# Clone the repository
git clone https://github.com/texano00/securedcontainer
cd securedcontainer

# Create a feature branch
git checkout develop
git checkout -b feature/my-feature

# Make your changes and test
make test

# Submit a PR to the develop branch
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## Security

For security concerns, please email security@securedcontainer.io or use GitHub Security Advisories.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

# 🛠️ Summary of Flow
1. **SecuredContainer** runs as an operator in your Kubernetes cluster.  
2. It reads a **configuration** defining what workloads to watch and how often.  
3. For each matched deployment/statefulset:
   - Retrieves the container image.  
   - Scans it using **Trivy**.  
   - Rebuilds the image with patched OS packages.  
   - Scans again to show improvements.  
   - Pushes the secured image (`-sc{datetime}` tag).  
   - Updates the deployment to use the new image.  
4. Sends **telemetry** to a local database for **Grafana** visualization.
