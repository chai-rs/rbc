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
	rules []PriceRule
}

// NewCalculator returns a Calculator that applies rules in the order given.
func NewCalculator(rules ...PriceRule) *Calculator {
	return &Calculator{rules: rules}
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
	}

	return
}
