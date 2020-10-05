package main

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// seeded determines if the random number generator is ready.
var seeded bool = false

// CreateTrackingID generates a tracking ID.
func CreateTrackingID(ctx context.Context, salt string) string {
	span, ctx := tracer.StartSpanFromContext(ctx, "CreateTrackingID")
	defer span.Finish()
	if !seeded {
		rand.Seed(time.Now().UnixNano())
		seeded = true
	}

	return fmt.Sprintf("%c%c-%d%s-%d%s",
		getRandomLetterCode(ctx),
		getRandomLetterCode(ctx),
		len(salt),
		getRandomNumber(ctx, 3),
		len(salt)/2,
		getRandomNumber(ctx, 7),
	)
}

// getRandomLetterCode generates a code point value for a capital letter.
func getRandomLetterCode(ctx context.Context) uint32 {
	span, ctx := tracer.StartSpanFromContext(ctx, "getRandomLetterCode")
	defer span.Finish()
	return 65 + uint32(rand.Intn(25))
}

// getRandomNumber generates a string representation of a number with the requested number of digits.
func getRandomNumber(ctx context.Context, digits int) string {
	span, ctx := tracer.StartSpanFromContext(ctx, "getRandomNumber")
	defer span.Finish()
	str := ""
	for i := 0; i < digits; i++ {
		str = fmt.Sprintf("%s%d", str, rand.Intn(10))
	}

	return str
}
