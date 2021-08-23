# consul-merge-controller
Kubernetes controller which merges Consul CRD resources.

<br />

<p align="left">
  <img src="https://github.com/NativeChat/consul-merge-controller/workflows/Build/badge.svg" alt="Build"/>
  <img src="https://github.com/NativeChat/consul-merge-controller/workflows/Test/badge.svg" alt="Test"/>

  <a href="https://github.com/NativeChat/consul-merge-controller/issues">
      <img src="https://img.shields.io/github/issues-raw/NativeChat/consul-merge-controller?style=flat" alt="github issues"/>
  </a>
  <a href="https://hub.docker.com/r/nchatsystem/consul-merge-controller/">
    <img src="https://img.shields.io/docker/pulls/nchatsystem/consul-merge-controller" alt="docker pulls"/>
  </a>
</p>

<hr />

<br />

## The controller provides the following merge functionality:
1. kind: `ServiceRouter` (apiVersion: `consul.hashicorp.com/v1alpha1`) using the `ConsulServiceRoute` CRD provided by this controller.

    Example input:
    ```YAML
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceRoute
    metadata:
      name: service-a-v1
      labels:
        service.consul.k8s.nativechat.com/service-router: service-a
    spec:
      route:
        match:
          http:
            pathPrefix: /v1
        destination:
          service: service-a-v1

    ---
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceRoute
    metadata:
      name: service-a-pr1
      labels:
        service.consul.k8s.nativechat.com/service-router: service-a
    spec:
      route:
        match:
          http:
            pathPrefix: /pr1
        destination:
          service: service-a-pr1

    ---
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceRoute
    metadata:
      name: service-a-pr2
      labels:
        service.consul.k8s.nativechat.com/service-router: service-a
    spec:
      route:
        match:
          http:
            header:
              - name: x-api-version
                exact: pr2
        destination:
          service: service-a-pr2
    ```
    Example result:
    ```YAML
    apiVersion: consul.hashicorp.com/v1alpha1
    kind: ServiceRouter
    metadata:
      name: service-a
    spec:
      routes:
        - match:
            http:
              pathPrefix: /v1
          destination:
            service: service-a-v1
        - match:
            http:
              pathPrefix: /pr1
          destination:
            service: service-a-pr1
        - match:
            http:
              header:
                - name: x-api-version
                  exact: pr2
          destination:
            service: service-a-pr2
    ```

<br />

2. kind: `ServiceIntentions` (apiVersion: `consul.hashicorp.com/v1alpha1`) using the `ConsulServiceIntentionsSource` CRD provided by this controller.

    Example input:
    ```YAML
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceIntentionsSource
    metadata:
      name: service-a-v1-to-service-b-v1
      labels:
        service.consul.k8s.nativechat.com/service-intentions: service-b-v1
    spec:
      source:
        name: service-a-v1
        action: allow

    ---
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceIntentionsSource
    metadata:
      name: service-a-pr1-to-service-b-v1
      labels:
        service.consul.k8s.nativechat.com/service-intentions: service-b-v1
    spec:
      source:
        name: service-a-pr1
        action: allow

    ---
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceIntentionsSource
    metadata:
      name: service-c-v1-to-service-b-v1
      labels:
        service.consul.k8s.nativechat.com/service-intentions: service-b-v1
    spec:
      source:
        name: service-c-v1
        action: allow
    ```
    Example result:
    ```YAML
    apiVersion: consul.hashicorp.com/v1alpha1
    kind: ServiceIntentions
    metadata:
      name: service-b-v1
    spec:
      destination:
        name: service-b-v1
      sources:
      - action: allow
        name: service-a-v1
      - action: allow
        name: service-a-pr1
      - action: allow
        name: service-c-v1
    ```

## Local development
1. Install the Golang dependencies
    ```bash
    go mod vendor
    ```
2. Run the controller
    ```bash
    make run
    ```

## Release
1. Change the `VERSION` variable in the `Makefile`.
2. Build the docker image
    ```bash
    make docker-build
    ```
3. Push the docker image
    ```bash
    make docker-push
    ```
4. Create release in GitHub
