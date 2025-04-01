# Build static golang binary
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /ioping_prober


# We need ioping>=1.2, from at least ubuntu:24.04
# As we use `-warmup 0` arg, introduced in 1.2.
FROM ubuntu:24.04
LABEL maintainer="Dale Smith <dalees@gmail.com>"

RUN apt-get update && \
    apt-get install -y ioping && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /ioping_prober /bin/ioping_prober

EXPOSE 9374
ENTRYPOINT  [ "/bin/ioping_prober" ]
