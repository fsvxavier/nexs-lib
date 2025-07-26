# Makefile for PostgreSQL module

.PHONY: mocks test test-unit test-integration coverage clean

# Generate mocks for all interfaces
mocks:
	@echo "Generating mocks..."
	go generate ./db/postgresql/providers/pgx/mocks/
	@echo "Mocks generated successfully!"

# Run all tests
test: test-unit test-integration

# Run unit tests with mocks
test-unit:
	@echo "Running unit tests..."
	go test -v -tags=unit -timeout=30s ./db/postgresql/...

# Run integration tests (requires database)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration -timeout=30s ./db/postgresql/...

# Generate test coverage report
coverage:
	@echo "Generating coverage report..."
	go test -v -tags=unit -coverprofile=coverage.out ./db/postgresql/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f coverage.out coverage.html
	find . -name "*_mock.go" -delete
	@echo "Clean completed!"

# Install required tools
tools:
	@echo "Installing required tools..."
	go install github.com/golang/mock/mockgen@latest
	@echo "Tools installed successfully!"


#=================================================================================================================

genmock:
	mockery --all --keeptree --dir=internal --output=./internal/mocks
.PHONY: genmock

cover:	
	go test ./... -coverprofile coverange/cover.out
	go tool cover -html=coverange/cover.out
.PHONY: cover

genproto:
	protoc --proto_path=internal/pb/proto  --go_out=internal/pb/gen --go_opt=paths=source_relative --go-grpc_out=require_unimplemented_servers=false:internal/pb/gen --go-grpc_opt=paths=source_relative internal/pb/proto/*.proto
.PHONY: genproto


runbench:
	go test -bench=. -count 5
.PHONY: runbench

#go install github.com/xo/xo@latest
xo-models:
	~/go/bin/xo --src my-tpl schema $(DB_URL)
.PHONY: xo-models

#go install github.com/go-jet/jet/v2/cmd/jet@latest
gen-models:
	~/go/bin/jet -dsn=postgres://docker:docker@localhost:5432/docker?sslmode=disable -schema=hms -path=./gen
.PHONY: gen-models

#go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec:
	~/go/bin/gosec -exclude-dir=rules -exclude-dir=vendor -fmt=json -out=results.json -stdout ./...
.PHONY: gosec

#go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
gocritic:
	~/go/bin/gocritic check -enableAll -disable='#experimental' ./...
.PHONY: gocritic

#go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golint-html:
	~/go/bin/golangci-lint --enable=gocritic --out-format=html run ./... >> ./lint/lint.html
.PHONY: golint-html

golint-xml:
	~/go/bin/golangci-lint --enable=gocritic --out-format=checkstyle run ./... >> ./lint/lint.xml
.PHONY: golint-xml

golint-local:
	~/go/bin/golangci-lint --enable=gocritic run ./...
.PHONY: golint-local

golint-getlinters:
	~/go/bin/golangci-lint help linters
.PHONY: golint-getlinters

#go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
gofieldalt:
	~/go/bin/fieldalignment -fix ./... 
.PHONY: gofieldalt

#go install github.com/4meepo/tagalign/cmd/tagalign@latest
tagalign:
	~/go/bin/tagalign -fix ./...
.PHONY: tagalign

#go install github.com/vburenin/ifacemaker@latest
gen-interfaces:
	go generate ./...
.PHONY: gen-interfaces

#go install golang.org/x/vuln/cmd/govulncheck@latest
vuln:
	~/go/bin/govulncheck ./...
.PHONY: vuln

#go install github.com/google/osv-scanner/cmd/osv-scanner@v1
osv-scanner:
	~/go/bin/osv-scanner -r ./...
.PHONY: osv-scanner

run-race:
	@GORACE="log_path={$PWD}/race_report.txt" go run -race main.go
.PHONY: run-race


go-gen:
	go generate ./...
.PHONY: go-gen
