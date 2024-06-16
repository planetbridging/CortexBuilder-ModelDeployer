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
successfully done testing - Add new neuron with random bias, connections, activation function and weight
successfully done testing - Add Connection Mutation: Creates a new connection between previously unconnected nodes, expanding the network's capacity for diverse interactions.
Connection Enable/Disable: Toggles the enabled state of connections, allowing the network to experiment with different neural pathways without permanent structural changes. (Will be developed later)
successfully done testing - Add Layer Mutation: Introduces entirely new layers to the network, significantly enhancing its depth and functional complexity.
successfully done testing - Activation Function Mutation: Alters the activation function of nodes to better suit different types of data processing needs, adapting to the specific characteristics of the input data.
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

			// Randomize a random neuron's activation function
			fmt.Println("Randomizing a random neuron's activation function")
			nnConfigCopy = deepcopy.Copy(nnConfig).(NetworkConfig)
			newActivationModel, hasActivationBeenChanged := randomizeRandomNeuronActivationType(&nnConfigCopy)
			outputs = feedforward(newActivationModel, inputValues)
			fmt.Println(hasActivationBeenChanged, outputs)

			/*nnConfigCopyBuilding := deepcopy.Copy(nnConfig).(NetworkConfig)

			for i := 0; i < 100; i++ {
				// Randomize a random neuron's activation function
				newNeuronModel, hasNewNeron := addRandomNeuronToHiddenLayer(&nnConfigCopyBuilding)
				outputs := feedforward(newNeuronModel, inputValues)
				fmt.Println(hasNewNeron, outputs)
				nnConfigCopyBuilding = *newNeuronModel
				//fmt.Println(newNeuronModel)
			}*/

			for i := 0; i < 5; i++ {
				// Randomize a random neuron's activation function
				fmt.Println("Randomizing a random connection")
				nnConfigCopy = deepcopy.Copy(nnConfig).(NetworkConfig)
				newRandomConnection, hasRandomConnectionBeenAdded := addRandomConnection(&nnConfigCopy)
				outputs = feedforward(newRandomConnection, inputValues)
				fmt.Println(hasRandomConnectionBeenAdded, outputs)
			}

			fmt.Println("Randomizing new layer")
			nnConfigCopy = deepcopy.Copy(nnConfig).(NetworkConfig)
			newModelLayer, hasNewLayer := addRandomHiddenLayer(&nnConfigCopy)
			outputs = feedforward(newModelLayer, inputValues)
			fmt.Println(hasNewLayer, outputs)
			fmt.Println(newModelLayer)
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

func randomizeRandomNeuronActivationType(nnConfig *NetworkConfig) (*NetworkConfig, bool) {
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

	// List of all activation types
	activationTypes := []string{"relu", "sigmoid", "tanh", "softmax", "leaky_relu", "swish", "elu", "selu", "softplus"}

	// Change its activation type
	if randomLayerType == "hidden" {
		for i, layer := range nnConfig.Layers.Hidden {
			for neuronID, _ := range layer.Neurons {
				if neuronID == randomNeuronID {
					// Get the neuron, change the activation type to a different one, and assign it back to the map
					neuron := nnConfig.Layers.Hidden[i].Neurons[neuronID]
					neuron.ActivationType = getRandomDifferentActivationType(neuron.ActivationType, activationTypes)
					nnConfig.Layers.Hidden[i].Neurons[neuronID] = neuron
					return nnConfig, true
				}
			}
		}
	} else if randomLayerType == "output" {
		for neuronID, _ := range nnConfig.Layers.Output.Neurons {
			if neuronID == randomNeuronID {
				// Get the neuron, change the activation type to a different one, and assign it back to the map
				neuron := nnConfig.Layers.Output.Neurons[neuronID]
				neuron.ActivationType = getRandomDifferentActivationType(neuron.ActivationType, activationTypes)
				nnConfig.Layers.Output.Neurons[neuronID] = neuron
				return nnConfig, true
			}
		}
	}

	return nnConfig, false
}

// Helper function to get a random activation type that is different from the current one
func getRandomDifferentActivationType(currentType string, activationTypes []string) string {
	for {
		newType := activationTypes[rand.Intn(len(activationTypes))]
		if newType != currentType {
			return newType
		}
	}
}

func addRandomNeuronToHiddenLayer(nnConfig *NetworkConfig) (*NetworkConfig, bool) {
	rand.Seed(time.Now().UnixNano())

	// List of all activation types
	activationTypes := []string{"relu", "sigmoid", "tanh", "softmax", "leaky_relu", "swish", "elu", "selu", "softplus"}

	// Select a random activation type for the new neuron
	activationType := activationTypes[rand.Intn(len(activationTypes))]

	// Create a new neuron with a random bias and the selected activation type
	newNeuron := Neuron{
		ActivationType: activationType,
		Bias:           rand.NormFloat64(),
		Connections:    make(map[string]Connection),
	}

	// Add random connections to the new neuron from the input layer
	for inputID := range nnConfig.Layers.Input.Neurons {
		if rand.Float64() < 0.5 { // 50% chance to connect to each input neuron
			newNeuron.Connections[inputID] = Connection{Weight: rand.NormFloat64()}
		}
	}

	// Add random connections to the new neuron from the hidden layers
	for _, layer := range nnConfig.Layers.Hidden {
		for neuronID := range layer.Neurons {
			if rand.Float64() < 0.5 { // 50% chance to connect to each neuron in the hidden layers
				newNeuron.Connections[neuronID] = Connection{Weight: rand.NormFloat64()}
			}
		}
	}

	// Add the new neuron to a random hidden layer
	randomLayerIndex := rand.Intn(len(nnConfig.Layers.Hidden))
	newNeuronID := fmt.Sprintf("n%d", countNeuronsPlusOne(nnConfig))
	nnConfig.Layers.Hidden[randomLayerIndex].Neurons[newNeuronID] = newNeuron

	return nnConfig, true
}

func addRandomConnection(nnConfig *NetworkConfig) (*NetworkConfig, bool) {
	rand.Seed(time.Now().UnixNano())

	// Create a list of potential source neurons (from input and hidden layers)
	var sourceNeurons []string
	for neuronID := range nnConfig.Layers.Input.Neurons {
		sourceNeurons = append(sourceNeurons, neuronID)
	}
	for _, layer := range nnConfig.Layers.Hidden {
		for neuronID := range layer.Neurons {
			sourceNeurons = append(sourceNeurons, neuronID)
		}
	}

	// Create a list of potential target neurons (from hidden and output layers)
	var targetNeurons []string
	var targetLayerTypes []string
	for _, layer := range nnConfig.Layers.Hidden {
		for neuronID := range layer.Neurons {
			targetNeurons = append(targetNeurons, neuronID)
			targetLayerTypes = append(targetLayerTypes, "hidden")
		}
	}
	for neuronID := range nnConfig.Layers.Output.Neurons {
		targetNeurons = append(targetNeurons, neuronID)
		targetLayerTypes = append(targetLayerTypes, "output")
	}

	// Shuffle the target neurons and layer types
	rand.Shuffle(len(targetNeurons), func(i, j int) {
		targetNeurons[i], targetNeurons[j] = targetNeurons[j], targetNeurons[i]
		targetLayerTypes[i], targetLayerTypes[j] = targetLayerTypes[j], targetLayerTypes[i]
	})

	// Iterate through the target neurons and try to add a connection
	for i, targetNeuronID := range targetNeurons {
		// Randomly select a source neuron
		sourceNeuronID := sourceNeurons[rand.Intn(len(sourceNeurons))]

		// Check if the connection already exists
		connectionExists := false
		if targetLayerTypes[i] == "hidden" {
			for _, layer := range nnConfig.Layers.Hidden {
				if _, ok := layer.Neurons[targetNeuronID].Connections[sourceNeuronID]; ok {
					connectionExists = true
					break
				}
			}
		} else if targetLayerTypes[i] == "output" {
			if _, ok := nnConfig.Layers.Output.Neurons[targetNeuronID].Connections[sourceNeuronID]; ok {
				connectionExists = true
			}
		}

		// If the connection doesn't exist, add it and return
		if !connectionExists {
			if targetLayerTypes[i] == "hidden" {
				for j, layer := range nnConfig.Layers.Hidden {
					if _, ok := layer.Neurons[targetNeuronID]; ok {
						nnConfig.Layers.Hidden[j].Neurons[targetNeuronID].Connections[sourceNeuronID] = Connection{Weight: rand.NormFloat64()}
						return nnConfig, true
					}
				}
			} else if targetLayerTypes[i] == "output" {
				nnConfig.Layers.Output.Neurons[targetNeuronID].Connections[sourceNeuronID] = Connection{Weight: rand.NormFloat64()}
				return nnConfig, true
			}
		}
	}

	// If no new connection was added, return the original model and false
	return nnConfig, false
}

func countNeurons(nnConfig *NetworkConfig) int {
	// Initialize the count with the number of neurons in the input layer
	count := len(nnConfig.Layers.Input.Neurons)

	// Add the number of neurons in each hidden layer
	for _, layer := range nnConfig.Layers.Hidden {
		count += len(layer.Neurons)
	}

	// Add the number of neurons in the output layer
	count += len(nnConfig.Layers.Output.Neurons)

	// Return the total count plus one
	return count
}

func countNeuronsPlusOne(nnConfig *NetworkConfig) int {
	// Initialize the count with the number of neurons in the input layer
	count := len(nnConfig.Layers.Input.Neurons)

	// Add the number of neurons in each hidden layer
	for _, layer := range nnConfig.Layers.Hidden {
		count += len(layer.Neurons)
	}

	// Add the number of neurons in the output layer
	count += len(nnConfig.Layers.Output.Neurons)

	// Return the total count plus one
	return count + 1
}

func addRandomHiddenLayer(nnConfig *NetworkConfig) (*NetworkConfig, bool) {
	rand.Seed(time.Now().UnixNano())

	// List of all activation types
	activationTypes := []string{"relu", "sigmoid", "tanh", "softmax", "leaky_relu", "swish", "elu", "selu", "softplus"}

	// Select a random activation type for the new neuron
	activationType := activationTypes[rand.Intn(len(activationTypes))]

	// Create a new neuron with a random bias and the selected activation type
	newNeuron := Neuron{
		ActivationType: activationType,
		Bias:           rand.NormFloat64(),
		Connections:    make(map[string]Connection),
	}

	// Create a new hidden layer and add the new neuron to it
	newLayer := Layer{
		Neurons: map[string]Neuron{
			fmt.Sprintf("n%d", countNeuronsPlusOne(nnConfig)): newNeuron,
		},
	}

	// Insert the new layer at a random position in the hidden layers
	insertIndex := rand.Intn(len(nnConfig.Layers.Hidden) + 1)
	nnConfig.Layers.Hidden = append(nnConfig.Layers.Hidden, Layer{})
	copy(nnConfig.Layers.Hidden[insertIndex+1:], nnConfig.Layers.Hidden[insertIndex:])
	nnConfig.Layers.Hidden[insertIndex] = newLayer

	// Add a random connection from the new neuron to a random neuron in the input or hidden layers
	randomNeuronID := ""
	randomLayerType := ""
	if rand.Float64() < 0.5 { // 50% chance to connect to a neuron in the input layer
		inputNeuronIDs := make([]string, 0, len(nnConfig.Layers.Input.Neurons))
		for neuronID := range nnConfig.Layers.Input.Neurons {
			inputNeuronIDs = append(inputNeuronIDs, neuronID)
		}
		randomNeuronID = inputNeuronIDs[rand.Intn(len(inputNeuronIDs))]
		randomLayerType = "input"
	} else { // 50% chance to connect to a neuron in the hidden layers
		hiddenNeuronIDs := make([]string, 0)
		for _, layer := range nnConfig.Layers.Hidden {
			for neuronID := range layer.Neurons {
				hiddenNeuronIDs = append(hiddenNeuronIDs, neuronID)
			}
		}
		randomNeuronID = hiddenNeuronIDs[rand.Intn(len(hiddenNeuronIDs))]
		randomLayerType = "hidden"
	}

	// Ensure the new neuron has at least one connection
	newNeuronID := fmt.Sprintf("n%d", countNeuronsPlusOne(nnConfig))
	if randomLayerType == "input" {
		neuron := nnConfig.Layers.Input.Neurons[randomNeuronID]
		if neuron.Connections == nil {
			neuron.Connections = make(map[string]Connection)
		}
		neuron.Connections[newNeuronID] = Connection{Weight: rand.NormFloat64()}
		nnConfig.Layers.Input.Neurons[randomNeuronID] = neuron
	} else if randomLayerType == "hidden" {
		for i, layer := range nnConfig.Layers.Hidden {
			if _, ok := layer.Neurons[randomNeuronID]; ok {
				neuron := nnConfig.Layers.Hidden[i].Neurons[randomNeuronID]
				if neuron.Connections == nil {
					neuron.Connections = make(map[string]Connection)
				}
				neuron.Connections[newNeuronID] = Connection{Weight: rand.NormFloat64()}
				nnConfig.Layers.Hidden[i].Neurons[randomNeuronID] = neuron
				break
			}
		}
	}

	return nnConfig, true
}
