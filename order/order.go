package order

import (
	"fmt"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

// Order is an immutable, validated purchase request that flows through the
// pricing rule chain. Fields are unexported so a value can only be obtained via
// NewOrder, which guarantees the invariants enforced by Validate.
type Order struct {
	basePrice    decimal.Decimal
	countryCode  string
	customerType CustomerType
	isFirstOrder bool
}

// OrderParams is the plain-data input accepted by NewOrder. It mirrors Order
// with exported fields so callers (CLI, HTTP, tests) can populate it directly.
type OrderParams struct {
	BasePrice    decimal.Decimal
	CountryCode  string
	CustomerType CustomerType
	IsFirstOrder bool
}

// NewOrder builds an Order from params and runs Validate. It returns an error
// if any field is missing or out of range; the returned *Order is nil on error.
func NewOrder(params OrderParams) (*Order, error) {
	order := &Order{
		basePrice:    params.BasePrice,
		countryCode:  params.CountryCode,
		customerType: params.CustomerType,
		isFirstOrder: params.IsFirstOrder,
	}

	if err := order.Validate(); err != nil {
		return nil, err
	}

	return order, nil
}

// Validate checks that basePrice is non-negative, countryCode is an ISO-3166-1
// alpha-2 code, and customerType is one of CustomerTypes.
func (o *Order) Validate() error {
	if err := v.Validate(o.basePrice.InexactFloat64(), v.Required, v.Min(0.0)); err != nil {
		return fmt.Errorf("basePrice: %w", err)
	}

	return v.ValidateStruct(o,
		v.Field(&o.countryCode, v.Required, is.CountryCode2),
		v.Field(&o.customerType, v.Required, v.In(lo.ToAnySlice(CustomerTypes)...)),
	)
}

// BasePrice returns the pre-rule price the calculator starts from.
func (o *Order) BasePrice() decimal.Decimal {
	return o.basePrice
}

// CountryCode returns the ISO-3166-1 alpha-2 destination country.
func (o *Order) CountryCode() string {
	return o.countryCode
}

// CustomerType returns the tier used to look up customer-tier discounts.
func (o *Order) CustomerType() CustomerType {
	return o.customerType
}

// IsFirstOrder reports whether the first-order discount rule should apply.
func (o *Order) IsFirstOrder() bool {
	return o.isFirstOrder
}
