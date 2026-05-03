package order

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var noDefault = decimal.NewFromInt(-1)

func TestCalculator_FullChain_RegularRepeatOrder(t *testing.T) {
	calc := newDefaultCalculator(t)

	ord := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	// 100 → *1.07 = 107 → first-order no-op → regular discount 0 → round = 107
	got, err := calc.Calculate(ord)
	require.NoError(t, err)
	require.True(t, got.Equal(decimal.NewFromInt(107)), "got %s", got)
}

func TestCalculator_FullChain_VIPFirstOrderFractional(t *testing.T) {
	calc := newDefaultCalculator(t)

	// 199.99 → *1.07 = 213.9893 → *0.9 = 192.59037 → -100 = 92.59037 → roundUp(2) = 92.60
	ord := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromFloat(199.99),
		CountryCode:  "TH",
		CustomerType: CustomerTypeVIP,
		IsFirstOrder: true,
	})

	got, err := calc.Calculate(ord)
	require.NoError(t, err)
	require.True(t, got.Equal(decimal.RequireFromString("92.60")), "got %s", got)
}

func TestCalculator_RuleOrderMatters(t *testing.T) {
	tax := mustTax(t, map[string]decimal.Decimal{"TH": decimal.NewFromFloat(0.07)}, noDefault)
	round := mustRound(t, 0)

	// round-first: ceil(100.4)=101 → *1.07 = 108.07
	// tax-first:   100.4*1.07 = 107.428 → ceil = 108
	ord := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromFloat(100.4),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	roundFirst, err := NewCalculator(DefaultZeroPriceAction, round, tax).Calculate(ord)
	require.NoError(t, err)
	taxFirst, err := NewCalculator(DefaultZeroPriceAction, tax, round).Calculate(ord)
	require.NoError(t, err)

	require.True(t, roundFirst.Equal(decimal.RequireFromString("108.07")), "round-first %s", roundFirst)
	require.True(t, taxFirst.Equal(decimal.NewFromInt(108)), "tax-first %s", taxFirst)
	require.False(t, roundFirst.Equal(taxFirst), "ordering should change result")
}

func TestCalculator_RuleErrorShortCircuits(t *testing.T) {
	tax := mustTax(t, map[string]decimal.Decimal{"TH": decimal.NewFromFloat(0.07)}, noDefault)
	probe := &probeRule{}

	ord := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "US",
		CustomerType: CustomerTypeRegular,
	})

	calc := NewCalculator(DefaultZeroPriceAction, tax, probe)
	got, err := calc.Calculate(ord)
	require.Error(t, err)
	require.True(t, got.Equal(decimal.Zero))
	require.False(t, probe.called, "rule after a failing rule must not run")
}

func TestCalculator_DefaultZeroPriceActionAborts(t *testing.T) {
	calc := newDefaultCalculator(t)

	// Small base + VIP + first-order forces customer discount to drive price
	// negative: 10 → 10.7 → 9.63 → -100 = -90.37.
	ord := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(10),
		CountryCode:  "TH",
		CustomerType: CustomerTypeVIP,
		IsFirstOrder: true,
	})

	got, err := calc.Calculate(ord)
	require.Error(t, err)
	require.True(t, got.Equal(decimal.Zero))
}

// --- helpers ---

type probeRule struct{ called bool }

func (p *probeRule) Apply(_ *Order, currentPrice decimal.Decimal) (decimal.Decimal, error) {
	p.called = true
	return currentPrice, nil
}

func newDefaultCalculator(t *testing.T) *Calculator {
	t.Helper()
	return NewCalculator(
		DefaultZeroPriceAction,
		mustTax(t, map[string]decimal.Decimal{
			"TH": decimal.NewFromFloat(0.07),
			"FR": decimal.NewFromFloat(0.20),
		}, noDefault),
		mustFirstOrder(t, decimal.NewFromFloat(0.10)),
		mustCustomer(t, map[CustomerType]decimal.Decimal{
			CustomerTypeRegular: decimal.Zero,
			CustomerTypeVIP:     decimal.NewFromInt(100),
		}),
		mustRound(t, 2),
	)
}

func mustTax(t *testing.T, taxes map[string]decimal.Decimal, defaultRate decimal.Decimal) *TaxPriceRule {
	t.Helper()
	r, err := NewTaxPriceRule(taxes, defaultRate)
	require.NoError(t, err)
	return r
}

func mustFirstOrder(t *testing.T, discount decimal.Decimal) *FirstOrderDiscountPriceRule {
	t.Helper()
	r, err := NewFirstOrderDiscountPriceRule(discount)
	require.NoError(t, err)
	return r
}

func mustCustomer(t *testing.T, discounts map[CustomerType]decimal.Decimal) *CustomerDiscountPriceRule {
	t.Helper()
	r, err := NewCustomerDiscountPriceRule(discounts)
	require.NoError(t, err)
	return r
}

func mustRound(t *testing.T, precision int) *RoundPriceRule {
	t.Helper()
	r, err := NewRoundPriceRule(precision)
	require.NoError(t, err)
	return r
}
