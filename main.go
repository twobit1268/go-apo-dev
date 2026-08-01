package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")

	// variables: declared with var, or shorthand :=
	var age = 30
	name := "Gopher"
	fmt.Println(name, "is", age, "years old")

	// basic types
	var pi = 3.14159
	isLearning := true
	fmt.Println(pi, isLearning)
}
