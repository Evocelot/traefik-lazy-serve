# traefik-lazy-serve
LazyServe is a Traefik middleware plugin that delays HTTP request forwarding when the backend service is unavailable (e.g., scaled to zero). It waits and retries for a configurable period while exposing Prometheus metrics about incoming requests to support auto-scaling triggers.
