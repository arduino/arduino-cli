GO = go
GOGET = $(GO) get -u

all: check lint

check:
	$(GO) test -cover ./...

install:
	$(GO) install -v ./...

clean:
	$(GO) clean -x ./...
	rm -rvf ./.test*

lint:
	gofmt -l -e -s . || :
	go vet . || :
	golint . || :
	gocyclo -over 15 . || :
	misspell ./* || :

deps:	
	$(GOGET) github.com/golang/lint/golint
	$(GOGET) github.com/fzipp/gocyclo
	$(GOGET) github.com/client9/misspell/cmd/misspell

.PHONY: all check install clean lint deps
