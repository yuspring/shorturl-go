# Shorturl Go

這個專案是用Go寫的一個簡易的短網址服務  
絕大部分由LLM生成，小部分有Code Review  
是我用來當作學習Go語言的一個Project  

這個專案主要目的是為了嘗試podman 容器化的進階管理方式  
全部功能都採用rootless環境，無須使用daemon及root權限管理   
其中包含了使用 podman kube play 來建立一個pod  
並且也採用了quadlet讓systemd控制這個service

## Run Server (dev environment)
直接跑server  
```
go mod tidy
go run .
```

## Build image
建立web-server的image  
```
podman build --layers -t web-server:latest .
```
雖然官方文檔說支援 --build 參數，但是因為相容性，所以沒有加上去
## Run podman kube (podman 5.7 up)

啟動和關閉pod 
```
podman kube play kube.yaml secret.yaml --network pasta
podman kube down kube.yaml secret.yaml
```

--force 參數可以移除volumes

## Run podman kube (podman 5.0 up)
啟動和關閉pod(podman 5.0以上)
```
cat secret.yaml kube.yaml | podman kube play - --network pasta
cat secret.yaml kube.yaml | podman kube down -
```

--force 參數可以移除volumes

## Test quadlet
把web-server.kube放在以下的目錄中  
```
~/.config/containers/systemd/
```

```
cp web-server.kube ~/.config/containers/systemd/
```
使用這行可以看web-server.kube有沒有問題
```
/usr/lib/systemd/system-generators/podman-system-generator --user --dryrun
```

## Run systemd
使用systemd啟動服務
```
systemctl --user daemon-reload
systemctl --user enable web-server --now
```


## Network
