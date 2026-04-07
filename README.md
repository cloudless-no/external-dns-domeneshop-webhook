# ExternalDNS - UNOFFICIAL Domeneshop Webhook

[ExternalDNS](https://github.com/kubernetes-sigs/external-dns) is a Kubernetes
add-on for automatically DNS records for Kubernetes services using different
providers. By default, Kubernetes manages DNS records internally, but
ExternalDNS takes this functionality a step further by delegating the management
of DNS records to an external DNS provider such as this one. This webhook allows
you to manage your Domeneshop domains inside your kubernetes cluster.


## Requirements

This webhook can be used in conjunction with **ExternalDNS v0.19.0 or higher**,
configured for using the webhook interface. Some examples for a working
configuration are shown in the next section.


## Kubernetes Deployment

The Domeneshop webhook is provided as a regular Open Container Initiative (OCI)
image released in the
[GitHub container registry](https://github.com/cloudless-no/external-dns-domeneshop-webhook/pkgs/container/external-dns-domeneshop-webhook).
The deployment can be performed in every way Kubernetes supports.

Here are provided examples using the
[External DNS chart](#using-the-externaldns-chart) and the
[Bitnami chart](#using-the-bitnami-chart).

In either case, a secret that stores the Domeneshop API key is required:

```yaml
kubectl create secret generic domeneshop-credentials --from-literal=APIToken='<EXAMPLE_PLEASE_REPLACE>' --from-literal=APISecret='<EXAMPLE_PLEASE_REPLACE>' -n external-dns
```

### Using the ExternalDNS chart

Skip this step if you already have the ExternalDNS repository added:

```shell
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
```

Update your helm chart repositories:

```shell
helm repo update
```

You can then create the helm values file, for example
`external-dns-domeneshop-values.yaml`:

```yaml
namespace: external-dns
policy: sync
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/cloudless-no/external-dns-domeneshop-webhook
      tag: v0.1.5
    env:
      - name: DOMENESHOP_TOKEN
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APIToken
      - name: DOMENESHOP_SECRET
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APISecret   
    livenessProbe:
      httpGet:
        path: /health
        port: http-webhook
      initialDelaySeconds: 10
      timeoutSeconds: 5
    readinessProbe:
      httpGet:
        path: /ready
        port: http-webhook
      initialDelaySeconds: 10
      timeoutSeconds: 5

extraArgs:
  - "--txt-prefix=reg-%{record_type}-"
```

And then:

```shell
# install external-dns with helm
helm install external-dns-domeneshop external-dns/external-dns -f external-dns-domeneshop-values.yaml -n external-dns
```

### Using the Bitnami chart

> [!NOTE]
> The Bitnami distribution model changed and most features are now paid for.

Skip this step if you already have the Bitnami repository added:

```shell
helm repo add bitnami https://charts.bitnami.com/bitnami
```

Update your helm chart repositories:

```shell
helm repo update
```

You can then create the helm values file, for example
`external-dns-domeneshop-values.yaml`:

```yaml
provider: webhook
policy: sync
extraArgs:
  webhook-provider-url: http://localhost:8888
  txt-prefix: "reg-%{record_type}-"

sidecars:
  - name: domeneshop-webhook
    image: ghcr.io/cloudless-no/external-dns-domeneshop-webhook:v0.12.0
    ports:
      - containerPort: 8888
        name: webhook
      - containerPort: 8080
        name: http-wh-metrics
    livenessProbe:
      httpGet:
        path: /health
        port: http-wh-metrics
      initialDelaySeconds: 10
      timeoutSeconds: 5
    readinessProbe:
      httpGet:
        path: /ready
        port: http-wh-metrics
      initialDelaySeconds: 10
      timeoutSeconds: 5
    env:
      - name: DOMENESHOP_TOKEN
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APIToken
      - name: DOMENESHOP_SECRET
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APISecret
```

And then:

```shell
# install external-dns with helm
helm install external-dns-domeneshop bitnami/external-dns -f external-dns-domeneshop-values.yaml -n external-dns
```

## Domeneshop labels

Domeneshop labels are supported from version **0.8.0** as provider-specific
annotations. This feature has some additional requirements to work properly:

- External DNS in use must be **0.19.0** or higher

The labels are set with an annotation prefixed with:
`external-dns.alpha.kubernetes.io/webhook-domeneshop-label-`.

For example, if we want to set these labels:

| Label      | Value      |
| ---------- | ---------- |
| it.env     | production |
| department | education  |

The annotation syntax will be:

```yaml
  external-dns.alpha.kubernetes.io/webhook-domeneshop-label-it.env: production
  external-dns.alpha.kubernetes.io/webhook-domeneshop-label-department: education
```

This kind of label:

| Label        | Value |
| ------------ | ----- |
| prefix/label | value |

requires an escape sequence for the slash part. By default this will be:
`--slash--`, so the label will be written as:

```yaml
  external-dns.alpha.kubernetes.io/webhook-domeneshop-label-prefix--slash--label: value
```



## Environment variables

The following environment variables can be used for configuring the application.

### Domeneshop DNS API calls configuration

These variables control the behavior of the webhook when interacting with
Domeneshop DNS API.

      - name: DOMENESHOP_TOKEN
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APIToken
      - name: DOMENESHOP_SECRET
        valueFrom:
          secretKeyRef:
            name: domeneshop-credentials
            key: APISecret  

| Variable        | Description                            | Notes                      |
| --------------- | -------------------------------------- | -------------------------- |
| DOMENESHOP_TOKEN | Domeneshop API token                      | Mandatory                  |
| DOMENESHOP_SECRET | Domeneshop API secret                      | Mandatory                  |
| SLASH_ESC_SEQ   | Escape sequence for label annotations  | Default: `--slash--`       |
| MAX_FAIL_COUNT  | Number of failed calls before shutdown | Default: `-1` (disabled)   |
| DOMAIN_CACHE_TTL  | TTL for the domain cache in seconds      | Default: `0` (disabled)    |


### Test and debug

These environment variables are useful for testing and debugging purposes.

| Variable        | Description                      | Notes            |
| --------------- | -------------------------------- | ---------------- |
| DRY_RUN         | If set, changes won't be applied | Default: `false` |
| DOMENESHOP_DEBUG   | Enables debugging messages       | Default: `false` |

### Socket configuration

These variables control the sockets that this application listens to.

| Variable        | Description                      | Notes                |
| --------------- | -------------------------------- | -------------------- |
| WEBHOOK_HOST    | Webhook hostname or IP address   | Default: `localhost` |
| WEBHOOK_PORT    | Webhook port                     | Default: `8888`      |
| METRICS_HOST    | Metrics hostname                 | Default: `0.0.0.0`   |
| METRICS_PORT    | Metrics port                     | Default: `8080`      |
| READ_TIMEOUT    | Sockets' read timeout in ms      | Default: `60000`     |
| WRITE_TIMEOUT   | Sockets' write timeout in ms     | Default: `60000`     |

Please notice that the following variables were **deprecated**:

| Variable    | Description                            |
| ----------- | -------------------------------------- |
| HEALTH_HOST | Metrics hostname (deprecated)          |
| HEALTH_PORT | Metrics port (deprecated)              |
| DEFAULT_TTL | The default TTL is taken from the domain |


### Domain filtering

Additional environment variables for domain filtering. When used, this webhook
will be able to work only on domains (domains) matching the filter.

| Environment variable           | Description                        |
| ------------------------------ | ---------------------------------- |
| DOMAIN_FILTER                  | Filtered domains                   |
| EXCLUDE_DOMAIN_FILTER          | Excluded domains                   |
| REGEXP_DOMAIN_FILTER           | Regex for filtered domains         |
| REGEXP_DOMAIN_FILTER_EXCLUSION | Regex for excluded domains         |

If the `REGEXP_DOMAIN_FILTER` is set, the following variables will be used to
build the filter:

 - REGEXP_DOMAIN_FILTER
 - REGEXP_DOMAIN_FILTER_EXCLUSION

 otherwise, the filter will be built using:

 - DOMAIN_FILTER
 - EXCLUDE_DOMAIN_FILTER

## Endpoints

This process exposes several endpoints, that will be available through these
sockets:

| Socket name | Socket address                |
| ----------- | ----------------------------- |
| Webhook     | `WEBHOOK_HOST`:`WEBHOOK_PORT` |
| Metrics     | `METRICS_HOST`:`METRICS_PORT` |

The environment variables controlling the socket addresses are not meant to be
changed, under normal circumstances, for the reasons explained in
[Tweaking the configuration](tweaking-the-configuration).
The endpoints
[expected by ExternalDNS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md)
are marked with *.

### Webhook socket

All these endpoints are
[required by ExternalDNS](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md).

| Endpoint           | Purpose                                        |
| ------------------ | ---------------------------------------------- |
| `/`                | Initialization and `DomainFilter` negotiations |
| `/record`          | Get and apply records                          |
| `/adjustendpoints` | Adjust endpoints before submission             |

### Metrics socket

ExternalDNS doesn't have functional requirements for this endpoint, but some
of them are
[recommended](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md).
In this table those endpoints are marked with  __*__.

| Endpoint           | * | Purpose                                            |
| ------------------ | - | -------------------------------------------------- |
| `/health`          |   | Implements the liveness probe                      |
| `/ready`           |   | Implements the readiness probe                     |
| `/healthz`         | * | Implements a combined liveness and readiness probe |
| `/metrics`         | * | Exposes the available metrics                      |

Please check the [Exposed metrics](#exposed-metrics) section for more
information.

## Tweaking the configuration

While tweaking the configuration, there are some points to take into
consideration:

- if `WEBHOOK_HOST` and `METRICS_HOST` are set to the same address/hostname or
  one of them is set to `0.0.0.0` remember to use different ports. Please note
  that it **highly recommendend** for `WEBHOOK_HOST` to be `localhost`, as
  any address reachable from outside the pod might be a **security issue**;
  besides this, changing these would likely need more tweaks than just setting
  the environment variables. The default settings are compatible with the
  [ExternalDNS assumptions](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/webhook-provider.md);
- if your records don't get deleted when applications are uninstalled, you
  might want to verify the policy in use for ExternalDNS: if it's `upsert-only`
  no deletion will occur. It must be set to `sync` for deletions to be
  processed. Please check that `external-dns-domeneshop-values.yaml` include:

  ```yaml
  policy: sync
  ```
- the `--txt-prefix` parameter should really include: `%{record_type}`, as any
  other value will cause a weird duplication of database records. Change the
  value provided in the sample configuration only if you really know what are
  you doing.

## Exposed metrics

The following metrics related to the API calls towards Domeneshop are available
for scraping.

| Name                         | Type      | Labels   | Description                                              |
| ---------------------------- | --------- | -------- | -------------------------------------------------------- |
| `successful_api_calls_total` | Counter   | `action` | The number of successful Domeneshop API calls               |
| `failed_api_calls_total`     | Counter   | `action` | The number of Domeneshop API calls that returned an error   |
| `filtered_out_domains`         | Gauge     | _none_   | The number of domains excluded by the domain filter        |
| `skipped_records`            | Gauge     | `domain`   | The number of skipped records per domain                 |
| `api_delay_hist`             | Histogram | `action` | Histogram of the delay (ms) when calling the Domeneshop API |

When using the Cloud API also the rate limit metrics will be available:

| Name                      | Type      | Labels   | Description                                         |
| ------------------------- | --------- | -------- | --------------------------------------------------- |
| `ratelimit_limit`         | Gauge     | _none_   | Total API calls that can be performed in a hour     |
| `ratelimit_remaining`     | Gauge     | _none_   | Remaining API calls until the next rate limit reset |
| `ratelimit_reset_seconds` | Gauge     | _none_   | UNIX timestamp for the next rate limit reset        |

The label `action` can assume one of the following values, depending on the
Domeneshop API endpoint called.

The actions supported by the legacy DNS provider are:

- `get_domains`
- `get_records`
- `create_record`
- `delete_record`
- `update_record`



The label `domain` can assume one of the domain names as its value.

## Development

The basic development tasks are provided by make. Run `make help` to see the
available targets.

## Credits

This Webhook was forked and modified from the [Hetzner Webhook](https://github.com/mconfalonieri/external-dns-hetzner-webhook) which was forked and modified from the [IONOS Webhook](https://github.com/ionos-cloud/external-dns-ionos-webhook)
to work with Domeneshop. It also contains huge parts from [DrBu7cher's hcloud provider](https://github.com/DrBu7cher/external-dns/tree/readding_hcloud_provider).

### Contributors

| Name                                         | Contribution                  |
| -------------------------------------------- | ----------------------------- |
| [mconfalonieri](https://github.com/mconfalonieri) | Most of the base code |
| [DerQue](https://github.com/DerQue)          | local CNAME fix               |
| [sschaeffner](https://github.com/sschaeffner)| build configuration for arm64 |
| [sgaluza](https://github.com/sgaluza)        | support for MX records        |
