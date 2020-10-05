package main

import (
	"fmt"
	"math"

	"golang.org/x/net/context"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Quote represents a currency value.
type Quote struct {
	Dollars uint32
	Cents   uint32
}

// String representation of the Quote.
func (q Quote) String() string {
	return fmt.Sprintf("$%d.%d", q.Dollars, q.Cents)
}

// CreateQuoteFromCount takes a number of items and returns a Price struct.
func CreateQuoteFromCount(ctx context.Context, count int) Quote {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateQuoteFromCount")
	defer span.Finish()
	return CreateQuoteFromFloat(ctx, quoteByCountFloat(ctx, count))
}

// CreateQuoteFromFloat takes a price represented as a float and creates a Price struct.
func CreateQuoteFromFloat(ctx context.Context, value float64) Quote {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateQuoteFromFloat")
	defer span.Finish()
	units, fraction := math.Modf(value)
	return Quote{
		uint32(units),
		uint32(math.Trunc(fraction * 100)),
	}
}

// quoteByCountFloat takes a number of items and generates a price quote represented as a float.
func quoteByCountFloat(ctx context.Context, count int) float64 {
	span, ctx := tracer.StartSpanFromContext(ctx, "quoteByCountFloat")
	defer span.Finish()
	if count == 0 {
		return 0
	}
	count64 := float64(count)
	var p = 1 + (count64 * 0.2)
	return count64 + math.Pow(3, p)
}
