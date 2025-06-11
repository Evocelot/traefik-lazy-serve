# traefik-lazy-serve

LazyServe is a Traefik middleware plugin that delays HTTP request forwarding when the backend service is unavailable (e.g., scaled to zero). It waits and retries for a configurable period before forwarding the response.

This plugin is especially useful in environments where backend services may be scaled to zero and automatically started on-demand (e.g., via KEDA, Knative, or custom auto-scaling logic). Instead of failing immediately, LazyServe gives the service time to start and respond.

## Configuration

```yaml
http:
  middlewares:
    retry-lazyserve:
      plugin:
        lazyserve:
          maxRetries: 5
          retryDelay: 3s
```

- maxRetries: Number of times to retry
- retryDelay: Time to wait between retries (e.g., "2s", "500ms")
