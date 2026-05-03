# Part 1: Coding Task

A small CLI that builds an order from flags, runs it through a configurable
pricing-rule chain, and prints the final price.

## How to run

The binary lives at `cmd/price_calculator`.

```sh
# directly
go run ./cmd/price_calculator --price 199.99 --country TH --customer-type vip --first-order

# or build first
go build -o bin/price_calculator ./cmd/price_calculator
./bin/price_calculator -p 199.99 -c TH -t vip -f
```

### Flags

| Flag              | Alias | Default   | Description                                |
| ----------------- | ----- | --------- | ------------------------------------------ |
| `--price`         | `-p`  | `0`       | Base price of the order                    |
| `--country`       | `-c`  | `TH`      | ISO-3166-1 alpha-2 country code (e.g. TH)  |
| `--customer-type` | `-t`  | `regular` | Customer tier: `regular` or `vip`          |
| `--first-order`   | `-f`  | `false`   | Apply the first-order discount             |

## Configuration (.env)

Pricing parameters come from environment variables (loaded via
[`envconfig`](https://github.com/kelseyhightower/envconfig)). A starter file
ships at `./.env.example` — copy it and source it before running:

```sh
cp .env.example .env
set -a; source .env; set +a
go run ./cmd/price_calculator -p 100 -c TH
```

### Variables

| Variable               | Type                       | Default                  | Meaning |
| ---------------------- | -------------------------- | ------------------------ | ------- |
| `TAXES`                | `country:rate,...`         | `TH:0.07,FR:0.20`        | Per-country tax rates as fractions in `[0, 1]`. |
| `DEFAULT_TAX_RATE`     | decimal                    | `0.2`                    | Fallback rate for countries missing from `TAXES`. **Set to a negative value (e.g. `-1`) to disable the fallback** — unknown countries will then return an error instead. |
| `FIRST_ORDER_DISCOUNT` | decimal in `[0, 1]`        | `0.1`                    | Percentage discount on a customer's first order (e.g. `0.1` = 10% off). |
| `CUSTOMER_DISCOUNTS`   | `tier:amount,...`          | `regular:0,vip:100`      | Flat per-tier discount **subtracted** from the price. Every `CustomerType` value used in production must have an entry. |
| `ROUND_PRECISION`      | int (≥ 0)                  | `2`                      | Decimal places to round up to (ceiling). A negative value falls back to `2`. |

### Rule chain order

The CLI composes the chain in this fixed order; ordering is significant:

```
TaxPriceRule → FirstOrderDiscountPriceRule → CustomerDiscountPriceRule → RoundPriceRule
```

If any rule drives the running price to zero or below, the configured
`ZeroPriceAction` runs. The CLI uses `DefaultZeroPriceAction`, which aborts
with an error.

### Worked example

With the defaults from `.env.example` and `-p 199.99 -c TH -t vip -f`:

```
199.99                       (base)
  → * 1.07     = 213.9893    (TH tax)
  → * (1-0.10) = 192.59037   (first-order discount)
  → -    100   =  92.59037   (vip tier discount)
  → roundUp(2) =  92.60      (final)
```

## Running the tests

```sh
go test ./...                  # everything
go test ./order/ -v            # unit + integration tests for the pricing engine
go test ./order/ -run TestCalculator   # integration tests only
```

Integration tests live in `order/calculator_integration_test.go` and use the
external `order_test` package (black-box). The rationale for which scenarios
are covered (and which are intentionally not) is documented in that file's
package comment.
