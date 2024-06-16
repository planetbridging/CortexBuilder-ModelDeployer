package main

import (
	"fmt"
	"sort"
	"strings"
)

type ModelPerformance struct {
	Name                   string
	TrainingDataAccuracy   float64
	PredictionDataAccuracy float64
}

func settingUpRanking() {
	ranking("localhost", "/path/testing")
}

func ranking(selectedComputer string, selectedProject string) {
	selectedComputerDataCache := selectedComputer + ":4123"
	newFolderItems, err := parseJSONFromURL("http://" + selectedComputerDataCache + selectedProject)

	if err != nil {
		fmt.Println(err)
	} else {
		dataProjectItems, dataProjectItemsOk := newFolderItems.([]map[string]interface{})
		_, evalFolders := getEvalFolders(dataProjectItems)

		if dataProjectItemsOk {
			//fmt.Println("----dataProjectItemsok----")
			for folder := range evalFolders {
				//fmt.Println(evalFolders)
				evalContents, errEvalContents := parseJSONFromURL("http://" + selectedComputerDataCache + selectedProject + "/" + folder)
				//fmt.Println(evalContents)
				if errEvalContents != nil {
					fmt.Println(errEvalContents)
				} else {
					if evalContentsSlice, ok := evalContents.([]map[string]interface{}); ok {
						if fileNameExists(evalContentsSlice, "ranking.json") {
							fmt.Println("The file ranking.json exists.")
						} else {
							fmt.Println("The file ranking.json does not exist in " + selectedComputerDataCache + selectedProject + "/" + folder)
							//http://localhost:4123/files/testing/init/01cb7faa-8786-4c2a-82ed-5585b9b80889.json
							createRanking(selectedComputerDataCache, selectedProject, folder, evalContentsSlice)
						}
					} else {
						fmt.Println("Error: evalContents is not of type []map[string]interface{}")
					}

				}
			}
		}

	}
}

func fileNameExists(evalContents []map[string]interface{}, fileName string) bool {
	for _, item := range evalContents {
		if name, ok := item["name"].(string); ok && name == fileName {
			return true
		}
	}
	return false
}

func createRanking(selectedComputerDataCache string, selectedProject string, folder string, evalContents []map[string]interface{}) {
	fmt.Printf("Printing evalContents for project %s on computer %s:\n", selectedProject, selectedComputerDataCache)

	var performances []ModelPerformance

	for _, item := range evalContents {
		if name, ok := item["name"].(string); ok {
			itemURL := "http://" + selectedComputerDataCache + strings.Replace(selectedProject, "/path/", "/files/", 1) + "/" + folder + "/" + name
			getEvalResults, errGetEvalResults := parseJSONFromURL(itemURL)
			if errGetEvalResults != nil {
				fmt.Println("errGetEvalResults", errGetEvalResults)
			} else {
				if evalData, ok := getEvalResults.(map[string]interface{}); ok {
					if trainingAccuracy, ok := evalData["Training_Data_Accuracy"].(float64); ok {
						if predictionAccuracy, ok := evalData["Prediction_Data_Accuracy"].(float64); ok {
							performances = append(performances, ModelPerformance{Name: name, TrainingDataAccuracy: trainingAccuracy, PredictionDataAccuracy: predictionAccuracy})
						}
					}
				}
			}
		} else {
			fmt.Println("Error: item name is not of type string")
		}
	}

	// Sort performances slice based on TrainingDataAccuracy
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].TrainingDataAccuracy > performances[j].TrainingDataAccuracy
	})

	// Create ranking
	var lstRanking []map[string]interface{}
	for i, performance := range performances {
		rank := map[string]interface{}{
			"rank":                     i,
			"file_name":                performance.Name,
			"Training_Data_Accuracy":   performance.TrainingDataAccuracy,
			"Prediction_Data_Accuracy": performance.PredictionDataAccuracy,
		}
		lstRanking = append(lstRanking, rank)
	}

	// Create postData map
	postData := map[string]interface{}{
		"lstRanking": lstRanking,
	}

	// Save ranking to file
	saveEvalLocation := selectedProject + "/" + folder + "/ranking.json"
	if strings.HasPrefix(saveEvalLocation, "/path/") {
		saveEvalLocation = strings.TrimPrefix(saveEvalLocation, "/path/")
	}
	saveDataToFile(saveEvalLocation, selectedComputerDataCache, postData)
}
