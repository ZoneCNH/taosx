#!/usr/bin/env bash
set -euo pipefail

echo "checking forbidden dependency on x.go..."

if go list -deps ./... | grep -q "github.com/bytechainx/x.go"; then
  echo "ERROR: base library template must not depend on x.go"
  exit 1
fi

echo "checking forbidden business terms..."

FORBIDDEN_TERMS=(
  "MacroRegime"
  "MarketRegime"
  "TradingSignal"
  "BTCUSDT"
  "ETHUSDT"
  "Kline"
  "OrderBook"
  "Position"
  "RiskGate"
)

for term in "${FORBIDDEN_TERMS[@]}"; do
  if grep -R "$term" ./pkg ./internal --exclude-dir=.git; then
    echo "ERROR: forbidden business term found: $term"
    exit 1
  fi
done

echo "boundary check passed"
