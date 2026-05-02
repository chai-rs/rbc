// Package order models orders and the chain of pricing rules used to compute
// their final price (taxes, discounts, rounding).
package order

// CustomerType classifies the buyer for the purpose of selecting a discount tier.
type CustomerType string

const (
	// CustomerTypeRegular is the default tier with no special discount.
	CustomerTypeRegular CustomerType = "regular"
	// CustomerTypeVIP is the preferred tier and typically receives a larger discount.
	CustomerTypeVIP CustomerType = "vip"
)

// CustomerTypes enumerates every valid CustomerType. Used by validators
// (e.g. ozzo-validation's v.In) to whitelist incoming values.
var CustomerTypes = []CustomerType{CustomerTypeRegular, CustomerTypeVIP}
