# traefik-lazy-serve

LazyServe is a Traefik middleware plugin that delays forwarding HTTP requests while the target service is unavailable (e.g., scaled to zero). Instead of failing fast with a 50x error, LazyServe waits – retrying at configurable intervals – until the service becomes available and can handle the request.

This plugin is especially useful in environments where services may be scaled to zero and automatically started on-demand (e.g., via KEDA, Knative, or custom auto-scaling logic).

## Installing the plugin in Traefik

Enable the plugin in traefik.yml (or traefik.toml) under the experimental.plugins section:

```yaml
entryPoints:
  web:
    address: ":8000"

providers:
  file:
    filename: "/etc/traefik/dynamic.yml"

experimental:
  plugins:
    lazyserve:
      moduleName: github.com/Evocelot/traefik-lazy-serve
      version: v0.1.0
```

> **_NOTE:_**  Traefik downloads and compiles third-party plugins on first start-up.
Make sure outbound traffic to github.com is permitted or pre-build the binary into a custom image (see Local test environment below).

## Configuring the LazyServe middleware

Add the middleware in your dynamic configuration file (e.g. dynamic.yml):

```yaml
http:
  middlewares:
    retry-lazyserve:
      plugin:
        lazyserve:
          maxRetries: 15
          retryDelay: 1000
          retryStatusCodes:
            - 502
            - 503
            - 504
```

Key | Type | Default | Description
--- | --- | --- | --- 
`maxRetries` | int | 5 | Maximum number of retry attempts if the target service is unavailable.
`retryDelay` | int | 1000 | Delay (in milliseconds) between each retry attempt.
`retryStatusCodes` | []int | 502, 503, 504 | 	List of HTTP status codes that should trigger a retry.
---

### Full example

```yaml
http:
  routers:
    hello-app-router:
      rule: PathPrefix(`/`)
      middlewares:
        - lazyserve
      service: hello-app-service

  middlewares:
    lazyserve:
      plugin:
        lazyserve:
          maxRetries: 15      # how many times to poll
          retryDelay: 1000    # delay between polls
          retryStatusCodes:   # override defaults if needed
            - 502
            - 503
            - 504

  services:
    hello-app-service:
      loadBalancer:
        servers:
          - url: http://hello-app:5678
```

Attach the middleware to any router you want (middlewares: - lazyserve).
The example above retries up to 15 s (15 × 1 s) before returning the last backend response.

## Local test environment

The repository ships with a Makefile that spins up everything you need to see LazyServe in action.

Target | What it does
--- | ---
`make create-network` | Creates a Docker network so Traefik and the demo service can talk.
`make build-docker-image` | Builds a custom Traefik image with `LazyServe pre-installed`.
`make start-traefik` | Starts the Traefik container using configs from `local-traefik-config/`.
`make start-hello-app` | Launches a tiny HTTP echo server so you can test request queuing.
`make clean` | Tears down containers, images, and the network.
---

### Quick test

```bash
make create-network
make build-docker-image
make start-traefik

# In another terminal
make start-hello-app

# In another terminal watch Traefik logs or curl the router:
curl -v http://localhost:8000/
```

1. Stop `hello-app` to simulate scale-to-zero (CTRL + C).
2. Send a request again; Traefik will pause and retry until the service is back (make start-hello-app).
3. Observe that no 502/503 reaches the client while the service boots.
