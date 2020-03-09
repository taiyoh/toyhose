GIT_VERSION=$(shell git describe --tags)
CURRENT_REVISION=$(shell git rev-parse --short HEAD)
LDFLAGS="-s -w -X main.version=$(GIT_VERSION) -X main.commit=$(CURRENT_REVISION)"

.PHONY: build release docker clean

clean:
	rm -rf pkg/$(GIT_VERSION)/ && mkdir -p pkg/$(GIT_VERSION)/dist

pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose -ldflags=$(LDFLAGS) cmd/toyhose/main.go

pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose_$(GIT_VERSION)_darwin_amd64.zip: pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose
	cd pkg/$(GIT_VERSION)/toyhose_darwin_amd64  && zip toyhose_$(GIT_VERSION)_darwin_amd64.zip * && mv toyhose_$(GIT_VERSION)_darwin_amd64.zip ../dist

pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose -ldflags=$(LDFLAGS) cmd/toyhose/main.go

pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose_$(GIT_VERSION)_linux_amd64.tar.gz: pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose
	cd pkg/$(GIT_VERSION)/toyhose_linux_amd64   && tar cvzf toyhose_$(GIT_VERSION)_linux_amd64.tar.gz toyhose && mv toyhose_$(GIT_VERSION)_linux_amd64.tar.gz ../dist

build: clean pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose_$(GIT_VERSION)_linux_amd64.tar.gz pkg/$(GIT_VERSION)/toyhose_darwin_amd64/toyhose_$(GIT_VERSION)_darwin_amd64.zip

release:
	ghr -n $(GIT_VERSION) $(GIT_VERSION) pkg/$(GIT_VERSION)/dist

docker: clean pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose
	mv pkg/$(GIT_VERSION)/toyhose_linux_amd64/toyhose /bin/toyhose
