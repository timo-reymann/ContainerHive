FROM busybox AS bin
COPY ./dist /binaries
RUN if [[ "$(arch)" == "x86_64" ]]; then \
        architecture="amd64"; \
    else \
        architecture="arm64"; \
    fi; \
    cp /binaries/linux-${architecture}/ch /bin/ch && \
    chmod +x /bin/ch && \
    chown 1000:1000 /bin/ch

FROM scratch AS meta
COPY LICENSE /
COPY THIRD_PARTY_NOTICES /

FROM chainguard/wolfi-base

ARG BUILD_TIME
ARG BUILD_VERSION
ARG BUILD_COMMIT_REF
LABEL 
LABEL org.opencontainers.image.licenses="gpl-3.0"org.opencontainers.image.title="ContainerHive" \
      org.opencontainers.image.description="Swarm it. Build it. Run it. — Managing container base and library images has never been easier." \
      org.opencontainers.image.ref.name="main" \
      org.opencontainers.image.licenses='GPLv3' \
      org.opencontainers.image.vendor="Timo Reymann <mail@timo-reymann.de>" \
      org.opencontainers.image.authors="Timo Reymann <mail@timo-reymann.de>" \
      org.opencontainers.image.url="https://github.com/timo-reymann/ContainerHive" \
      org.opencontainers.image.documentation="https://github.com/timo-reymann/ContainerHive" \
      org.opencontainers.image.source="https://github.com/timo-reymann/ContainerHive.git" \
      org.opencontainers.image.created=$BUILD_TIME \
      org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.revision=$BUILD_COMMIT_REF

RUN apk add --no-cache bash \
    && adduser -D -u 1000 container-hive

COPY --from=bin /bin/ch /bin/ch
WORKDIR /usr/share/doc/containerhive
COPY --from=meta / ./

RUN chmod +x /bin/ch && \
    chown 1000:1000 /bin/ch

WORKDIR /workspace
USER 1000:1000
ENTRYPOINT ["/bin/ch"]
