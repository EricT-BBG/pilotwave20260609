# Project Docs

This directory contains project Markdown notes that are not required at the repository root.

- [Architecture](ARCHITECTURE.md): backend/frontend structure, request flow, auth model, and high-risk areas.
- [AI Change Playbook](AI_CHANGE_PLAYBOOK.md): scoped workflow guidance for AI-assisted changes.
- [Frontend Notes](FRONTEND.md): current Vue 3 + Vite frontend stack, structure, and verification commands.
- [Makefile Guide](MAKEFILE.md): Makefile targets, variables, and common workflows.
- [Helm Deployment](HELM.md): Helm install flows, image handling, values parameters, and rollback.
- [Local Legacy Kubernetes Environment](LOCAL_LEGACY_K8S_ENVIRONMENT.md): standalone Colima/k3s/Istio 1.7.5 setup for Pilotwave local compatibility validation.
- [Release Checklist](RELEASE_CHECKLIST.md): build, scan, push, Helm deployment, validation, and rollback checklist.
- [TODO](TODO.md): Istio compatibility and safe-update backlog.
- [Dependency Security Notes](DEPENDENCY_SECURITY.md): dependency update summary and security verification commands.

Operational note: Pilotwave exposes Prometheus metrics at `/metrics`; use [Helm Deployment](HELM.md) as the canonical Helm install/deploy guide. Raw YAML under [`../manifests/`](../manifests/) is only for direct smoke tests, troubleshooting, or chart-output comparison.
