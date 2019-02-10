FROM golang:1.11.3-stretch
COPY . /go/src/arduino-cli
WORKDIR /go/src/arduino-cli
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-s -w -extldflags "-static"' .

FROM frolvlad/alpine-glibc
RUN apk add ca-certificates python
WORKDIR /root
COPY --from=0 /go/bin/arduino-cli /usr/bin/arduino-cli 
ENV USER root
COPY dot-cli-config.yml /usr/bin/.cli-config.yml
RUN arduino-cli core update-index --debug
RUN arduino-cli core install esp8266:esp8266
RUN arduino-cli board listall
RUN arduino-cli sketch new blink --debug
COPY blink.ino /root/Arduino/blink/blink.ino
RUN arduino-cli compile --fqbn esp8266:esp8266:nodemcuv2 Arduino/blink
