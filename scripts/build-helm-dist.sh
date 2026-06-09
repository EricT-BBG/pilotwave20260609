#!/usr/bin/env bash
set -euo pipefail

chart_dir="${HELM_CHART_DIR:-build/helm/pilotwave}"
dist_dir="${DIST_DIR:-build/dist}"
image_tag="${IMAGE_TAG:-}"
helm_guide="${HELM_GUIDE:-docs/HELM.md}"

chart_yaml="${chart_dir}/Chart.yaml"
if [[ ! -f "${chart_yaml}" ]]; then
  echo "Chart.yaml not found: ${chart_yaml}" >&2
  exit 1
fi

chart_name="$(awk -F': *' '/^name:/ {print $2; exit}' "${chart_yaml}")"
chart_version="$(awk -F': *' '/^version:/ {print $2; exit}' "${chart_yaml}")"
app_version="$(awk -F': *' '/^appVersion:/ {print $2; exit}' "${chart_yaml}" | tr -d '"')"

if [[ -z "${chart_name}" || -z "${chart_version}" ]]; then
  echo "Unable to read chart name/version from ${chart_yaml}" >&2
  exit 1
fi

rm -rf "${dist_dir}"
mkdir -p "${dist_dir}"

helm lint "${chart_dir}"
helm template "${chart_name}" "${chart_dir}" --namespace pilotwave >/dev/null

helm package "${chart_dir}" --destination "${dist_dir}" >/dev/null

package_file="${dist_dir}/${chart_name}-${chart_version}.tgz"
if [[ ! -f "${package_file}" ]]; then
  echo "Expected Helm package not found: ${package_file}" >&2
  exit 1
fi

cp "${chart_dir}/values-production.example.yaml" "${dist_dir}/values-production.example.yaml"
cp "${chart_dir}/README.md" "${dist_dir}/HELM_README.md"
if [[ -f "${helm_guide}" ]]; then
  cp "${helm_guide}" "${dist_dir}/HELM_DEPLOYMENT.md"
fi

(
  cd "${dist_dir}"
  checksum_files=("${chart_name}-${chart_version}.tgz" values-production.example.yaml HELM_README.md)
  if [[ -f HELM_DEPLOYMENT.md ]]; then
    checksum_files+=(HELM_DEPLOYMENT.md)
  fi
  shasum -a 256 "${checksum_files[@]}" > SHA256SUMS
)

cat > "${dist_dir}/INSTALL.md" <<EOF
# Pilotwave Helm Release

Artifacts:

- \`${chart_name}-${chart_version}.tgz\`: Helm chart package.
- \`values-production.example.yaml\`: production-oriented values template.
- \`HELM_DEPLOYMENT.md\`: operator-facing Helm deployment guide.
- \`HELM_README.md\`: full Helm install and values reference.
- \`SHA256SUMS\`: checksums for release artifacts.

Chart version: \`${chart_version}\`
App version: \`${app_version}\`
Suggested image tag from this build: \`${image_tag}\`

## Install

\`\`\`sh
cp values-production.example.yaml values-prod.yaml

helm upgrade --install pilotwave ./${chart_name}-${chart_version}.tgz \\
  --namespace pilotwave --create-namespace \\
  -f values-prod.yaml \\
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \\
  --set image.tag=<image-tag> \\
  --set imagePullSecret.existingSecret=<pull-secret-name>
\`\`\`

Before deploying, edit \`values-prod.yaml\` for the target cluster. Do not store live registry tokens, TLS private keys, kubeconfigs, or other secrets in committed values files.

## Verify Package

\`\`\`sh
shasum -a 256 -c SHA256SUMS
helm lint ./${chart_name}-${chart_version}.tgz
helm template pilotwave ./${chart_name}-${chart_version}.tgz --namespace pilotwave >/dev/null
\`\`\`

See \`HELM_DEPLOYMENT.md\` for the recommended deployment flows and common values. See \`HELM_README.md\` for chart-local examples and template details.
EOF

cat <<EOF
Helm distribution created in ${dist_dir}
Package: ${package_file}
Install notes: ${dist_dir}/INSTALL.md
Checksums: ${dist_dir}/SHA256SUMS
EOF
