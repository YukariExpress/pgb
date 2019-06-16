BINS = pgb

DOCKER = docker

.PHONY: bins $(BINS) tag_latest version

bins: $(BINS)

VERSION := $(shell git describe --tags --always --dirty)

$(BINS):
	$(DOCKER) build --rm -f Dockerfile -t $@:$(VERSION) .

version:
	@echo $(VERSION)

tag_latest:
	for bin in $(BINS); do \
		$(DOCKER) tag $$bin:$(VERSION) $$bin:latest; \
	done
