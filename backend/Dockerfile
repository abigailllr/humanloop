FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /build/server .
COPY extractor/ /extractor/
RUN pip install --no-cache-dir -r /extractor/requirements.txt
EXPOSE 8080
ENV EXTRACTOR_PATH=/extractor/main.py
CMD ["./server"]
