GIT_VERSION=$(shell git describe --tags)
CURRENT_REVISION=$(shell git rev-parse --short HEAD)
LDFLAGS="-s -w -X main.version=$(GIT_VERSION) -X main.commit=$(CURRENT_REVISION)"

.PHONY: test install

install:
	cd cmd/toyhose && go install

test:
	docker-compose up -d
	go test -v -race -timeout 30s
	docker-compose down

build:
	rm -rf pkg/$(GIT_VERSION)/ && mkdir -p pkg/$(GIT_VERSION)/dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose       -ldflags=$(LDFLAGS) cmd/toyhose/main.go
	cd pkg/$(GIT_VERSION)/toyhose_linux_amd64   && tar cvzf toyhose_$(GIT_VERSION)_linux_amd64.tar.gz toyhose && mv toyhose_$(GIT_VERSION)_linux_amd64.tar.gz ../dist
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose      -ldflags=$(LDFLAGS) cmd/toyhose/main.go
	cd pkg/$(GIT_VERSION)/toyhose_darwin_amd64  && zip toyhose_$(GIT_VERSION)_darwin_amd64.zip * && mv toyhose_$(GIT_VERSION)_darwin_amd64.zip ../dist

release:
	ghr -b "$(shell ghch --format=markdown --latest)" -n $(GIT_VERSION) $(GIT_VERSION) pkg/$(GIT_VERSION)/dist
