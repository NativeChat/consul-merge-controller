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
      serviceRouter: service-a
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
      serviceRouter: service-a
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

## Local development
1. Install the Golang dependencies
    ```bash
    go mod vendor
    ```
2. Fix the dependency versioning issue with `consul-k8s`. This step will be removed when `consul-k8s` starts using the new `controller-runtime`.
    ```bash
    make go-mod-vendor-hack
    ```
3. Run the controller
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
