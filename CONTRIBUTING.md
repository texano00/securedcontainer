# Contribution Guidelines

## Development Process

We use a trunk-based development workflow with the following branches:

- `main`: Production-ready code
- `develop`: Integration branch for new features
- Feature branches: Short-lived branches for individual features

### Branch Strategy

1. **Main Branch (`main`)**
   - Contains production-ready code
   - Tagged with semantic versions
   - Protected branch - requires PR and reviews
   - Only accepts merges from `develop` or hotfix branches

2. **Development Branch (`develop`)**
   - Integration branch for new features
   - CI runs all tests
   - Merges to `main` trigger release process

3. **Feature Branches**
   - Created from `develop`
   - Format: `feature/brief-description`
   - Merged back to `develop`
   - Deleted after merge

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backward-compatible functionality
- **PATCH** version for backward-compatible bug fixes

Version Format:
- Release versions: `v1.2.3`
- Development versions: `v1.2.3-dev.commit`
- Feature branch versions: `v1.2.3-develop.commit`

### Release Process

1. **Automated Releases**
   - Tags on `main` trigger releases
   - Creates GitHub releases
   - Publishes container images
   - Updates Helm charts

2. **Development Builds**
   - Built from `develop` branch
   - Tagged with commit hash
   - Available for testing

### CI/CD Pipeline

The pipeline includes:
1. Version determination
2. Container image building
3. Helm chart packaging
4. Artifact publishing

### Getting Started

1. Fork the repository
2. Create feature branch from `develop`
3. Make changes and test
4. Submit PR to `develop`

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation
- chore: Maintenance
- refactor: Code restructuring
- test: Adding tests