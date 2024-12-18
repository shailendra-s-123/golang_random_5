package main  
import (  
    "fmt"
    "math/big"
)

// Adapter interface for scientific notation calculations
type ScientificCalculator interface {
    Add(a, b float64) float64
    Subtract(a, b float64) float64
    Multiply(a, b float64) float64
    Divide(a, b float64) float64
    Power(base, exponent float64) float64
}

// Implementation of the Adapter interface using big.Float
type BigFloatCalculator struct{}

func (c *BigFloatCalculator) Add(a, b float64) float64 {
    result := new(big.Float).Add(big.NewFloat(a), big.NewFloat(b))
    return result.Float64()
}

func (c *BigFloatCalculator) Subtract(a, b float64) float64 {
    result := new(big.Float).Sub(big.NewFloat(a), big.NewFloat(b))
    return result.Float64()
}

func (c *BigFloatCalculator) Multiply(a, b float64) float64 {
    result := new(big.Float).Mul(big.NewFloat(a), big.NewFloat(b))
    return result.Float64()
}

func (c *BigFloatCalculator) Divide(a, b float64) float64 {
    result := new(big.Float).Quo(big.NewFloat(a), big.NewFloat(b))
    return result.Float64()
}

func (c *BigFloatCalculator) Power(base, exponent float64) float64 {
    result := new(big.Float).Pow(big.NewFloat(base), big.NewFloat(exponent))
    return result.Float64()
}


func main() {  
    // Create an instance of the BigFloatCalculator as an Adapter
    calc := &BigFloatCalculator{}

    // Use the Adapter interface for calculations
    fmt.Println("Adding 1.2345678901234567e+30 and 2.345678901234567e+30:")
    sum := calc.Add(1.2345678901234567e+30, 2.345678901234567e+30)
    fmt.Println(sum)

    fmt.Println("\nSubtracting 1.2345678901234567e+30 from 2.345678901234567e+30:")
    difference := calc.Subtract(2.345678901234567e+30, 1.2345678901234567e+30)
    fmt.Println(difference)

    // The rest of the calculations using the Adapter interface can be done here
}  