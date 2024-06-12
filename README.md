# Video Microservice GoLang

## Usage

```sh
curl http://localhost:8080/video -F "file=@./data/example.mp4"
```

```sh
curl http://localhost:8080/video/:id
```

```sh
curl http://localhost:8080/video/:id/frame/:index/content -o ./data/frame.jpg
```
