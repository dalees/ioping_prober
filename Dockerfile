# Build latest ioping.
# Build golang binary
# Create final container image, COPY both.

ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
LABEL maintainer="Dale Smith <dalees@gmail.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/ioping_prober /bin/ioping_prober

EXPOSE 9374
ENTRYPOINT  [ "/bin/ioping_prober" ]
