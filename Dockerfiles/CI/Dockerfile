FROM golang:1.12

RUN apt-get update && apt-get install -y --no-install-recommends \
		bzip2 \
        unzip \
    && rm -rf /var/lib/apt/lists/*

ENV PROTOBUF_VER 3.8.0

# NOTE: most of the following assume WORDKIR is '/'
RUN set -ex \
    # Task executor, will be installed in /bin
    && curl -sL https://taskfile.dev/install.sh | sh \
    # Go runtime dependencies
    && go get github.com/golangci/govet \
    && go get golang.org/x/lint/golint \
    # Protobuf tooling
    && go get github.com/golang/protobuf/protoc-gen-go \
    && curl -LO https://github.com/google/protobuf/releases/download/v$PROTOBUF_VER/protoc-$PROTOBUF_VER-linux-x86_64.zip \
    && mkdir protobuf && unzip protoc-$PROTOBUF_VER-linux-x86_64.zip && cp protobuf/bin/* /bin/ && rm -rf protobuf
