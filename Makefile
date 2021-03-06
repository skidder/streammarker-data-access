GO15VENDOREXPERIMENT=1

COVERAGEDIR = ./coverage
all: clean build test cover

clean: 
	if [ -d $(COVERAGEDIR) ]; then rm -rf $(COVERAGEDIR); fi
	if [ -d streammarker-data-access ]; then rm -f streammarker-data-access; fi

all: build test

build:
	if [ ! -d bin ]; then mkdir bin; fi
	go build -v -o streammarker-data-access

static-build:
	CGO_ENABLED=0 go build -a -ldflags '-s' -installsuffix cgo -v -o streammarker-data-access

fmt:
	go fmt ./...

test:
	if [ ! -d $(COVERAGEDIR) ]; then mkdir $(COVERAGEDIR); fi
	go test -v ./handlers -race -cover -coverprofile=$(COVERAGEDIR)/handlers.coverprofile

cover:
	go tool cover -html=$(COVERAGEDIR)/handlers.coverprofile -o $(COVERAGEDIR)/handlers.html

bench:
	go test ./... -cpu 2 -bench .

run: build
	$(CURDIR)/streammarker-data-access
