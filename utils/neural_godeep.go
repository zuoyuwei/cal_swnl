package main

import (
	"github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"
)

func main() {
	var data = training.Examples{
		{[]float64{2.7810836, 2.550537003}, []float64{0}},
		{[]float64{1.465489372, 2.362125076}, []float64{0}},
		{[]float64{3.396561688, 4.400293529}, []float64{0}},
		{[]float64{1.38807019, 1.850220317}, []float64{0}},
		{[]float64{7.627531214, 2.759262235}, []float64{1}},
		{[]float64{5.332441248, 2.088626775}, []float64{1}},
		{[]float64{6.922596716, 1.77106367}, []float64{1}},
		{[]float64{8.675418651, -0.242068655}, []float64{1}},
	}

	n := deep.NewNeural(&deep.Config{
		/* Input dimensionality */
		Inputs: 2,
		/* Two hidden layers consisting of two neurons each, and a single output */
		Layout: []int{2, 2, 1},
		/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
		Activation: deep.ActivationSigmoid,
		/* Determines output layer activation & loss function:
		   ModeRegression: linear outputs with MSE loss
		   ModeMultiClass: softmax output with Cross Entropy loss
		   ModeMultiLabel: sigmoid output with Cross Entropy loss
		   ModeBinary: sigmoid output with binary CE loss */
		Mode: deep.ModeBinary,
		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
		Weight: deep.NewNormal(1.0, 0.0),
		/* Apply bias */
		Bias: true,
	})

	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := training.NewSGD(0.05, 0.1, 1e-6, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := training.NewTrainer(optimizer, 50)

	training, heldout := data.Split(0.5)
	trainer.Train(n, training, heldout, 1000) // training, validation, iterations
}
