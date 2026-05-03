package order

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPriceRule_TaxPriceRule(t *testing.T) {
	var taxes = map[string]decimal.Decimal{
		"TH": decimal.NewFromFloat(0.07),
		"FR": decimal.NewFromFloat(0.20),
	}

	thOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	frOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "FR",
		CustomerType: CustomerTypeRegular,
	})

	usOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "US",
		CustomerType: CustomerTypeRegular,
	})

	defaultRate := decimal.NewFromFloat(0.10)
	noDefault := decimal.NewFromInt(-1)

	type input struct {
		order       *Order
		taxes       map[string]decimal.Decimal
		defaultRate decimal.Decimal
	}

	tests := []struct {
		name         string
		in           input
		currentPrice decimal.Decimal
		want         decimal.Decimal
		wantErr      bool
	}{
		{
			name:         "TH applies 7% tax",
			in:           input{order: thOrder, taxes: taxes},
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(107),
		},
		{
			name:         "FR applies 20% tax",
			in:           input{order: frOrder, taxes: taxes},
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(120),
		},
		{
			name:         "zero price stays zero",
			in:           input{order: thOrder, taxes: taxes, defaultRate: noDefault},
			currentPrice: decimal.Zero,
			want:         decimal.Zero,
		},
		{
			name:         "negative default disables fallback",
			in:           input{order: usOrder, taxes: taxes, defaultRate: noDefault},
			currentPrice: decimal.NewFromInt(100),
			wantErr:      true,
		},
		{
			name:         "empty tax map with negative default errors",
			in:           input{order: thOrder, taxes: map[string]decimal.Decimal{}, defaultRate: noDefault},
			currentPrice: decimal.NewFromInt(100),
			wantErr:      true,
		},
		{
			name:         "default rate fills in for unknown country",
			in:           input{order: usOrder, taxes: taxes, defaultRate: defaultRate},
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(110),
		},
		{
			name:         "default rate ignored when country present",
			in:           input{order: thOrder, taxes: taxes, defaultRate: defaultRate},
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(107),
		},
		{
			name:         "default rate with empty map",
			in:           input{order: usOrder, taxes: map[string]decimal.Decimal{}, defaultRate: defaultRate},
			currentPrice: decimal.NewFromInt(200),
			want:         decimal.NewFromInt(220),
		},
		{
			name:         "zero default rate applies 0% tax",
			in:           input{order: usOrder, taxes: taxes, defaultRate: decimal.Zero},
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(100),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewTaxPriceRule(tc.in.taxes, tc.in.defaultRate)
			require.NoError(t, err)

			got, err := rule.Apply(tc.in.order, tc.currentPrice)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, got.Equal(tc.want), "got %s, want %s", got, tc.want)
		})
	}
}

func TestPriceRule_NewTaxPriceRule(t *testing.T) {
	tests := []struct {
		name        string
		taxes       map[string]decimal.Decimal
		defaultRate decimal.Decimal
		wantErr     bool
	}{
		{
			name: "valid map",
			taxes: map[string]decimal.Decimal{
				"TH": decimal.NewFromFloat(0.07),
				"FR": decimal.NewFromFloat(0.20),
			},
			defaultRate: decimal.NewFromInt(-1),
		},
		{
			name:        "empty map",
			taxes:       map[string]decimal.Decimal{},
			defaultRate: decimal.NewFromInt(-1),
		},
		{
			name:        "rate at upper bound",
			taxes:       map[string]decimal.Decimal{"TH": decimal.NewFromInt(1)},
			defaultRate: decimal.NewFromInt(-1),
		},
		{
			name:        "invalid country code",
			taxes:       map[string]decimal.Decimal{"XX": decimal.NewFromFloat(0.07)},
			defaultRate: decimal.NewFromInt(-1),
			wantErr:     true,
		},
		{
			name:        "negative rate",
			taxes:       map[string]decimal.Decimal{"TH": decimal.NewFromFloat(-0.01)},
			defaultRate: decimal.NewFromInt(-1),
			wantErr:     true,
		},
		{
			name:        "rate above 1",
			taxes:       map[string]decimal.Decimal{"TH": decimal.NewFromFloat(1.01)},
			defaultRate: decimal.NewFromInt(-1),
			wantErr:     true,
		},
		{
			name:        "valid default rate",
			taxes:       map[string]decimal.Decimal{"TH": decimal.NewFromFloat(0.07)},
			defaultRate: decimal.NewFromFloat(0.05),
		},
		{
			name:        "negative default rate disables fallback",
			taxes:       map[string]decimal.Decimal{},
			defaultRate: decimal.NewFromFloat(-0.01),
		},
		{
			name:        "very negative default rate accepted as sentinel",
			taxes:       map[string]decimal.Decimal{},
			defaultRate: decimal.NewFromInt(-100),
		},
		{
			name:        "default rate at upper bound",
			taxes:       map[string]decimal.Decimal{},
			defaultRate: decimal.NewFromInt(1),
		},
		{
			name:        "default rate above 1",
			taxes:       map[string]decimal.Decimal{},
			defaultRate: decimal.NewFromFloat(1.01),
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewTaxPriceRule(tc.taxes, tc.defaultRate)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, rule)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rule)
		})
	}
}

func TestPriceRule_FirstOrderDiscountPriceRule(t *testing.T) {
	firstOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
		IsFirstOrder: true,
	})

	repeatOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	tests := []struct {
		name         string
		discount     decimal.Decimal
		order        *Order
		currentPrice decimal.Decimal
		want         decimal.Decimal
	}{
		{
			name:         "first order applies 10% discount",
			discount:     decimal.NewFromFloat(0.10),
			order:        firstOrder,
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(90),
		},
		{
			name:         "repeat order unchanged",
			discount:     decimal.NewFromFloat(0.10),
			order:        repeatOrder,
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(100),
		},
		{
			name:         "zero discount keeps price",
			discount:     decimal.Zero,
			order:        firstOrder,
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.NewFromInt(100),
		},
		{
			name:         "full discount zeros price",
			discount:     decimal.NewFromInt(1),
			order:        firstOrder,
			currentPrice: decimal.NewFromInt(100),
			want:         decimal.Zero,
		},
		{
			name:         "zero price stays zero",
			discount:     decimal.NewFromFloat(0.10),
			order:        firstOrder,
			currentPrice: decimal.Zero,
			want:         decimal.Zero,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewFirstOrderDiscountPriceRule(tc.discount)
			require.NoError(t, err)

			got, err := rule.Apply(tc.order, tc.currentPrice)
			require.NoError(t, err)
			require.True(t, got.Equal(tc.want), "got %s, want %s", got, tc.want)
		})
	}
}

func TestPriceRule_NewFirstOrderDiscountPriceRule(t *testing.T) {
	tests := []struct {
		name     string
		discount decimal.Decimal
		wantErr  bool
	}{
		{name: "zero", discount: decimal.Zero},
		{name: "fractional", discount: decimal.NewFromFloat(0.1)},
		{name: "one", discount: decimal.NewFromInt(1)},
		{name: "negative", discount: decimal.NewFromFloat(-0.01), wantErr: true},
		{name: "above one", discount: decimal.NewFromFloat(1.01), wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewFirstOrderDiscountPriceRule(tc.discount)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, rule)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rule)
		})
	}
}

func TestPriceRule_CustomerDiscountPriceRule(t *testing.T) {
	discounts := map[CustomerType]decimal.Decimal{
		CustomerTypeRegular: decimal.Zero,
		CustomerTypeVIP:     decimal.NewFromInt(100),
	}

	regularOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(500),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	vipOrder := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(500),
		CountryCode:  "TH",
		CustomerType: CustomerTypeVIP,
	})

	tests := []struct {
		name         string
		discounts    map[CustomerType]decimal.Decimal
		order        *Order
		currentPrice decimal.Decimal
		want         decimal.Decimal
		wantErr      bool
	}{
		{
			name:         "regular gets no discount",
			discounts:    discounts,
			order:        regularOrder,
			currentPrice: decimal.NewFromInt(500),
			want:         decimal.NewFromInt(500),
		},
		{
			name:         "vip gets flat 100 off",
			discounts:    discounts,
			order:        vipOrder,
			currentPrice: decimal.NewFromInt(500),
			want:         decimal.NewFromInt(400),
		},
		{
			name:         "missing customer type returns error",
			discounts:    map[CustomerType]decimal.Decimal{CustomerTypeRegular: decimal.Zero},
			order:        vipOrder,
			currentPrice: decimal.NewFromInt(500),
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewCustomerDiscountPriceRule(tc.discounts)
			require.NoError(t, err)

			got, err := rule.Apply(tc.order, tc.currentPrice)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, got.Equal(tc.want), "got %s, want %s", got, tc.want)
		})
	}
}

func TestPriceRule_NewCustomerDiscountPriceRule(t *testing.T) {
	tests := []struct {
		name      string
		discounts map[CustomerType]decimal.Decimal
		wantErr   bool
	}{
		{
			name: "valid map",
			discounts: map[CustomerType]decimal.Decimal{
				CustomerTypeRegular: decimal.Zero,
				CustomerTypeVIP:     decimal.NewFromInt(100),
			},
		},
		{
			name:      "empty map",
			discounts: map[CustomerType]decimal.Decimal{},
		},
		{
			name:      "unknown customer type",
			discounts: map[CustomerType]decimal.Decimal{CustomerType("guest"): decimal.NewFromInt(10)},
			wantErr:   true,
		},
		{
			name:      "negative discount",
			discounts: map[CustomerType]decimal.Decimal{CustomerTypeVIP: decimal.NewFromInt(-1)},
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewCustomerDiscountPriceRule(tc.discounts)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, rule)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, rule)
		})
	}
}

func TestPriceRule_RoundPriceRule(t *testing.T) {
	order := mustOrder(t, OrderParams{
		BasePrice:    decimal.NewFromInt(100),
		CountryCode:  "TH",
		CustomerType: CustomerTypeRegular,
	})

	tests := []struct {
		name         string
		precision    int
		currentPrice decimal.Decimal
		want         decimal.Decimal
	}{
		{
			name:         "rounds up to 2 decimals",
			precision:    2,
			currentPrice: decimal.RequireFromString("1.231"),
			want:         decimal.RequireFromString("1.24"),
		},
		{
			name:         "exact value preserved",
			precision:    2,
			currentPrice: decimal.RequireFromString("1.20"),
			want:         decimal.RequireFromString("1.20"),
		},
		{
			name:         "rounds up to integer",
			precision:    0,
			currentPrice: decimal.RequireFromString("1.01"),
			want:         decimal.NewFromInt(2),
		},
		{
			name:         "negative precision falls back to default",
			precision:    -1,
			currentPrice: decimal.RequireFromString("1.231"),
			want:         decimal.RequireFromString("1.24"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewRoundPriceRule(tc.precision)
			require.NoError(t, err)

			got, err := rule.Apply(order, tc.currentPrice)
			require.NoError(t, err)
			require.True(t, got.Equal(tc.want), "got %s, want %s", got, tc.want)
		})
	}
}

func TestPriceRule_NewRoundPriceRule(t *testing.T) {
	tests := []struct {
		name      string
		precision int
		want      int32
	}{
		{name: "zero precision", precision: 0, want: 0},
		{name: "positive precision", precision: 4, want: 4},
		{name: "negative clamps to default", precision: -1, want: int32(DefaultRoundPrecision)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule, err := NewRoundPriceRule(tc.precision)
			require.NoError(t, err)
			require.Equal(t, tc.want, rule.Precision)
		})
	}
}

func mustOrder(t *testing.T, params OrderParams) *Order {
	t.Helper()

	order, err := NewOrder(params)
	if err != nil {
		t.Fatalf("NewOrder: %v", err)
	}

	return order
}
