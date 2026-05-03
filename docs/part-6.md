# Part 6: Scaling the Design

You are asked to support 10 more countries next month.

Do not design a full rule engine. Explain the simplest approach you would start with.

## Assignments

### 1. Would you keep the current design?

Yes, the current design is modular and makes it easy to add new country codes and their corresponding tax rates through configuration, what we have to do is to update the `TAXES` environment variable with the new country codes and their tax rates and then reload the service configuration again.

### 2. What would you change?

In production, I would move this configuration to a database or a dedicated configuration service so it can scale better and be managed more easily as the number of supported countries grows?

### 3. Would you use configuration, database, or code? Why?

I would use a database or a dedicated configuration service because it allows for easier management, scalability, and dynamic updates without requiring code changes or redeployments?

### 4. How would you handle a rule like: Applies only in TH, for VIP customers, and only if order > 5000?

I would implement the new rule as a separate `PriceRule` that checks for it and applies the diescount, for the example code it would look like this:

```go
// order/price_rule.go
var _ PriceRule = (*SpecificDiscountPriceRule)(nil)

type DiscountPriceRule struct{
	countryCode string // TH
	minBasePrice decimal.Decimal // 5,000
	customerType CustomerType // VIP
	discount decimal.Decimal // 100
}

func (r *DiscountPriceRule) Apply(order *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
  if order.Country() == r.countryCode && 
  order.CustomerType() == r.customerType && 
  price.GreaterThan(r.minBasePrice) {
    return price.Sub(r.discount), nil
  }
  
  return price, nil
}

```
