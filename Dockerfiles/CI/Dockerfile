FROM golang:1.12

RUN apt-get update && apt-get install -y --no-install-recommends \
    bzip2 \
    unzip \
    && rm -rf /var/lib/apt/lists/*

ENV PROTOBUF_VER 3.9.1
ENV PATH="/miniconda/bin:${PATH}"

# NOTE: most of the following assume WORDKIR is '/'
RUN set -ex \
    # Task executor, will be installed in /bin
    && curl -sL https://taskfile.dev/install.sh | sh \
    # Codecov uploader
    && curl -o /bin/codecov -LO https://codecov.io/bash && chmod +x /bin/codecov \
    # Go runtime dependencies
    && go get github.com/golangci/govet \
    && go get golang.org/x/lint/golint \
    # Protobuf tooling
    && go get github.com/golang/protobuf/protoc-gen-go \
    && mkdir protobuf && cd protobuf \
    && curl -LO https://github.com/google/protobuf/releases/download/v$PROTOBUF_VER/protoc-$PROTOBUF_VER-linux-x86_64.zip \
    && unzip protoc-$PROTOBUF_VER-linux-x86_64.zip \
    && cp ./bin/* /bin/  \
    # protoc will search for default includes in the path of the binary
    && cp -r ./include /bin/ \
    && cd .. &&  rm -rf protobuf \
    # Install a recent version of Python
    && curl -o $HOME/miniconda.sh -LO https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh \
    && chmod +x $HOME/miniconda.sh  && bash ~/miniconda.sh -b -p /miniconda && rm -f $HOME/miniconda.sh