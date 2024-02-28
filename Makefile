GO ?= go
GOBIN ?= $$($(GO) env GOPATH)/bin
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.56.2

APP_ID ?= com.solarpunk.swarmmobile
APP_NAME ?= swarm-mobile
BUILD_NUMBER ?= 1
RELEASE ?= false
TARGET_OS ?= android/arm64
APP_VERSION ?= "$(shell git describe --tags --abbrev=0 | cut -c2-)"
COMMIT_HASH ?= "$(shell git describe --long --dirty --always --match "" || true)"

.PHONY: lint
lint: linter
	$(GOLANGCI_LINT) run

.PHONY: linter
linter:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: get-fnye
get-fnye:
	go get fyne.io/fyne/v2/cmd/fyne@latest

.PHONY: package
package:
	fyne package -os ${TARGET_OS} -appID ${APP_ID} -name ${APP_NAME}  -appVersion ${APP_VERSION} -appBuild=${BUILD_NUMBER} -release=${RELEASE} -metadata commithash=${COMMIT_HASH}
