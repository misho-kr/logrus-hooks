# ---------------------------------------------------------------------

SRC	?= $(PWD)

GO_CMD	?= go
GOFMT	?= gofmt
GOLINT	?= golint

CGO 	?=

ifneq ($(CGO),)
  GO_CMD := CGO_ENABLED=$(CGO) $(GO_CMD) 
endif

COVERAGE           ?= no
COVERAGE_OPTION    ?= text
COVERAGE_REPORTS   ?= $(SRC)/coverage-report

GOFMT_OPTIONS  	   ?= -l
GOFMT_REPORT_LINES ?= 10
GOFMT_REPORTS 	   ?= $(SRC)/gofmt-report

GOCLEAN_OPTIONS    ?= -cache

VERBOSE_FLAG = 
ifeq ($(VERBOSE), yes)
  VERBOSE_FLAG := -v
endif

BUILD_DATETIME  := $(shell date +"%Y%m%d-%H%M%S")

GOFMT_REPORT 	?= $(GOFMT_REPORTS).$(BUILD_DATETIME).txt
COVERAGE_REPORT ?= $(COVERAGE_REPORTS).$(BUILD_DATETIME).txt

# ---------------------------------------------------------------------

.PHONY: dev all travis
.PHONY: clean build test codecov coverage vet lint format
.PHONY:	show_coverage doc

dev: 	vet test build
all:  format lint vet build
ci:  	all coverage
github: all codecov
travis: all codecov

# ---------------------------------------------------------------------

build:
	$(call announce,go $@)
	@$(GO_CMD) build $(VERBOSE_FLAG) ./...

test:
	$(call announce,go $@)
	@$(GO_CMD) test $(VERBOSE_FLAG) ./...

coverage:
	$(call announce,go $@ -> $(COVERAGE_REPORT))
	@$(GO_CMD) test -coverprofile=$(COVERAGE_REPORT) ./...
	@$(GO_CMD) tool cover -func="$(COVERAGE_REPORT)"

codecov:
	$(call announce,go $@)
	@$(GO_CMD) test -coverprofile=coverage.txt -race -covermode=atomic

vet:
	$(call announce,go $@)
	@$(GO_CMD) vet $(VERBOSE_FLAG) ./...

lint:
	$(call announce,go $@)
	@$(GOLINT) -set_exit_status ./...

format:
	$(call announce,go $@)
	@$(GOFMT) $(GOFMT_OPTIONS) . > $(GOFMT_REPORT)
	@if [ -s "$(GOFMT_REPORT)" ]; then \
		if [ "$(GOFMT_OPTIONS)" = "-l" ]; then \
			echo "the following files failed the formatting checks:"; \
			while read f; do echo "  - $${f#$(SRC)/}"; done < $(GOFMT_REPORT); \
			echo; echo "add GOFMT_OPTIONS=\"-d\" option to display the offending lines"; echo; \
			rm $(GOFMT_REPORT); \
		else \
			echo "showing first $(GOFMT_REPORT_LINES) lines of go-format report: $(GOFMT_REPORT)"; echo; \
			head -$(GOFMT_REPORT_LINES) $(GOFMT_REPORT); \
			echo "..."; echo; \
		fi; \
		false; \
	else \
		rm $(GOFMT_REPORT); \
	fi

clean:
	$(call announce,go $@)
	@rm -f $(GOFMT_REPORTS).* $(COVERAGE_REPORTS).* coverage.txt
	@$(GO_CMD) clean $(GOCLEAN_OPTION)

# misc targets --------------------------------------------------------

show_coverage:
ifeq ($(COVERAGE_FORMAT), html)
	@$(GO_CMD) tool cover -html="$(COVERAGE_REPORT)"
else
	@$(GO_CMD) tool cover -func="$(COVERAGE_REPORT)"
endif

doc:
	godoc -http=:8080 -index

# ---------------------------------------------------------------------

define announce
  @echo "# $(1)"; echo
endef
