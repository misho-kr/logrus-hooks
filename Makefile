# ---------------------------------------------------------------------

SRC	?= $(PWD)

GO_CMD	?= go
GOFMT	?= gofmt
GOLINT	?= golint

COVERAGE           ?= no
COVERAGE_OPTION    ?= text
COVERAGE_REPORTS   ?= $(SRC)/test-coverage-report

GOFMT_OPTIONS  	   ?= -l
GOFMT_REPORT_LINES ?= 10
GOFMT_REPORTS 	   ?= $(SRC)/gofmt-report

GOCLEAN_OPTIONS    ?= -cache

VERBOSE_FLAG = 
ifeq ($(VERBOSE), yes)
  VERBOSE_FLAG := -v
endif

GOFMT_REPORT 	?= $(GOFMT_REPORTS).$(BUILD_DATETIME_TAG).txt
COVERAGE_REPORT ?= $(COVERAGE_REPORTS).$(BUILD_DATETIME_TAG).txt

# ---------------------------------------------------------------------

.PHONY: all pr
.PHONY: clean build test cover vet lint format

all: test build
pr:  format lint vet build coverage

# ---------------------------------------------------------------------

build:
	$(call announce,Go $@)
	@$(GO_CMD) build $(VERBOSE_FLAG) ./...

test:
	$(call announce,Go $@)
	@$(GO_CMD) test $(VERBOSE_FLAG) ./...

coverage:
	$(call announce,Go $@ -> $(COVERAGE_REPORT))
	@$(GO_CMD) test -coverprofile=$(COVERAGE_REPORT) ./...
	@$(GO_CMD) tool cover -func="$(COVERAGE_REPORT)"

vet:
	$(call announce,Go $@)
	@$(GO_CMD) vet $(VERBOSE_FLAG) ./...

lint:
	$(call announce,Go $@)
	@$(GOLINT) -set_exit_status ./...

format:
	$(call announce,Go $@)
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
	$(call announce,Go $@)
	@rm -f $(GOFMT_REPORTS).* $(COVERAGE_REPORTS).*
	@$(GO_CMD) clean $(GOCLEAN_OPTION)

# misc targets --------------------------------------------------------

.PHONY:	show_coverage

show_coverage:
ifeq ($(COVERAGE_FORMAT), html)
	@$(GO_CMD) tool cover -html="$(COVERAGE_REPORT)"
else
	@$(GO_CMD) tool cover -func="$(COVERAGE_REPORT)"
endif

# ---------------------------------------------------------------------

define announce
  @echo "# $(1)"; echo
endef

# ---------------------------------------------------------------------
# eof
#
