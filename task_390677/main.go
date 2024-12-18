package main

import (
	"fmt"
	"math/big"
)

// ScientificNumber represents a number in scientific notation with high precision.
type ScientificNumber struct {
	Base      *big.Float
	Exponent  int
}

// NewScientificNumber creates a new ScientificNumber given a base and an exponent.
func NewScientificNumber(base float64, exponent int) *ScientificNumber {
	baseBig := big.NewFloat(base)
	return &ScientificNumber{
		Base:     baseBig,
		Exponent: exponent,
	}
}

// Add adds two ScientificNumbers together and returns a new ScientificNumber.
func (n *ScientificNumber) Add(other *ScientificNumber) *ScientificNumber {
	// Adjust exponents to match
	if n.Exponent > other.Exponent {
		scale := int64(n.Exponent - other.Exponent)
		other.Base = new(big.Float).Mul(other.Base, big.NewFloat(float64(scale)))
		other.Exponent = n.Exponent
	} else if other.Exponent > n.Exponent {
		scale := int64(other.Exponent - n.Exponent)
		n.Base = new(big.Float).Mul(n.Base, big.NewFloat(float64(scale)))
		n.Exponent = other.Exponent
	}
	// Perform the addition
	result := new(big.Float).Add(n.Base, other.Base)
	return &ScientificNumber{
		Base:     result,
		Exponent: n.Exponent,
	}
}

// Multiply multiplies two ScientificNumbers together and returns a new ScientificNumber.
func (n *ScientificNumber) Multiply(other *ScientificNumber) *ScientificNumber {
	// Multiply the bases and add the exponents
	resultBase := new(big.Float).Mul(n.Base, other.Base)
	resultExponent := n.Exponent + other.Exponent
	return &ScientificNumber{
		Base:     resultBase,
		Exponent: resultExponent,
	}
}

// String converts the ScientificNumber into a string representation in scientific notation.
func (n *ScientificNumber) String() string {
	return fmt.Sprintf("%.12fe+%d", n.Base, n.Exponent)
}

// Calculator provides the structure for performing operations on ScientificNumbers.
type Calculator struct {
	Number  *ScientificNumber
	Strategy func(*ScientificNumber, *ScientificNumber) *ScientificNumber
}

// PerformOperation performs the given operation using the strategy.
func (c *Calculator) PerformOperation(other *ScientificNumber) *ScientificNumber {
	return c.Strategy(c.Number, other)
}

func main() {
	// Create two numbers in scientific notation
	num1 := NewScientificNumber(1.234, 2)  // 1.234e+2
	num2 := NewScientificNumber(3.456, 2)  // 3.456e+2

	// Perform addition
	calculator := Calculator{Number: num1, Strategy: (*ScientificNumber).Add}
	addResult := calculator.PerformOperation(num2)
	fmt.Println("Addition Result:", addResult.String())  // Expected output: 4.690000000000e+2

	// Perform multiplication
	calculator.Strategy = (*ScientificNumber).Multiply
	multiplyResult := calculator.PerformOperation(num2)
	fmt.Println("Multiplication Result:", multiplyResult.String()) // Expected output: 4.259456000000e+4
}
