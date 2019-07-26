FROM dockercore/golang-cross

# Couldn't fine a more elegant way to fix this
RUN ln -s /usr/include/asm-generic /usr/include/asm

RUN set -ex \
    && apt-get update -qq \
    && apt-get install -y -q --no-install-recommends \
    g++-multilib

# Run in a different step to ease dependencies resolution
RUN set -ex \    
    && apt-get install -y -q --no-install-recommends \
    gcc-aarch64-linux-gnu \
    gcc-arm-linux-gnueabihf

# Install goreleaser
ENV GORELEASER_VER 0.113.0
ENV GORELEASER_SHA 2379beebb6369b75ccead7f7a43269de700b51821feae3857701d106ed72bd63
RUN set -ex \
    && curl -o goreleaser.tar.gz -LO https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VER}/goreleaser_Linux_x86_64.tar.gz \
    && echo "$GORELEASER_SHA goreleaser.tar.gz" | sha256sum -c - || exit 1 \
    && tar -xzf goreleaser.tar.gz -C /usr/bin/ goreleaser \
    && rm goreleaser.tar.gz
