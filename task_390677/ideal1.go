package main

import (
	"math/big"
	"fmt"
	"math"
)

// Number represents a number in scientific notation with high precision.
type Number interface {
	Add(Number) Number
	String() string
}

// DecimalAddition implements a simple decimal addition strategy.
type DecimalAddition struct {
	Number float64
	Eponent int
}

// Add implements the Number interface.
func (a DecimalAddition) Add(b Number) Number {
	if b, ok := b.(DecimalAddition); ok {
		if a.Eponent != b.Eponent {
			panic("Exponents must be the same for decimal addition")
		}
		return DecimalAddition{Number: a.Number + b.Number, Eponent: a.Eponent}
	}
	panic("Invalid number type for addition")
}

// String implements the Number interface.
func (a DecimalAddition) String() string {
	return fmt.Sprintf("%.12e", a.Number*math.Pow10(a.Eponent)) // Use a.Eponent directly as int
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
			panic("Exponents must be the same for big decimal addition")
		}
		a.Number.SetPrec(20)
		b.Number.SetPrec(20)
		a.Number.Add(a.Number, b.Number)
		return BigDecimalAddition{Number: a.Number, Eponent: a.Eponent}
	}
	panic("Invalid number type for addition")
}

// String implements the Number interface.
func (a BigDecimalAddition) String() string {
	// Corrected to use fmt.Sprintf for formatting
	result := new(big.Float).Mul(a.Number, big.NewFloat(math.Pow10(a.Eponent))) // Use int for exponent
	return fmt.Sprintf("%.12e", result) // Using fmt.Sprintf to format the result properly
}

// Calculator encapsulates a number with a strategy for addition.
type Calculator struct {
	Number  Number
	Strategy Number
}

// Add adds two numbers using the current strategy.
func (c Calculator) Add(b Number) Number {
	return c.Strategy.Add(c.Number)
}

func main() {
	decimal := DecimalAddition{Number: 1.234, Eponent: 2}
	bigDecimal := BigDecimalAddition{Number: big.NewFloat(1.234), Eponent: 2}

	calculator := Calculator{Number: decimal, Strategy: decimal}
	fmt.Println("Decimal Addition:", calculator.Add(DecimalAddition{Number: 5.678, Eponent: 2})) // 6.912e+00

	calculator = Calculator{Number: bigDecimal, Strategy: bigDecimal}
	fmt.Println("Big Decimal Addition:", calculator.Add(BigDecimalAddition{Number: big.NewFloat(5.678), Eponent: 2})) // 6.912e+00
}

