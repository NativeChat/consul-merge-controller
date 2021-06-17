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

consul_version="1.9.6"
consul_version_string="Consul v${consul_version}"

consul_k8e_version="0.25.0"
consul_k8e_version_string="consul-k8s v${consul_k8e_version}"

if [[ $(${dest_dir}/consul --version | grep "${consul_version_string}" || echo false) != "${consul_version_string}" ]]; then
    curl -sL "https://releases.hashicorp.com/consul/${consul_version}/consul_${consul_version}_${os_name}_amd64.zip" -o "${consul_zip_path}"
    unzip -o "${consul_zip_path}" -d "${dest_dir}"
else
    echo "Using the existing consul binary from ${dest_dir}"
fi

if [[ "$(${dest_dir}/consul-k8s version || echo false)" != "${consul_k8e_version_string}" ]]; then
    curl -sL "https://releases.hashicorp.com/consul-k8s/${consul_k8e_version}/consul-k8s_${consul_k8e_version}_${os_name}_amd64.zip" -o "${consul_k8s_zip_path}"
    unzip -o "${consul_k8s_zip_path}" -d "${dest_dir}"
else
    echo "Using the existing consul-k8s binary from ${dest_dir}"
fi

