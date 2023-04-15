FROM golang:1.19 as builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY helloworld.go helloworld.go
COPY worker/ worker/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o workerprocess worker/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/workerprocess .
USER 65532:65532

ENTRYPOINT ["/workerprocess"]
