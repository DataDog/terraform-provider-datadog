TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=datadog
DIR=~/.terraform.d/plugins
ZORKIAN_VERSION=master
API_CLIENT_VERSION=master

default: build

build: fmtcheck
	go install

install: fmtcheck
	mkdir -vp $(DIR)
	go build -o $(DIR)/terraform-provider-datadog

uninstall:
	@rm -vf $(DIR)/terraform-provider-datadog

test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4
	DD_API_KEY=fake DD_APP_KEY=fake RECORD=false TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout=15m

testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

cassettes: fmtcheck
	RECORD=true TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"


test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

update-go-client:
	echo "Updating the Zorkian client to ${ZORKIAN_VERSION} and the API Client to ${API_CLIENT_VERSION}"
	go get github.com/zorkian/go-datadog-api@$(ZORKIAN_VERSION)
	go get github.com/DataDog/datadog-api-client-go@${API_CLIENT_VERSION}
	go mod vendor
	go mod tidy

.PHONY: build test testacc cassettes vet fmt fmtcheck errcheck test-compile
