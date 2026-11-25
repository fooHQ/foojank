FROM alpine:3.24@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

RUN apk add --no-cache ca-certificates

RUN addgroup -g 65532 -S foojankd \
    && adduser -u 65532 -S -D -G foojankd foojankd \
    && mkdir -p /home/foojankd/.config/foojank \
    && chown -R 65532:65532 /home/foojankd \
    && chmod 700 /home/foojankd /home/foojankd/.config /home/foojankd/.config/foojank

ENV HOME=/home/foojankd

ARG TARGETARCH
COPY --chmod=555 build/foojankd-linux-${TARGETARCH} /usr/local/bin/foojankd
COPY --chmod=755 install/roles.json /usr/share/foojank/
COPY --chmod=755 docker/foojankd/entrypoint.sh /entrypoint.sh

ARG VERSION=unknown
ARG REVISION=unknown
LABEL org.opencontainers.image.title="foojankd" \
      org.opencontainers.image.source="https://github.com/foohq/foojank" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

USER 65532

ENTRYPOINT ["/entrypoint.sh"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/usr/local/bin/foojankd", "check"]
