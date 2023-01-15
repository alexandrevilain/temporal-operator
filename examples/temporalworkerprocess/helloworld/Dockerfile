FROM docker.io/library/golang:1.18-alpine
RUN mkdir -p /app/bin
WORKDIR /app
ENV GOBIN=/app/bin
COPY . .

RUN go install . ./worker

FROM docker.io/alpine:latest
LABEL ios.k8s.display-name="helloworld-worker" \
    maintainer="Keith Tenzer <ktenzer@redhat.com>"

RUN mkdir -p /app/bin
WORKDIR /app/bin
COPY --from=0 /app/bin/worker /app/bin
CMD ["/app/bin/worker" ]
