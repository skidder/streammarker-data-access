GO15VENDOREXPERIMENT=1

COVERAGEDIR = ./coverage
all: clean build test cover

clean: 
	if [ -d $(COVERAGEDIR) ]; then rm -rf $(COVERAGEDIR); fi
	if [ -d bin ]; then rm -rf bin; fi

all: build test

build:
	if [ ! -d bin ]; then mkdir bin; fi
	go build -v -o bin/streammarker-data-access

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

docker-build:
	docker info
	docker build -t skidder/streammarker-data-access:latest .

docker-deploy:
	docker login -e ${DOCKER_EMAIL} -u ${DOCKER_USER} -p ${DOCKER_PASS}
	docker push skidder/streammarker-data-access:latest
