GO_BIN ?= go
CURL_BIN ?= curl
SHELL_BIN ?= sh

deps: check-gopath
	$(GO_BIN) get -u github.com/mafredri/cdp
	$(GO_BIN) get -u github.com/mafredri/cdp/devtool
	$(GO_BIN) get -u github.com/mafredri/cdp/protocol/css
	$(GO_BIN) get -u github.com/mafredri/cdp/protocol/emulation
	$(GO_BIN) get -u github.com/mafredri/cdp/protocol/page
	$(GO_BIN) get -u github.com/mafredri/cdp/protocol/security
	$(GO_BIN) get -u github.com/mafredri/cdp/rpcc

	$(GO_BIN) get -u github.com/garyhouston/jpegsegs

	# linters
	$(CURL_BIN) -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | $(SHELL_BIN) -s -- -b ${GOPATH}/bin v1.15.0
	$(GO_BIN) get -u github.com/Quasilyte/go-consistent

lint:
	golangci-lint run
	go-consistent -v ./...

check-gopath:
ifndef GOPATH
	$(error GOPATH is undefined)
endif
