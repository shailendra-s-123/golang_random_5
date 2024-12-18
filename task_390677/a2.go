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
	return &ScientificNumber{
		Base:     big.NewFloat(base),
		Exponent: exponent,
	}
}

// String provides a readable string representation of the ScientificNumber in scientific notation.
func (n *ScientificNumber) String() string {
	return fmt.Sprintf("%.12fe%+d", n.Base, n.Exponent)
}

// Strategy defines the interface for calculation operations on ScientificNumbers.
type Strategy interface {
	Apply(a, b *ScientificNumber) (*ScientificNumber, error)
}

// AddStrategy implements the Strategy interface for addition.
type AddStrategy struct{}

// Apply adds two ScientificNumbers.
func (s *AddStrategy) Apply(a, b *ScientificNumber) (*ScientificNumber, error) {
	// Adjust exponents to match
	if a.Exponent != b.Exponent {
		return nil, fmt.Errorf("cannot add numbers with different exponents: %s and %s", a, b)
	}
	// Perform the addition
	resultBase := new(big.Float).Add(a.Base, b.Base)
	return &ScientificNumber{
		Base:     resultBase,
		Exponent: a.Exponent,
	}, nil
}

// MultiplyStrategy implements the Strategy interface for multiplication.
type MultiplyStrategy struct{}

// Apply multiplies two ScientificNumbers.
func (s *MultiplyStrategy) Apply(a, b *ScientificNumber) (*ScientificNumber, error) {
	// Multiply the bases and add the exponents
	resultBase := new(big.Float).Mul(a.Base, b.Base)
	resultExponent := a.Exponent + b.Exponent
	return &ScientificNumber{
		Base:     resultBase,
		Exponent: resultExponent,
	}, nil
}

// Calculator allows executing operations on ScientificNumbers using strategies.
type Calculator struct {
	Strategy Strategy
}

// PerformOperation performs the defined operation on two ScientificNumbers.
func (c *Calculator) PerformOperation(a, b *ScientificNumber) (*ScientificNumber, error) {
	return c.Strategy.Apply(a, b)
}

func main() {
	// Create two numbers in scientific notation
	num1 := NewScientificNumber(1.234, 2)  // 1.234e+2
	num2 := NewScientificNumber(3.456, 2)  // 3.456e+2

	// Addition
	addCalculator := Calculator{Strategy: &AddStrategy{}}
	addResult, err := addCalculator.PerformOperation(num1, num2)
	if err != nil {
		fmt.Println("Addition Error:", err)
	} else {
		fmt.Println("Addition Result:", addResult.String())  // Expected output: 4.690000000000e+2
	}

	// Multiplication
	multiplyCalculator := Calculator{Strategy: &MultiplyStrategy{}}
	multiplyResult, err := multiplyCalculator.PerformOperation(num1, num2)
	if err != nil {
		fmt.Println("Multiplication Error:", err)
	} else {
		fmt.Println("Multiplication Result:", multiplyResult.String()) // Expected output: 4.259456000000e+4
	}

	// Addition with different exponents should result in an error
	num3 := NewScientificNumber(1.234, 3)  // 1.234e+3
	addCalculatorWithError := Calculator{Strategy: &AddStrategy{}}
	_, errWithError := addCalculatorWithError.PerformOperation(num1, num3)
	fmt.Println("Error in Addition:", errWithError) // Expected error: cannot add numbers with different exponents
}