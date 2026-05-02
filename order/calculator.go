package order

import (
	"fmt"

	"github.com/shopspring/decimal"
)

// Calculator runs a fixed sequence of PriceRule values against an Order to
// produce the final price. Rule order matters — each rule sees the price
// produced by the previous rule, so callers should pass rules in the desired
// application order (typically: tax, discounts, rounding last).
type Calculator struct {
	rules           []PriceRule
	zeroPriceAction ZeroPriceAction
}

// CalculatorParams is the plain-data input accepted by NewCalculator. It mirrors Calculator with exported fields so callers (CLI, HTTP, tests) can populate it directly.
type CalculatorParams struct {
	ZeroPriceAction ZeroPriceAction
	Rules           []PriceRule
}

// NewCalculator returns a Calculator that applies rules in the order given.
func NewCalculator(action ZeroPriceAction, rules ...PriceRule) *Calculator {
	return &Calculator{
		zeroPriceAction: action,
		rules:           rules,
	}
}

// Calculate threads the order's BasePrice through every rule and returns the
// final price. It short-circuits on the first rule error.
func (c *Calculator) Calculate(order *Order) (price decimal.Decimal, err error) {
	if order == nil {
		return decimal.Zero, fmt.Errorf("order must not be nil")
	}

	price = order.BasePrice()

	for _, rule := range c.rules {
		price, err = rule.Apply(order, price)
		if err != nil {
			return decimal.Zero, err
		}

		if price.LessThanOrEqual(decimal.Zero) {
			curPrice, resume, err := c.zeroPriceAction(order, price)
			if err != nil {
				return decimal.Zero, err
			}

			if !resume {
				return curPrice, nil
			}
		}
	}

	return
}

// ZeroPriceAction is invoked by Calculator after any rule drives the running
// price to zero or below. It receives the order and the offending price, and
// returns:
//
//   - price:  the price the calculator should use going forward (or as the
//     final result, if resume is false);
//   - resume: if true, calculation continues with the next rule using price;
//     if false, calculator stops immediately and returns price;
//   - err:    a non-nil error aborts calculation and is propagated to the
//     caller.
//
// Use a custom ZeroPriceAction to implement policies like "clamp to zero",
// "log and continue", or "treat as free order". Use DefaultZeroPriceAction to
// reject zero/negative prices outright.
type ZeroPriceAction func(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, bool, error)

// DefaultZeroPriceAction is the strict default: any rule that drives the price
// to zero or below is treated as a configuration error and aborts calculation.
func DefaultZeroPriceAction(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, bool, error) {
	return decimal.Zero, false, fmt.Errorf("price must be positive after applying rule")
}
