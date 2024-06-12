# Video Microservice Go

A microservice written in Go that accepts video files and extracts frames and metadata

## Usage

Run the following command to start the server:

```sh
go run ./cmd/server
```

## Examples

POST a new video

```sh
curl http://localhost:8080/videos -F "file=@./data/sample.mp4"
```

GET a video

```sh
curl http://localhost:8080/videos/:video_id
```

GET the video content

```sh
curl http://localhost:8080/videos/:video_id/content -o ./data/video.mp4
```

GET a frame of a video

```sh
curl http://localhost:8080/videos/:video_id/frames/:frame_number
```

GET the frame content of a video

```sh
curl http://localhost:8080/videos/:id/frames/:frame_number -o ./data/frame.png
```

POST 100 new videos

```sh
for i in {1..100}; do
    curl http://localhost:8080/videos -F "file=@./data/sample.mp4"
    echo # Print a newline
done
```
