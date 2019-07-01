FROM golang:1.12

RUN apt-get update && apt-get install -y --no-install-recommends \
		bzip2 \
    && rm -rf /var/lib/apt/lists/*

RUN set -ex \
    && curl -sL https://taskfile.dev/install.sh | sh