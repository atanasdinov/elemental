#!/bin/bash

set -uo pipefail

declare -A hosts

{{- range .Nodes }}
hosts[{{ .Hostname }}]={{ .Type }}
{{- end }}

# This is to support both static and DHCP configurations
HOSTNAME=$(</etc/hostname)
[[ -z "${HOSTNAME}" ]] \
  && HOSTNAME=$(</proc/sys/kernel/hostname)

NODETYPE="${hosts[${HOSTNAME}]:-server}"
CONFIGFILE="{{ .KubernetesDir }}/${NODETYPE}.yaml"

if [[ "${HOSTNAME}" = "{{ .InitNode.Hostname }}" ]]; then
  echo "Setting up init node"
  CONFIGFILE={{ .KubernetesDir }}/init.yaml
fi

mkdir -p /etc/rancher/rke2
echo "Copying RKE2 config file ${CONFIGFILE}"
cp ${CONFIGFILE} /etc/rancher/rke2/config.yaml

{{- if and .APIVIP4 .APIHost }}
grep -q "{{ .APIVIP4 }} {{ .APIHost }}" /etc/hosts \
  || echo "{{ .APIVIP4 }} {{ .APIHost }}" >> /etc/hosts
{{- end }}

{{- if and .APIVIP6 .APIHost }}
grep -q "{{ .APIVIP6 }} {{ .APIHost }}" /etc/hosts \
  || echo "{{ .APIVIP6 }} {{ .APIHost }}" >> /etc/hosts
{{- end }}

echo "Installing RKE2 from embedded artifacts..."

export INSTALL_RKE2_ARTIFACT_PATH={{ .InstallPath }}
export INSTALL_RKE2_TAR_PREFIX=/opt/rke2

sh {{ .InstallScript }}

systemctl enable --now rke2-${NODETYPE}.service
