podman build --layers -t web-server:latest .

podman kube play kube.yaml secret.yaml --network pasta

podman kube down kube.yaml secret.yaml