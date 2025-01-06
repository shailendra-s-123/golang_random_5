package main

import (
	"math/big"
	"fmt"
)

const (
	maxExponent int = 1000
	minExponent int = -1000
)

// Number represents a number in scientific notation with high precision.
type Number interface {
	Add(Number) Number
	String() string
}

// BigDecimalAddition implements addition using big.Float for high precision.
type BigDecimalAddition struct {
	Number *big.Float
	Eponent int
}

// Add implements the Number interface.
func (a BigDecimalAddition) Add(b Number) Number {
	if b, ok := b.(BigDecimalAddition); ok {
		if a.Eponent != b.Eponent {
			return handleExponentMismatch(a, b)
		}
		a.Number.SetPrec(20) // Define precision explicitly
		b.Number.SetPrec(20)
		sum, ok := a.Number.Add(a.Number, b.Number)
		if !ok {
			return handleOverflow(a, b)
		}
		return BigDecimalAddition{Number: sum, Eponent: a.Eponent}
	}
	return handleInvalidNumber(a, b)
}

// String implements the Number interface.
func (a BigDecimalAddition) String() string {
	result := new(big.Float).Mul(a.Number, big.NewFloat(math.Pow10(float64(a.Eponent))))
	return fmt.Sprintf("%.12e", result)
}

func handleExponentMismatch(a, b BigDecimalAddition) Number {
	panic("Exponents must be the same for addition")
}

func handleOverflow(a, b BigDecimalAddition) Number {
	panic("Overflow occurred during addition")
}

func handleInvalidNumber(a, b Number) Number {
	panic("Invalid number type for addition")
}

func main() {
	bigDecimal1 := BigDecimalAddition{Number: big.NewFloat(1.234), Eponent: 2}
	bigDecimal2 := BigDecimalAddition{Number: big.NewFloat(5.678), Eponent: 2}

	fmt.Println("Big Decimal Addition:", bigDecimal1.Add(bigDecimal2)) // 6.912e+00

	// Test with large exponents
	bigDecimal3 := BigDecimalAddition{Number: big.NewFloat(1.234), Eponent: maxExponent - 1}
	bigDecimal4 := BigDecimalAddition{Number: big.NewFloat(5.678), Eponent: maxExponent - 1}
	fmt.Println("Large Exponent Addition:", bigDecimal3.Add(bigDecimal4))
}