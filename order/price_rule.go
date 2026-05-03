package order

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	// DefaultRoundPrecision is the fallback decimal precision used by
	// NewRoundPriceRule when the configured precision is negative.
	DefaultRoundPrecision = 2
)

// Compile-time guarantees that every concrete rule satisfies PriceRule.
var (
	_ PriceRule = (*TaxPriceRule)(nil)
	_ PriceRule = (*FirstOrderDiscountPriceRule)(nil)
	_ PriceRule = (*CustomerDiscountPriceRule)(nil)
	_ PriceRule = (*RoundPriceRule)(nil)
)

// PriceRules is the envconfig-friendly bundle of pricing inputs loaded from the
// process environment. The `default` tags are read by kelseyhightower/envconfig
// when the corresponding env var is unset.
type PriceRules struct {
	Taxes              map[string]decimal.Decimal       `default:"TH:0.07,FR:0.20"`
	FirstOrderDiscount decimal.Decimal                  `split_words:"true" default:"100"`
	CustomerDiscounts  map[CustomerType]decimal.Decimal `split_words:"true" default:"regular:0,vip:100"`
	RoundPrecision     int                              `split_words:"true" default:"2"`
}

// PriceRule transforms currentPrice in light of order. Implementations must be
// pure functions of their inputs so the Calculator can compose them safely.
type PriceRule interface {
	Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error)
}

// TaxPriceRule applies a country-specific multiplicative tax: price * (1 + rate).
type TaxPriceRule struct {
	Taxes map[string]decimal.Decimal
}

// NewTaxPriceRule validates that every key is an ISO-3166-1 alpha-2 code and
// every rate is in [0, 1].
func NewTaxPriceRule(taxes map[string]decimal.Decimal) (*TaxPriceRule, error) {
	for code, rate := range taxes {
		if err := v.Validate(code, is.CountryCode2); err != nil {
			return nil, fmt.Errorf("invalid country code: %s", code)
		}

		if rate.LessThan(decimal.Zero) || rate.GreaterThan(decimal.NewFromInt(1)) {
			return nil, fmt.Errorf("invalid tax rate for country %s: %s", code, rate)
		}
	}

	return &TaxPriceRule{Taxes: taxes}, nil
}

// Apply returns currentPrice grossed up by the tax rate for the order's
// country. It errors when the country has no configured rate.
func (r *TaxPriceRule) Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	taxRate, ok := r.Taxes[order.CountryCode()]
	if !ok {
		return decimal.Zero, fmt.Errorf("tax is not supported for country: %s", order.CountryCode())
	}

	return currentPrice.Add(currentPrice.Mul(taxRate)), nil
}

// FirstOrderDiscountPriceRule applies a percentage discount on the first order.
// Discount is interpreted as a fraction (e.g. 0.10 = 10% off).
type FirstOrderDiscountPriceRule struct {
	DiscountPercent decimal.Decimal
}

// NewFirstOrderDiscountPriceRule rejects negative discount values.
func NewFirstOrderDiscountPriceRule(discount decimal.Decimal) (*FirstOrderDiscountPriceRule, error) {
	if discount.LessThan(decimal.Zero) || discount.GreaterThan(decimal.NewFromInt(1)) {
		return nil, fmt.Errorf("invalid first order discount: %s, the value must be between 0 and 1", discount)
	}

	return &FirstOrderDiscountPriceRule{DiscountPercent: discount}, nil
}

// Apply discounts currentPrice when order.IsFirstOrder is true; otherwise it
// returns the price unchanged.
func (r *FirstOrderDiscountPriceRule) Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !order.IsFirstOrder() {
		return currentPrice, nil
	}

	return currentPrice.Sub(currentPrice.Mul(r.DiscountPercent)), nil
}

// CustomerDiscountPriceRule subtracts a flat per-tier discount from the price.
// Unlike FirstOrderDiscountPriceRule, the value is an absolute amount rather
// than a fraction.
type CustomerDiscountPriceRule struct {
	CustomerDiscounts map[CustomerType]decimal.Decimal
}

// NewCustomerDiscountPriceRule validates that every key is a known
// CustomerType and every discount is non-negative.
func NewCustomerDiscountPriceRule(customerDiscounts map[CustomerType]decimal.Decimal) (*CustomerDiscountPriceRule, error) {
	for customerType, discount := range customerDiscounts {
		if err := v.Validate(customerType, v.In(lo.ToAnySlice(CustomerTypes)...)); err != nil {
			return nil, fmt.Errorf("invalid customer type: %s", customerType)
		}

		if discount.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("discount cannot be negative for customer type %s: %s", customerType, discount)
		}
	}

	return &CustomerDiscountPriceRule{CustomerDiscounts: customerDiscounts}, nil
}

// Apply subtracts the tier-specific discount. It errors if the order's
// customer type has no entry in the discount map.
func (r *CustomerDiscountPriceRule) Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	discount, ok := r.CustomerDiscounts[order.CustomerType()]
	if !ok {
		return currentPrice, fmt.Errorf("customer type is invalid: %s", order.CustomerType())
	}

	return currentPrice.Sub(discount), nil
}

// RoundPriceRule rounds the running price up to a fixed decimal precision.
// Place this rule last in the chain so earlier discounts are not lost.
type RoundPriceRule struct {
	Precision int32
}

// NewRoundPriceRule clamps a negative precision to DefaultRoundPrecision.
func NewRoundPriceRule(precision int) (*RoundPriceRule, error) {
	if precision < 0 {
		precision = DefaultRoundPrecision
	}

	return &RoundPriceRule{Precision: int32(precision)}, nil
}

// Apply rounds up (ceiling) to r.Precision decimal places.
func (r *RoundPriceRule) Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	return currentPrice.RoundUp(r.Precision), nil
}
