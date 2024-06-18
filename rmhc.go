package main

import (
	"fmt" 
	"strings"
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


				createNewGeneration(selectedComputer,selectedProject,selectedDataPath,selectedComputerDataCache,"eval_init")
				

			} else {
				fmt.Println("The key 'eval_init' does not exist in the evalFolders map.")
			}
		}

		break
	}
}


func createNewGeneration(selectedComputer string, selectedProject string, selectedDataPath string, selectedComputerDataCache string, folder string) {
    // http://localhost:4123/files/testing/eval_init/ranking.json
	// http://localhost:4123/files/testing
	//http://localhost:4123/files/testingeval_init/ranking.json
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
			fmt.Println(rankingsMap)

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

			// Loop over the first 10 entries
			for i, entry := range rankingSlice {
				if i >= 10 {
					break
				}
				fmt.Printf("Entry %d: %v\n", i+1, entry)
			}
		}
	}
}
