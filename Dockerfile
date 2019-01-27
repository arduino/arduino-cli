FROM golang:1.11.3-stretch
COPY . /go/src/arduino-cli
WORKDIR /go/src/arduino-cli
RUN go get .
RUN CGO_ENABLED=0 GOOS=linux go install -a -ldflags '-s -w -extldflags "-static"' .
ENTRYPOINT ["/bin/cp", "-v", "/go/bin/arduino-cli", "/out"]
