package main

import (
	"fmt"
	"math"
	"math/big"
)

// Number represents a number in scientific notation with high precision.
type Number interface {
	Add(Number) Number
	String() string
}

// DecimalAddition implements a simple decimal addition strategy with edge case handling.
type DecimalAddition struct {
	Mantissa float64
	Exponent int
}

// Add implements the Number interface with edge case handling.
func (a DecimalAddition) Add(b Number) Number {
	if b, ok := b.(DecimalAddition); ok {
		if a.Exponent != b.Exponent {
			panic("Exponents must be the same for decimal addition")
		}

		// Check for overflow
		if math.IsInf(a.Mantissa, 1) || math.IsInf(b.Mantissa, 1) {
			panic("Overflow occurred during addition")
		}

		resultMantissa := a.Mantissa + b.Mantissa

		// Check for precision loss
		if math.Abs(resultMantissa-a.Mantissa) > 0.0001 || math.Abs(resultMantissa-b.Mantissa) > 0.0001 {
			return BigDecimalAddition{Number: big.NewFloat(resultMantissa), Exponent: a.Exponent} // Use BigDecimal for higher precision
		}

		return DecimalAddition{Mantissa: resultMantissa, Exponent: a.Exponent}
	}
	panic("Invalid number type for addition")
}

// String implements the Number interface.
func (a DecimalAddition) String() string {
	if a.Exponent == 0 {
		return fmt.Sprintf("%.12f", a.Mantissa)
	}
	return fmt.Sprintf("%.12e", a.Mantissa*math.Pow10(a.Exponent))
}

// BigDecimalAddition implements addition using big.Float for high precision and edge case handling.
type BigDecimalAddition struct {
	Number *big.Float
	Exponent int
}

// Add implements the Number interface with edge case handling.
func (a BigDecimalAddition) Add(b Number) Number {
	if b, ok := b.(BigDecimalAddition); ok {
		if a.Exponent != b.Exponent {
			panic("Exponents must be the same for big decimal addition")
		}

		result := new(big.Float).Copy(a.Number)
		result.Add(result, b.Number)

		// Check for overflow
		if result.IsInf() {
			panic("Overflow occurred during addition")
		}

		return BigDecimalAddition{Number: result, Exponent: a.Exponent}
	}
	panic("Invalid number type for addition")
}

// String implements the Number interface.
func (a BigDecimalAddition) String() string {
	if a.Exponent == 0 {
		return fmt.Sprintf("%.12f", a.Number.Float64())
	}
	result := new(big.Float).Mul(a.Number, big.NewFloat(math.Pow10(a.Exponent)))
	return fmt.Sprintf("%.12e", result)
}

// Calculator encapsulates a number with a strategy for addition and edge case handling.
type Calculator struct {
	Number  Number
	Strategy Number
}

// Add adds two numbers using the current strategy with edge case handling.
func (c Calculator) Add(b Number) Number {
	result := c.Strategy.Add(c.Number)

	// Check for precision loss in BigDecimalAddition
	if bda, ok := result.(BigDecimalAddition); ok {
		if math.Abs(bda.Number.Float64()-c.Strategy.Add(b).(BigDecimalAddition).Number.Float64()) > 0.0001 {
			return BigDecimalAddition{Number: new(big.Float).SetPrec(20).Copy(bda.Number), Exponent: bda.Exponent} // Increase precision
		}