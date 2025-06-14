# Create a Podman network named 'traefik-network'
create-network:
	podman network create traefik-network

# Build a local Traefik image with the lazyserve plugin included
build-docker-image:
	podman build -t local-traefik-with-lazyserve-plugin:latest .

# Start a Traefik container using the built image, attach it to the custom network,
# map the dashboard port, and mount static and dynamic configuration files.
start-traefik:
	podman run --rm --name custom-traefik \
		--network traefik-network \
		-p 8000:8000 \
		-v ./local-traefik-config/traefik.yml:/etc/traefik/traefik.yml \
		-v ./local-traefik-config/dynamic.yml:/etc/traefik/dynamic.yml \
		local-traefik-with-lazyserve-plugin:latest

# Start a simple HTTP echo server using the 'http-echo' image,
# attach it to the custom network and expose it on port 5678.
start-hello-app:
	podman run --rm \
		--name hello-app \
		--network traefik-network \
		-p 5678:5678 \
		hashicorp/http-echo:1.0 \
		-text="Hello world!"

# Cleans up the local development environment.
clean:
	podman rm -f hello-app 2>/dev/null
	podman rm -f custom-traefik 2>/dev/null
	@podman network exists traefik-network && podman network rm traefik-network || echo "Network already deleted"
	@podman image exists local-traefik-with-lazyserve-plugin:latest && \
	  podman image rm local-traefik-with-lazyserve-plugin:latest || \
	  echo "Docker image already deleted"
