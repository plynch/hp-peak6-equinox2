.DEFAULT_GOAL := help

.PHONY: help dev dev-live-epl verify test fixture-demo list-clusters live-inspect live-epl route-order clean

CLUSTER ?=
SIDE ?= buy_yes
LIMIT ?= 0.60
SIZE ?= 1000
LIVE_LIMIT ?= 1
LIVE_EVENT_LIMIT ?= 20
ROUTEABLE_ONLY ?= 0

help:
	@echo "Available targets:"
	@echo "  make dev                              # start local web UI at http://127.0.0.1:8080"
	@echo "  make dev-live-epl LIVE_EVENT_LIMIT=20 # start the web UI backed by live EPL data"
	@echo "  make verify                           # run tests and fixture demo"
	@echo "  make list-clusters ROUTEABLE_ONLY=1   # inspect current routeable proposition clusters"
	@echo "  make route-order CLUSTER=prop-001"
	@echo "  make route-order CLUSTER=prop-001 SIDE=sell_yes LIMIT=0.55"
	@echo "  make live-inspect LIVE_LIMIT=1        # optional public API ingestion check"
	@echo "  make live-epl LIVE_EVENT_LIMIT=20     # fetch upcoming EPL games and route all routeable propositions"
	@echo "  make clean"

dev:
	go run ./cmd/equinox serve

dev-live-epl:
	go run ./cmd/equinox serve --source live-epl --event-limit $(LIVE_EVENT_LIMIT)

verify: test fixture-demo

test:
	go test ./...

fixture-demo:
	go run ./cmd/equinox fixture-demo

list-clusters:
	go run ./cmd/equinox list-clusters $(if $(filter 1,$(ROUTEABLE_ONLY)),--routeable-only,)

live-inspect:
	go run ./cmd/equinox live-inspect --limit $(LIVE_LIMIT)

live-epl:
	go run ./cmd/equinox live-epl --event-limit $(LIVE_EVENT_LIMIT)

route-order:
	@test -n "$(CLUSTER)" || (echo "CLUSTER is required. Run 'make list-clusters ROUTEABLE_ONLY=1' first."; exit 1)
	go run ./cmd/equinox route-order --cluster $(CLUSTER) --side $(SIDE) --limit $(LIMIT) --size $(SIZE)

clean:
	rm -rf artifacts equinox.db
