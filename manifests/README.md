# Pilotwave Raw Kubernetes YAML

This directory contains the raw YAML deployment path:

- `rbac.yaml` is the current RBAC manifest for the Pilotwave service account and permissions.
- `portal-deploy.yaml` is the current raw deployment/service manifest.

Prefer the Helm chart in `build/helm/pilotwave/` for packaged installs and repeatable environment overrides. Use [`../docs/HELM.md`](../docs/HELM.md) for the maintained Helm deployment guide. Use these raw manifests only for direct smoke tests, debugging, or comparing generated chart output.

## Apply

```sh
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/portal-deploy.yaml
```

Before applying in a real environment, replace the placeholder
`pilotwave-auth` Secret value in `portal-deploy.yaml` or create the Secret from
your deployment automation.

## Validation

The default local validation context remains `colima-legacy-1-18`:

```sh
ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio
ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio
```

## Secrets

Do not commit live secrets, tokens, kubeconfigs, or cluster-specific values into this directory. Supply secrets through Kubernetes Secrets, Helm values, or environment-specific automation.
