package main

import "fmt"

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
			fmt.Println(nonEvalFolders)
			fmt.Println(evalFolders)

			if _, ok := evalFolders["eval_init"]; ok {
				fmt.Println("The key 'yourKey' exists in the evalFolders map.")
			} else {
				fmt.Println("The key 'yourKey' does not exist in the evalFolders map.")
			}
		}

		break
	}
}
