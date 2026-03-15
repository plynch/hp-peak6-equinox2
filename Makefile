.DEFAULT_GOAL := help

.PHONY: help dev dev-live-epl dev-live-fed verify test fixture-demo list-clusters live-inspect live-epl live-fed scan showcase demo-cli route-order clean

CLUSTER ?=
EVENT_QUERY ?=
PROP_QUERY ?=
SIDE ?= buy_yes
LIMIT ?= 0.60
SIZE ?= 1000
LIVE_LIMIT ?= 1
LIVE_MATCHWEEKS ?= 4
FED_MEETINGS ?= 4
ROUTEABLE_ONLY ?= 0
SOURCE ?= fixture

help:
	@echo "Available targets:"
	@echo "  make dev                              # start local web UI at http://127.0.0.1:8080"
	@echo "  make dev-live-epl LIVE_MATCHWEEKS=4   # start the web UI backed by live EPL data"
	@echo "  make dev-live-fed FED_MEETINGS=4      # start the web UI backed by live Fed data"
	@echo "  make demo-cli                         # run the full terminal showcase across fixture + live-fed + live-epl"
	@echo "  make verify                           # run tests and fixture demo"
	@echo "  make scan SOURCE=fixture              # source-aware terminal scan (fixture, live-fed, live-epl, all-live)"
	@echo "  make showcase                         # run fixture + live-fed + live-epl terminal demo"
	@echo "  make list-clusters ROUTEABLE_ONLY=1   # inspect current routeable proposition clusters"
	@echo "  make route-order CLUSTER=prop-001"
	@echo "  make route-order EVENT_QUERY='fomc march 2026' PROP_QUERY='fed hike rate march meeting' LIMIT=0.60"
	@echo "  make route-order EVENT_QUERY='liverpool vs tottenham' PROP_QUERY='liverpool win' LIMIT=0.76"
	@echo "  make scan SOURCE=live-fed FED_MEETINGS=2          # prints current live selector-ready commands"
	@echo "  make scan SOURCE=live-epl LIVE_MATCHWEEKS=2       # prints current live selector-ready commands"
	@echo "  make scan SOURCE=all-live FED_MEETINGS=2 LIVE_MATCHWEEKS=1   # prints every currently routeable live event"
	@echo "  make live-inspect LIVE_LIMIT=1        # optional public API ingestion check"
	@echo "  make live-epl LIVE_MATCHWEEKS=4       # fetch current + next EPL matchweek windows and route all routeable propositions"
	@echo "  make live-fed FED_MEETINGS=4          # fetch current + next Fed meetings and route all routeable propositions"
	@echo "  make clean"

dev:
	go run ./cmd/equinox serve

dev-live-epl:
	go run ./cmd/equinox serve --source live-epl --matchweeks $(LIVE_MATCHWEEKS)

dev-live-fed:
	go run ./cmd/equinox serve --source live-fed --meetings $(FED_MEETINGS)

verify: test fixture-demo

test:
	go test ./...

fixture-demo:
	go run ./cmd/equinox fixture-demo

scan:
	go run ./cmd/equinox scan --source $(SOURCE) --matchweeks $(LIVE_MATCHWEEKS) --meetings $(FED_MEETINGS)

showcase:
	go run ./cmd/equinox showcase --matchweeks $(LIVE_MATCHWEEKS) --meetings $(FED_MEETINGS)

demo-cli: showcase

list-clusters:
	go run ./cmd/equinox list-clusters --source $(SOURCE) --matchweeks $(LIVE_MATCHWEEKS) --meetings $(FED_MEETINGS) $(if $(filter 1,$(ROUTEABLE_ONLY)),--routeable-only,)

live-inspect:
	go run ./cmd/equinox live-inspect --limit $(LIVE_LIMIT)

live-epl:
	go run ./cmd/equinox live-epl --matchweeks $(LIVE_MATCHWEEKS)

live-fed:
	go run ./cmd/equinox live-fed --meetings $(FED_MEETINGS)

route-order:
	@if [ -z "$(CLUSTER)" ] && [ -z "$(EVENT_QUERY)" ] && [ -z "$(PROP_QUERY)" ]; then \
		echo "Provide CLUSTER=prop-001 or EVENT_QUERY='...' plus PROP_QUERY='...'. Run 'make list-clusters ROUTEABLE_ONLY=1 SOURCE=$(SOURCE)' first."; \
		exit 1; \
	fi
	go run ./cmd/equinox route-order --source $(SOURCE) --matchweeks $(LIVE_MATCHWEEKS) --meetings $(FED_MEETINGS) $(if $(CLUSTER),--cluster "$(CLUSTER)",) $(if $(EVENT_QUERY),--event-query "$(EVENT_QUERY)",) $(if $(PROP_QUERY),--prop-query "$(PROP_QUERY)",) --side $(SIDE) --limit $(LIMIT) --size $(SIZE)

clean:
	rm -rf artifacts equinox.db
