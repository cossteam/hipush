[English](README.md) | 简体中文

# Hipush

Hipush是一个推送服务器，集成多个移动平台的推送通知，它支持HTTP和GRPC接口。

## Configuration

请参阅默认 [YAML config example](example.yaml):

## Deploy

直接运行项目

```bash
go run cmd/main.go -config xxx.yaml
```

使用Docker运行项目

```bash
docker run -d --name hipush \
-v "$(pwd)/config.yaml:/config/config.yaml" \
-p 7070:7070 \
-p 7071:7071 \
hub.hitosea.com/cossim/hipush \
-config /config/config.yaml
```

使用Docker Compose运行项目 [docker-compose.yaml](docker-compose.yaml)

```bash
docker-compose up -d
```

## Usage

更多示例在 [example](example)

### HTTP

Pushing to iOS

```shell
curl --location --request POST 'http://<hipush-server>:7070/api/v1/platforms/ios/apps/<appName>/devices/<deviceToken>/message'
      --header 'Content-Type: application/json'
      --data-raw '{
      "title": "hipush",
      "content": "hello hipush",
      "priority": "low"
      }'
```
