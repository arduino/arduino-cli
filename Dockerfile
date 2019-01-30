FROM golang:1.11.3-stretch
COPY . /go/src/arduino-cli
WORKDIR /go/src/arduino-cli
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-s -w -extldflags "-static"' .
WORKDIR /root
ENV USER root
COPY dot-cli-config.yml /go/bin/.cli-config.yml
RUN arduino-cli core update-index --debug
RUN arduino-cli core install esp8266:esp8266
RUN arduino-cli board listall
