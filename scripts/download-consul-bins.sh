#!/bin/bash

set -euo pipefail

dest_dir="${1}/bin"

os_name="linux"
if [[ "${OSTYPE}" == "darwin"* ]]; then
    os_name="darwin"
fi

tmp_root="/tmp"

consul_zip_path="${tmp_root}/consul.zip"
consul_k8s_zip_path="${tmp_root}/consul-k8s.zip"

consul_version_string="Consul v1.9.2"
consul_k8e_version_string="consul-k8s v0.23.0"

if [[ $(${dest_dir}/consul --version | grep "${consul_version_string}" || echo false) != "${consul_version_string}" ]]; then
    curl -sL "https://releases.hashicorp.com/consul/1.9.2/consul_1.9.2_${os_name}_amd64.zip" -o "${consul_zip_path}"
    unzip -o "${consul_zip_path}" -d "${dest_dir}"
else
    echo "Using the existing consul binary from ${dest_dir}"
fi

if [[ "$(${dest_dir}/consul-k8s version || echo false)" != "${consul_k8e_version_string}" ]]; then
    curl -sL "https://releases.hashicorp.com/consul-k8s/0.23.0/consul-k8s_0.23.0_${os_name}_amd64.zip" -o "${consul_k8s_zip_path}"
    unzip -o "${consul_k8s_zip_path}" -d "${dest_dir}"
else
    echo "Using the existing consul-k8s binary from ${dest_dir}"
fi

