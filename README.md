# consul-merge-controller
Kubernetes controller which merges Consul CRD resources.

## The controller provides the following merge functionality:
1. kind: `ServiceRouter` (apiVersion: `consul.hashicorp.com/v1alpha1`) using the `ConsulServiceRoute` CRD provided by this controller.

    Example input:
    ```YAML
    apiVersion: service.consul.k8s.nativechat.com/v1alpha1
    kind: ConsulServiceRoute
    metadata:
      name: service-a-v1
    spec:
      serviceRouter: service-a
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
            service: service-a
        - match:
            http:
              pathPrefix: /pr1
          destination:
            service: service-a
        - match:
            http:
              header:
                - name: x-api-version
                  exact: pr2
          destination:
            service: service-a
    ```
