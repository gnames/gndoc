VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')

FLAG_MODULE = GO111MODULE=on
FLAGS_SHARED = $(FLAG_MODULE) GOARCH=amd64
NO_C = CGO_ENABLED=0
FLAGS_LINUX = $(FLAGS_SHARED) GOOS=linux
FLAGS_MAC = $(FLAGS_SHARED) GOOS=darwin
FLAGS_WIN = $(FLAGS_SHARED) GOOS=windows
FLAGS_LD=-ldflags "-s -w -X github.com/gnames/gnfinder-doc.Build=${DATE} \
                  -X github.com/gnames/gnfinder-doc.Version=${VERSION}"
GOCMD = go
GOBUILD = $(GOCMD) build $(FLAGS_LD)
GOINSTALL = $(GOCMD) install $(FLAGS_LD)
GOCLEAN = $(GOCMD) clean
GOGET = $(GOCMD) get

RELEASE_DIR ?= "/tmp"
BUILD_DIR ?= "."
CLIB_DIR ?= "."

all: install

test: deps install
	$(FLAG_MODULE) go test -race ./...

test-build: deps build

deps:
	$(GOCMD) mod download;

tools: deps
	@echo Installing tools from tools.go
	@cat gnfinder-doc/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

build:
	cd gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) $(NO_C) $(GOBUILD) -o $(BUILD_DIR) 

install:
	cd gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) $(NO_C) $(GOINSTALL)

release: peg dockerhub
	cd gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_LINUX) $(NO_C) $(GOBUILD); \
	tar zcf $(RELEASE_DIR)/gnfinder-doc-$(VER)-linux.tar.gz gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_MAC) $(NO_C) $(GOBUILD); \
	tar zcf $(RELEASE_DIR)/gnfinder-doc-$(VER)-mac.tar.gz gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_WIN) $(NO_C) $(GOBUILD); \
	zip -9 $(RELEASE_DIR)/gnfinder-doc-$(VER)-win-64.zip gnfinder-doc.exe; \
	$(GOCLEAN);

nightly: peg
	cd gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_LINUX) $(NO_C) $(GOBUILD); \
	tar zcf $(RELEASE_DIR)/gnfinder-doc-linux.tar.gz gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_MAC) $(NO_C) $(GOBUILD); \
	tar zcf $(RELEASE_DIR)/gnfinder-doc-mac.tar.gz gnfinder-doc; \
	$(GOCLEAN); \
	$(FLAGS_WIN) $(NO_C) $(GOBUILD); \
	zip -9 $(RELEASE_DIR)/gnfinder-doc-win-64.zip gnfinder-doc.exe; \
	$(GOCLEAN);

dc: asset build
	docker-compose build;

docker: build
	docker build -t gnames/gnfinder-doc:latest -t gnames/gnfinder-doc:$(VERSION) .; \
	cd gnfinder-doc; \
	$(GOCLEAN);

dockerhub: docker
	docker push gnames/gnfinder-doc; \
	docker push gnames/gnfinder-doc:$(VERSION)
