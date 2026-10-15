GO ?= go
TERRAFORM ?= terraform
GOCACHE ?= /tmp/tfkit-go/cache
GOPATH ?= /tmp/tfkit-go/path
GOMODCACHE ?= /tmp/tfkit-go/mod
export GOCACHE GOPATH GOMODCACHE

.PHONY: fmt check test build terraform-check consumer command-smoke acceptance ci clean

fmt:
	@files="$$(gofmt -l .)"; test -z "$$files" || { echo "Go files need formatting:"; echo "$$files"; exit 1; }
	$(TERRAFORM) fmt -check -recursive

check: fmt
	$(GO) vet ./...

test:
	$(GO) test -race ./...

build:
	mkdir -p bin
	$(GO) build -trimpath -o bin/tfcheck ./cmd/tfcheck
	$(GO) build -trimpath -o bin/tfplan ./cmd/tfplan

terraform-check:
	$(TERRAFORM) init -backend=false -input=false
	$(TERRAFORM) validate
	$(TERRAFORM) test
	$(TERRAFORM) plan -input=false -out=/tmp/tfkit-plan.tfplan -var environment=prod -var service=billing-api
	$(TERRAFORM) show -json /tmp/tfkit-plan.tfplan > /tmp/tfkit-plan.json
	./bin/tfplan /tmp/tfkit-plan.json
	rm -f /tmp/tfkit-plan.tfplan /tmp/tfkit-plan.json
	$(TERRAFORM) -chdir=examples/environment init -backend=false -input=false
	$(TERRAFORM) -chdir=examples/environment validate

consumer: build
	$(TERRAFORM) -chdir=examples/consumer init -backend=false -input=false
	$(TERRAFORM) -chdir=examples/consumer validate
	$(TERRAFORM) -chdir=examples/consumer plan -input=false -lock=false -out=/tmp/tfkit-consumer.tfplan
	$(TERRAFORM) -chdir=examples/consumer show -json /tmp/tfkit-consumer.tfplan > /tmp/tfkit-consumer.json
	./bin/tfplan /tmp/tfkit-consumer.json
	rm -f /tmp/tfkit-consumer.tfplan /tmp/tfkit-consumer.json

command-smoke: build
	./bin/tfcheck .
	./bin/tfplan tests/fixtures/plans/safe.json
	@if ./bin/tfplan tests/fixtures/plans/destructive.json; then echo "destructive fixture unexpectedly passed"; exit 1; else test $$? -eq 1; fi
	./bin/tfplan -approve-destructive tests/fixtures/plans/destructive.json
	@if ./bin/tfplan tests/fixtures/plans/missing.json; then echo "missing fixture unexpectedly passed"; exit 1; else test $$? -eq 2; fi

acceptance: command-smoke terraform-check consumer

ci: check test acceptance

clean:
	rm -f bin/tfcheck bin/tfplan
