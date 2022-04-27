package main

import (
	// "fmt"

	"fmt"

	"github.com/NOX73/go-neural"
	"github.com/NOX73/go-neural/learn"
	"github.com/NOX73/go-neural/persist"
)

//...

func Model(input, idealOutput []float64) {
	// Learning speed [0..1]
	var speed float64
  
	input = []float64{0,1,0,1,1,1,0,1,0}
	idealOutput = []float64{1}
	speed = 0.1
	// Network has 9 enters and 3 layers 
	// ( 9 neurons, 9 neurons and 4 neurons).
	// Last layer is network output.
	n := neural.NewNetwork(9, []int{9,9,1})
	// Randomize sypaseses weights
	n.RandomizeSynapses()
	// for i:=0;i<1000;i++ {
	learn.Learn(n, input, idealOutput, speed)
	// }
	// fmt.Println(learn)
	result := n.Calculate([]float64{0,1,0,1,1,1,0,1,0})
	fmt.Println(result)
	persist.ToFile("/model.json", n)
}