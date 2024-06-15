package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/mohae/deepcopy"
)

/*
successfully done testing - Weights Mutation: Modifies the weights of connections to adapt and refine the network's responses to inputs.
successfully done testing - Bias Mutation: Adjusts the biases of neurons to fine-tune the activation potential, enhancing the network's ability to fit complex patterns.
Add Node Mutation: Inserts a new node by splitting an existing connection, increasing the network's depth and potential for complexity.
Add Connection Mutation: Creates a new connection between previously unconnected nodes, expanding the network's capacity for diverse interactions.
Connection Enable/Disable: Toggles the enabled state of connections, allowing the network to experiment with different neural pathways without permanent structural changes. (Will be developed later)
Add Layer Mutation: Introduces entirely new layers to the network, significantly enhancing its depth and functional complexity.
Activation Function Mutation: Alters the activation function of nodes to better suit different types of data processing needs, adapting to the specific characteristics of the input data.
*/

func setupMutation() {
	getModelLink := "http://localhost:4124/files/localhost:4123/testing/init/01cb7faa-8786-4c2a-82ed-5585b9b80889.json"
	mutateModel(getModelLink)
}

func mutateModel(getModelLink string) {
	// Fetch the neural network configuration
	nnConfigJSON, err := getRequest(getModelLink)
	if err != nil {
		fmt.Println("Error fetching data:", err)
	} else {
		// Unmarshal the JSON into a NetworkConfig struct
		var nnConfig NetworkConfig
		err = json.Unmarshal(nnConfigJSON, &nnConfig)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
		} else {
			// Feed the input values into the neural network
			// Example input values - adjust based on your actual input configuration
			inputValues := map[string]float64{
				"1": 1,
				"2": 0.5,
				"3": 0.75,
			}
			outputs := feedforward(&nnConfig, inputValues)
			fmt.Println(outputs)

			// List of all activation types
			activationTypes := []string{"relu", "sigmoid", "tanh", "softmax", "leaky_relu", "swish", "elu", "selu", "softplus"}

			// Loop over all activation types
			for _, activationType := range activationTypes {
				fmt.Println("changing model ", activationType)
				// Create a deep copy of the nnConfig
				nnConfigCopy := deepcopy.Copy(nnConfig).(NetworkConfig)
				newWeightModel, hasItBeenChanged := randomizeWeightByActivationType(&nnConfigCopy, activationType)
				outputs = feedforward(newWeightModel, inputValues)
				fmt.Println(hasItBeenChanged, outputs)
			}

			// Randomize a random neuron's bias
			fmt.Println("Randomizing a random neuron's bias")
			nnConfigCopy := deepcopy.Copy(nnConfig).(NetworkConfig)
			newBiasModel, hasBiasBeenChanged := randomizeRandomNeuronBias(&nnConfigCopy)
			outputs = feedforward(newBiasModel, inputValues)
			fmt.Println(hasBiasBeenChanged, outputs)
		}
	}
}

func randomizeWeightByActivationType(nnConfig *NetworkConfig, activationType string) (*NetworkConfig, bool) {
	rand.Seed(time.Now().UnixNano())

	// Create a list of neurons with the specified activation type
	var neuronsList []string
	var layerTypes []string

	// Iterate over each layer in hidden layers
	for _, layer := range nnConfig.Layers.Hidden {
		// Iterate over each neuron in the layer
		for neuronID, neuron := range layer.Neurons {
			// If the neuron's activation type matches the provided activation type
			if neuron.ActivationType == activationType {
				neuronsList = append(neuronsList, neuronID)
				layerTypes = append(layerTypes, "hidden")
			}
		}
	}

	// Iterate over each neuron in output layer
	for neuronID, neuron := range nnConfig.Layers.Output.Neurons {
		// If the neuron's activation type matches the provided activation type
		if neuron.ActivationType == activationType {
			neuronsList = append(neuronsList, neuronID)
			layerTypes = append(layerTypes, "output")
		}
	}

	// If there are no neurons with the specified activation type, return the original config and false
	if len(neuronsList) == 0 {
		return nnConfig, false
	}

	// Select a random neuron from the list
	randomIndex := rand.Intn(len(neuronsList))
	randomNeuronID := neuronsList[randomIndex]
	randomLayerType := layerTypes[randomIndex]

	// Randomize one of its weights
	if randomLayerType == "hidden" {
		for i, layer := range nnConfig.Layers.Hidden {
			for neuronID, neuron := range layer.Neurons {
				if neuronID == randomNeuronID {
					// Get all connection IDs
					connectionIDs := make([]string, 0, len(neuron.Connections))
					for connectionID := range neuron.Connections {
						connectionIDs = append(connectionIDs, connectionID)
					}

					// Select a random connection
					randomConnectionID := connectionIDs[rand.Intn(len(connectionIDs))]

					// Randomize the weight of the selected connection
					nnConfig.Layers.Hidden[i].Neurons[neuronID].Connections[randomConnectionID] = Connection{Weight: rand.NormFloat64()}
					return nnConfig, true
				}
			}
		}
	} else if randomLayerType == "output" {
		for neuronID, neuron := range nnConfig.Layers.Output.Neurons {
			if neuronID == randomNeuronID {
				// Get all connection IDs
				connectionIDs := make([]string, 0, len(neuron.Connections))
				for connectionID := range neuron.Connections {
					connectionIDs = append(connectionIDs, connectionID)
				}

				// Select a random connection
				randomConnectionID := connectionIDs[rand.Intn(len(connectionIDs))]

				// Randomize the weight of the selected connection
				nnConfig.Layers.Output.Neurons[neuronID].Connections[randomConnectionID] = Connection{Weight: rand.NormFloat64()}
				return nnConfig, true
			}
		}
	}

	return nnConfig, false
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func randomizeRandomNeuronBias(nnConfig *NetworkConfig) (*NetworkConfig, bool) {
	rand.Seed(time.Now().UnixNano())

	// Create a list of neurons
	var neuronsList []string
	var layerTypes []string

	// Iterate over each layer in hidden layers
	for _, layer := range nnConfig.Layers.Hidden {
		// Iterate over each neuron in the layer
		for neuronID := range layer.Neurons {
			neuronsList = append(neuronsList, neuronID)
			layerTypes = append(layerTypes, "hidden")
		}
	}

	// Iterate over each neuron in output layer
	for neuronID := range nnConfig.Layers.Output.Neurons {
		neuronsList = append(neuronsList, neuronID)
		layerTypes = append(layerTypes, "output")
	}

	// If there are no neurons, return the original config and false
	if len(neuronsList) == 0 {
		return nnConfig, false
	}

	// Select a random neuron from the list
	randomIndex := rand.Intn(len(neuronsList))
	randomNeuronID := neuronsList[randomIndex]
	randomLayerType := layerTypes[randomIndex]

	// Randomize its bias
	if randomLayerType == "hidden" {
		for i, layer := range nnConfig.Layers.Hidden {
			for neuronID := range layer.Neurons {
				if neuronID == randomNeuronID {
					// Get the neuron, randomize the bias, and assign it back to the map
					neuron := nnConfig.Layers.Hidden[i].Neurons[neuronID]
					neuron.Bias = rand.NormFloat64()
					nnConfig.Layers.Hidden[i].Neurons[neuronID] = neuron
					return nnConfig, true
				}
			}
		}
	} else if randomLayerType == "output" {
		for neuronID := range nnConfig.Layers.Output.Neurons {
			if neuronID == randomNeuronID {
				// Get the neuron, randomize the bias, and assign it back to the map
				neuron := nnConfig.Layers.Output.Neurons[neuronID]
				neuron.Bias = rand.NormFloat64()
				nnConfig.Layers.Output.Neurons[neuronID] = neuron
				return nnConfig, true
			}
		}
	}

	return nnConfig, false
}