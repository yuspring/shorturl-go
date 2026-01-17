podman build --layers -t web-server:latest .

podman kube play kube.yaml secret.yaml --network pasta

cat secret.yaml kube.yaml | podman kube play - --network pasta

podman kube down kube.yaml secret.yaml

cat secret.yaml kube.yaml | podman kube down -