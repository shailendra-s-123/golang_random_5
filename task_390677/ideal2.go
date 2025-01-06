package main

import (
	"fmt"
	"math"
	"math/big"
)

const (
	maxExponent int = 1000
	minExponent int = -1000 
)

type Number interface {
	Add(Number) Number
	String() string
}

// BigDecimalAddition handles high-precision addition of scientific numbers.
type BigDecimalAddition struct {
	Number  *big.Float
	Exponent int
}

// Add adds another BigDecimalAddition to the current one, ensuring proper exponent handling.
func (a BigDecimalAddition) Add(b Number) Number {
	// Check if the types match
	if b, ok := b.(BigDecimalAddition); ok {
		// Ensure exponents match for addition
		if a.Exponent != b.Exponent {
			return handleExponentMismatch(a, b)
		}
		
		// Set precision explicitly
		a.Number.SetPrec(128) // Set high precision to avoid loss
		b.Number.SetPrec(128)
		
		// Perform addition
		result := new(big.Float).Add(a.Number, b.Number)
		
		// If an overflow happens during addition, handle it
		if result == nil {
			return handleOverflow(a, b)
		}
		
		// Return the result as a new BigDecimalAddition
		return BigDecimalAddition{Number: result, Exponent: a.Exponent}
	}
	return handleInvalidNumber(a, b)
}

// String returns the string representation of the number in scientific notation.
func (a BigDecimalAddition) String() string {
	// Handle precision and scale
	result := new(big.Float).Mul(a.Number, big.NewFloat(math.Pow10(int(a.Exponent)))) // Fix the type conversion here
	return fmt.Sprintf("%.12e", result)
}

// handleExponentMismatch throws an error if the exponents of the numbers don't match.
func handleExponentMismatch(a, b BigDecimalAddition) Number {
	panic("Exponents must be the same for addition")
}

// handleOverflow throws an error if an overflow occurs during the addition.
func handleOverflow(a, b BigDecimalAddition) Number {
	panic("Overflow occurred during addition")
}

// handleInvalidNumber throws an error if the number types do not match.
func handleInvalidNumber(a, b Number) Number {
	panic("Invalid number type for addition")
}

// main function to test the enhanced code
func main() {
	// Test cases with valid inputs
	bigDecimal1 := BigDecimalAddition{Number: big.NewFloat(1.234), Exponent: 2}
	bigDecimal2 := BigDecimalAddition{Number: big.NewFloat(5.678), Exponent: 2}
	fmt.Println("Big Decimal Addition:", bigDecimal1.Add(bigDecimal2)) // Expected: 6.912e+00

	// Test with large exponents
	bigDecimal3 := BigDecimalAddition{Number: big.NewFloat(1.234), Exponent: maxExponent - 1}
	bigDecimal4 := BigDecimalAddition{Number: big.NewFloat(5.678), Exponent: maxExponent - 1}
	fmt.Println("Large Exponent Addition:", bigDecimal3.Add(bigDecimal4)) // Expected output with large exponents

	// Test with edge cases: Exponent mismatch
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()
	bigDecimal5 := BigDecimalAddition{Number: big.NewFloat(3.456), Exponent: 2}
	bigDecimal6 := BigDecimalAddition{Number: big.NewFloat(1.234), Exponent: 3}
	fmt.Println("Edge Case Addition:", bigDecimal5.Add(bigDecimal6)) // Should trigger panic
}



