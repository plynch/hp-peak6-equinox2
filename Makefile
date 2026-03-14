.DEFAULT_GOAL := help

.PHONY: help dev verify test fixture-demo live-inspect route-order clean

CLUSTER ?= prop-001
SIDE ?= buy_yes
LIMIT ?= 0.60
SIZE ?= 1000
LIVE_LIMIT ?= 1

help:
	@echo "Available targets:"
	@echo "  make dev                              # start local web UI at http://127.0.0.1:8080"
	@echo "  make verify                           # run tests and fixture demo"
	@echo "  make route-order                      # route default buy_yes order for prop-001"
	@echo "  make route-order SIDE=sell_yes LIMIT=0.55"
	@echo "  make live-inspect LIVE_LIMIT=1        # optional public API ingestion check"
	@echo "  make clean"

dev:
	go run ./cmd/equinox serve

verify: test fixture-demo

test:
	go test ./...

fixture-demo:
	go run ./cmd/equinox fixture-demo

live-inspect:
	go run ./cmd/equinox live-inspect --limit $(LIVE_LIMIT)

route-order:
	go run ./cmd/equinox route-order --cluster $(CLUSTER) --side $(SIDE) --limit $(LIMIT) --size $(SIZE)

clean:
	rm -rf artifacts equinox.db
