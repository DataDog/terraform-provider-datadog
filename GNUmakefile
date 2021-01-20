TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
PKG_NAME=datadog
DIR=~/.terraform.d/plugins
ZORKIAN_VERSION=master
API_CLIENT_VERSION=master

# Local variables for installing the plugin to a local
# plugin mirror, used for manual build/testing with terraform 0.13
VERSION=0.0.1
LOCAL_PROVIDERS="$$HOME/.terraform.d/plugins_local"
BINARY_PATH="registry.terraform.io/datadog/datadog/${VERSION}/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-datadog_v${VERSION}"


default: build

build: fmtcheck
	go install

# Builds the provider and adds it to an independently configured filesystem_mirror folder.
# Taken from - https://github.com/hashicorp/terraform/issues/25906#issuecomment-676495452
build_013:
	@echo "Please configure your .terraformrc file to contain a filesystem_mirror block pointed at '${LOCAL_PROVIDERS}' for 'registry.terraform.io/Datadog/datadog'"
	@echo "You MUST delete existing cached plugins from any .terraform directories in Terraform installations you want to test against so that it will perform a lookup on the local mirror"
	go build -o "${LOCAL_PROVIDERS}/${BINARY_PATH}"

install: fmtcheck
	mkdir -vp $(DIR)
	go build -o $(DIR)/terraform-provider-datadog

uninstall:
	@rm -vf $(DIR)/terraform-provider-datadog

test: get-test-deps fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 gotestsum --format testname -- $(TESTARGS) -timeout=30s -parallel=4
	DD_API_KEY=fake DD_APP_KEY=fake RECORD=false TF_ACC=1 gotestsum --format testname -- $(TEST) -v $(TESTARGS) -timeout=15m

testacc: get-test-deps fmtcheck
	TF_ACC=1 gotestsum --format testname -- $(TEST) -v $(TESTARGS) -timeout 120m

cassettes: get-test-deps fmtcheck
	RECORD=true TF_ACC=1 gotestsum --format testname -- $(TEST) -v $(TESTARGS) -timeout 120m

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


test-compile: get-test-deps
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	gotestsum --format testname -- -c $(TEST) $(TESTARGS)

update-go-client:
	echo "Updating the Zorkian client to ${ZORKIAN_VERSION} and the API Client to ${API_CLIENT_VERSION}"
	go get github.com/zorkian/go-datadog-api@$(ZORKIAN_VERSION)
	go get github.com/DataDog/datadog-api-client-go@${API_CLIENT_VERSION}
	go mod tidy

get-test-deps:
	gotestsum --version || (cd `mktemp -d`;	GO111MODULE=off GOFLAGS='' go get -u gotest.tools/gotestsum; cd -)

tools:
	go generate -tags tools tools/tools.go

.PHONY: build test testacc cassettes vet fmt fmtcheck errcheck test-compile get-test-deps
