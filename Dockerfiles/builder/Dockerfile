FROM multiarch/crossbuild

ENV PATH $PATH:/usr/local/go/bin

# Install tooling
ENV GORELEASER_VER 0.113.0
ENV GORELEASER_SHA 2379beebb6369b75ccead7f7a43269de700b51821feae3857701d106ed72bd63
ENV GOVER 1.13
RUN set -ex \
    && curl -o goreleaser.tar.gz -LO https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VER}/goreleaser_Linux_x86_64.tar.gz \
    && echo "$GORELEASER_SHA goreleaser.tar.gz" | sha256sum -c - || exit 1 \
    && tar -xzf goreleaser.tar.gz -C /usr/bin/ goreleaser \
    && rm goreleaser.tar.gz \
    && curl -LO https://dl.google.com/go/go${GOVER}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GOVER}.linux-amd64.tar.gz \
    && rm go${GOVER}.linux-amd64.tar.gz
