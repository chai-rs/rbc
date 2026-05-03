// Command price_calculator is a small CLI that builds an order from flags,
// runs it through the configured pricing rule chain, and prints the final
// price. Pricing parameters (taxes, discounts, rounding) come from environment
// variables via envconfig; per-invocation order details come from CLI flags.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chai-rs/rbc/order"
	"github.com/kelseyhightower/envconfig"
	"github.com/shopspring/decimal"
	"github.com/urfave/cli/v3"
)

type Config struct {
	order.PriceRules
}

func main() {
	// Load pricing rules from the environment. MustProcess panics on bad input,
	// which is the desired behaviour for a misconfigured deploy.
	var cfg Config
	envconfig.MustProcess("", &cfg)

	// Build each rule. Order matters in NewCalculator below — taxes first,
	// discounts next, rounding last so it doesn't get partially undone.
	taxPriceRule := must(order.NewTaxPriceRule(cfg.Taxes, cfg.DefaultTaxRate))
	firstOrderDiscountPriceRule := must(order.NewFirstOrderDiscountPriceRule(cfg.FirstOrderDiscount))
	customerDiscountPriceRule := must(order.NewCustomerDiscountPriceRule(cfg.CustomerDiscounts))
	roundPriceRule := must(order.NewRoundPriceRule(cfg.RoundPrecision))

	calculator := order.NewCalculator(
		order.DefaultZeroPriceAction,
		taxPriceRule,
		firstOrderDiscountPriceRule,
		customerDiscountPriceRule,
		roundPriceRule,
	)

	order := must(createOrder(context.Background()))

	finalPrice := must(calculator.Calculate(order))

	fmt.Println("Final Price:", finalPrice)
}

// createOrder parses os.Args via urfave/cli and constructs an *order.Order.
// The actual construction happens inside cmd.Action because that is when v3
// has populated the flag values; the named return `built` is captured by the
// closure and surfaced once cmd.Run returns.
//
// If cmd.Run completes without invoking Action (e.g. the user passed -h on a
// build that exposes help), `built` stays nil and the process exits cleanly.
func createOrder(ctx context.Context) (built *order.Order, err error) {
	cmd := &cli.Command{
		Name:  "Price Calculator",
		Usage: "Calculate the final price of an order based on the defined rules",
		Flags: []cli.Flag{
			&cli.Float64Flag{
				Name:    "price",
				Aliases: []string{"p"},
				Value:   0,
				Usage:   "Base price of the order",
			},
			&cli.StringFlag{
				Name:    "country",
				Aliases: []string{"c"},
				Value:   "TH",
				Usage:   "Country code of the order (e.g., TH, FR)",
			},
			&cli.StringFlag{
				Name:    "customer-type",
				Aliases: []string{"t"},
				Value:   string(order.CustomerTypeRegular),
				Usage:   "Customer type of the order (e.g., regular, vip)",
			},
			&cli.BoolFlag{
				Name:    "first-order",
				Aliases: []string{"f"},
				Value:   false,
				Usage:   "Whether the order is the customer's first order",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			built, err = order.NewOrder(order.OrderParams{
				BasePrice:    decimal.NewFromFloat(c.Float64("price")),
				CountryCode:  c.String("country"),
				CustomerType: order.CustomerType(c.String("customer-type")),
				IsFirstOrder: c.Bool("first-order"),
			})

			return err
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		return nil, err
	}

	if built == nil {
		os.Exit(0)
	}

	return
}

// must unwraps a (T, error) pair: it returns T on success, otherwise it logs
// and exits with status 1. Used to keep main() linear for what is, at this
// stage, a single-shot CLI without recoverable failure modes.
func must[T any](v T, err error) T {
	if err != nil {
		log.Default().Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return v
}
