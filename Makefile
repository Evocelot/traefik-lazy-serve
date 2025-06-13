build-docker-image:
	podman build -t custom-traefik-image:0.1.0 .

start-traefik:
	podman run --rm --name custom-traefik \
		-p 8080:8003 \
		-v ./traefik-config/traefik.yml:/etc/traefik/traefik.yml \
		-v ./traefik-config/dynamic.yml:/etc/traefik/dynamic.yml \
		custom-traefik-image:0.1.0 
