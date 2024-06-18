package main

import (
	"fmt" 
	"strings"
	"math"
	"math/rand"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
)

func testingSetupRMHC() {
	runRMHC("localhost", "/path/testing", "./host/data.csv", 10)
}

func printOutMap(js map[string]interface{}) {
	fmt.Println("-------printOutMap-------")
	for key, value := range js {
		fmt.Printf("Key: %s, Value: %v\n", key, value)
	}
}

func runRMHC(selectedComputer string, selectedProject string, selectedDataPath string, amountOfGenerations int) {
	selectedComputerDataCache := selectedComputer + ":4123"
	for i := 0; i < amountOfGenerations; i++ {
		dataProjectItemsInterface, err := parseJSONFromURL("http://" + selectedComputerDataCache + selectedProject)
		if err != nil {
			fmt.Println("runRMHC error:", err)
		} else {
			//fmt.Println(dataProjectItemsInterface)

			js := map[string]interface{}{
				"selectedComputer": selectedComputer,
				"selectedDataPath": selectedDataPath,
				"selectedProject":  selectedProject,
				"testing":          30,
				"training":         70,
				"clientID":         "127.0.0.1:36478",
				"aiPod":            "localhost:12346",
			}
			//fmt.Println(js)
			startEval(js)

			dataProjectItems, dataProjectItemsOk := dataProjectItemsInterface.([]map[string]interface{})
			if dataProjectItemsOk {
				fmt.Println(dataProjectItems)
			}

			ranking(selectedComputer, selectedProject)

			nonEvalFolders, evalFolders := getEvalFolders(dataProjectItems)
			//fmt.Println(nonEvalFolders)
			//fmt.Println(evalFolders)
			_ = nonEvalFolders
			if _, ok := evalFolders["eval_init"]; ok {
				fmt.Println("The key 'eval_init' exists in the evalFolders map.")


				createNewGeneration(selectedComputer,selectedProject,selectedDataPath,selectedComputerDataCache,"eval_init", i)
				

			} else {
				fmt.Println("The key 'eval_init' does not exist in the evalFolders map.")
			}

			if _, ok := evalFolders["eval_" + strconv.Itoa(i)]; ok {
				createNewGeneration(selectedComputer,selectedProject,selectedDataPath,selectedComputerDataCache,"eval_" + strconv.Itoa(i), i + 1)
			} 
		}

		//break
	}
}


func createNewGeneration(selectedComputer string, selectedProject string, selectedDataPath string, selectedComputerDataCache string, folder string, newFolderNumber int) {
    // http://localhost:4123/files/testing/eval_init/ranking.json
	// http://localhost:4123/files/testing
	//http://localhost:4123/files/testingeval_init/ranking.json

	createFolder(selectedProject, strconv.Itoa(newFolderNumber), selectedComputerDataCache)

    modifiedProject := strings.Replace(selectedProject, "/path/", "/files/", -1) + "/"+folder + "/ranking.json"

    getOldRankingsUrl := "http://" + selectedComputerDataCache + modifiedProject
    fmt.Println(getOldRankingsUrl)
	getRankings, getRankingsErr := parseJSONFromURL(getOldRankingsUrl)
	if getRankingsErr != nil{
		fmt.Println(getRankingsErr)
	}else{
		rankingsMap, ok := getRankings.(map[string]interface{})
		if !ok {
			fmt.Println("Error: Unable to assert getRankings as map[string]interface{}")
		}else{
			//fmt.Println(rankingsMap)

			// Access the "lstlstRanking" key
			lstRanking, lstlstRankingExists := rankingsMap["lstRanking"]
			if !lstlstRankingExists {
				fmt.Println("Error: 'lstlstRanking' key not found in the map.")
				return
			}

			// Type assertion to ensure lstlstRanking is a slice of interfaces
			rankingSlice, ok := lstRanking.([]interface{})
			if !ok {
				fmt.Println("Error: Unable to assert lstRanking as []interface{}")
				return
			}
			groups := int(math.Ceil(float64(len(rankingSlice)) / 10))
			fmt.Printf("Model count: %d, splitting into %d groups\n", len(rankingSlice), groups)

			lstMutations := []string{
				"weightsMutation",
				"biasMutation",
				"addNewNeuronMutation",
				"addConnectionMutation",
				"addLayerMutation",
				"activationFunctionMutation",
			}
			fmt.Println("Randomly choosing mutation from", lstMutations)

			modelsFolder := strings.Replace(folder, "eval_", "", 1)
			modelStartingLink :=  "http://" + selectedComputerDataCache + strings.Replace(selectedProject, "/path/", "/files/", -1) + "/"+modelsFolder + "/"
			fmt.Println(modelStartingLink,newFolderNumber)
			// Loop over the first 10 entries
			for i, entry := range rankingSlice {
				if i >= 10 {
					break
				}
				fmt.Printf("Entry %d: %v\n", i+1, entry)
				

				entryMap, ok := entry.(map[string]interface{})
				if !ok {
					fmt.Println("Error: Unable to assert entry as map[string]interface{}")
				}else{
					fileName, ok := entryMap["file_name"].(string)
				if !ok {
					fmt.Println("Error: Unable to extract file_name from entry")
						
					}else{
						fmt.Println("File name:", fileName)
						getModelLink := modelStartingLink + fileName
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
								fmt.Println(nnConfig)

								for m := 0; m < groups; m++ {
									randomIndex := rand.Intn(len(lstMutations))

									var newWeightModel *NetworkConfig
									var hasItBeenChanged bool

									switch lstMutations[randomIndex]{
										case "weightsMutation":
											/*lstUsingActiveFunctions := getActivationFunctions(&nnConfig)
											fmt.Println("-----------------------",lstUsingActiveFunctions)
											if len(lstUsingActiveFunctions) > 0 {
												randomSelectActivationFunction := rand.Intn(len(lstUsingActiveFunctions))
												newWeightModelTmp, hasItBeenChangedTmp := randomizeWeightByActivationType(&nnConfig, lstUsingActiveFunctions[randomSelectActivationFunction])
												newWeightModel = newWeightModelTmp
												hasItBeenChanged = hasItBeenChangedTmp
											}*/
											
										case "biasMutation":
											newWeightModelTmp, hasItBeenChangedTmp := randomizeRandomNeuronBias(&nnConfig)
											newWeightModel = newWeightModelTmp
											hasItBeenChanged = hasItBeenChangedTmp
										case "addNewNeuronMutation":
											newWeightModelTmp, hasItBeenChangedTmp := addRandomNeuronToHiddenLayer(&nnConfig)
											newWeightModel = newWeightModelTmp
											hasItBeenChanged = hasItBeenChangedTmp
										case "addConnectionMutation":
											newWeightModelTmp, hasItBeenChangedTmp := addRandomConnection(&nnConfig)
											newWeightModel = newWeightModelTmp
											hasItBeenChanged = hasItBeenChangedTmp
										case "addLayerMutation":
											newWeightModelTmp, hasItBeenChangedTmp := addRandomHiddenLayer(&nnConfig)
											newWeightModel = newWeightModelTmp
											hasItBeenChanged = hasItBeenChangedTmp
										case "activationFunctionMutation":
											newWeightModelTmp, hasItBeenChangedTmp := randomizeRandomNeuronActivationType(&nnConfig)
											newWeightModel = newWeightModelTmp
											hasItBeenChanged = hasItBeenChangedTmp
									}

									

									

									if hasItBeenChanged{
										fmt.Println(newWeightModel)
										modelName := uuid.New()
										saveEvalLocation := selectedProject + "/" + strconv.Itoa(newFolderNumber) + "/" + modelName.String() + ".json"
										if strings.HasPrefix(saveEvalLocation, "/path/") {
											saveEvalLocation = strings.TrimPrefix(saveEvalLocation, "/path/")
										}

										convertModel,err := convertToMap(*newWeightModel)
										if err != nil{
											fmt.Println("failed to convert model")
										}else{
											saveDataToFile(saveEvalLocation, selectedComputerDataCache, convertModel)
										}											
									}
									//break
								}

							}
						}
					}
				}


				
				

			}
		}
	}
}


func getActivationFunctions(config *NetworkConfig) []string {
	activationFunctions := make([]string, 0)

	// Collect activation functions from hidden layers
	for _, layer := range config.Layers.Hidden {
		for _, node := range layer.Neurons {
			activationFunctions = append(activationFunctions, node.ActivationType)
		}
	}

	// Collect activation function from output layer
	for _, node := range config.Layers.Output.Neurons {
		activationFunctions = append(activationFunctions, node.ActivationType)
	}

	return activationFunctions
}

func convertToMap(newWeightModel NetworkConfig) (map[string]interface{}, error) {
	modelMap := make(map[string]interface{})

	// Create the layers map
	layersMap := make(map[string]interface{})

	// Convert Hidden layers
	hiddenLayers := make([]map[string]interface{}, len(newWeightModel.Layers.Hidden))
	for i, hiddenLayer := range newWeightModel.Layers.Hidden {
		hiddenNeurons := make(map[string]interface{})
		for neuronID, neuron := range hiddenLayer.Neurons {
			hiddenNeurons[neuronID] = map[string]interface{}{
				"activationType": neuron.ActivationType,
				"bias":           neuron.Bias,
				"connections":    neuron.Connections,
			}
		}
		hiddenLayers[i] = map[string]interface{}{"neurons": hiddenNeurons}
	}
	layersMap["hidden"] = hiddenLayers

	// Convert Input layer
	inputNeurons := make(map[string]interface{})
	for inputID, inputNeuron := range newWeightModel.Layers.Input.Neurons {
		inputNeurons[inputID] = map[string]interface{}{
			"activationType": inputNeuron.ActivationType,
			"bias":           inputNeuron.Bias,
		}
	}
	layersMap["input"] = map[string]interface{}{"neurons": inputNeurons}

	// Convert Output layer
	outputNeurons := make(map[string]interface{})
	for outputID, outputNeuron := range newWeightModel.Layers.Output.Neurons {
		outputNeurons[outputID] = map[string]interface{}{
			"activationType": outputNeuron.ActivationType,
			"bias":           outputNeuron.Bias,
			"connections":    outputNeuron.Connections,
		}
	}
	layersMap["output"] = map[string]interface{}{"neurons": outputNeurons}

	modelMap["layers"] = layersMap

	return modelMap, nil
}