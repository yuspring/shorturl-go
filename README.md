# Shorturl Go

這個專案是用Go寫的一個簡易的短網址服務
基本上由LLM生成，小部分有Code Review
是我用來當作學習Go的一個Pproject
其中也包含了使用podman來建離一個pods
裡面演示了使用podman kube play 以及使用 quadlet 的方案

## Run Server
```
go mod tidy
go run .
```

## Build image
podman build --layers -t web-server:latest .

## Run podman kube (5.7 up)
```
podman kube play kube.yaml secret.yaml --network pasta
podman kube down kube.yaml secret.yaml
```

## Run podman kube (All version)

```
cat secret.yaml kube.yaml | podman kube play - --network pasta
cat secret.yaml kube.yaml | podman kube down -
```
## Test quadlet

```
~/.config/containers/systemd/
```

```
/usr/lib/systemd/system-generators/podman-system-generator --user --dryrun
```