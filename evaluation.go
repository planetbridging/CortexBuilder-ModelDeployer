package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func testingEval() {
	envSavedDataCache := os.Getenv("envSavedDataCache")

	if envSavedDataCache != "" {
		fmt.Println("Running testing eval:", envSavedDataCache)
		envSavedDataCachePort := envSavedDataCache + ":4123"

		dataConfigInterface, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/files/config.json")
		dataConfig, ok := dataConfigInterface.(map[string]interface{})

		if !ok {
			fmt.Println("Error: config data is not an object")
			return
		}

		dataMounted, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/mounted")
		dataMountedConfig, ok := dataMounted.([]map[string]interface{})
		fmt.Println(dataMountedConfig)
		if !ok {
			fmt.Println("Error: project items data is not an array of objects")
			return
		}

		setProjectPath, ok := dataConfig["setProjectPath"].(string)
		if !ok {
			fmt.Println("Error: setProjectPath is not a string")
			return
		}

		dataProjectItemsInterface, err := parseJSONFromURL("http://" + envSavedDataCachePort + setProjectPath)
		if err != nil {
			fmt.Println("Error fetching data:", err)
			return
		}

		dataProjectItems, ok := dataProjectItemsInterface.([]map[string]interface{})
		if !ok {
			fmt.Println("Error: project items data is not an array of objects")
			return
		}

		nonEvalFolders := make(map[string]bool)
		evalFolders := make(map[string]bool)
		for _, item := range dataProjectItems {
			name, ok := item["name"].(string)
			if ok {
				if strings.HasPrefix(name, "eval_") {
					evalFolders[name] = true
				} else {
					nonEvalFolders[name] = true
				}
			}
		}

		for folder := range nonEvalFolders {
			newFolderItems, err := parseJSONFromURL("http://" + envSavedDataCachePort + setProjectPath + "/" + folder)
			if err != nil {
				fmt.Println("Error fetching data:", err)
				return
			}

			for _, item := range newFolderItems.([]map[string]interface{}) {
				itemName, ok := item["name"].(string)
				if ok {
					itemURL := "http://" + envSavedDataCachePort + strings.Replace(setProjectPath, "/path/", "/files/", 1) + "/" + folder + "/" + itemName
					fmt.Println(item)
					fmt.Println(itemURL)

					// Fetch the neural network configuration
					nnConfigJSON, err := getRequest(itemURL)
					if err != nil {
						fmt.Println("Error fetching data:", err)
						return
					}

					// Unmarshal the JSON into a NetworkConfig struct
					var nnConfig NetworkConfig
					err = json.Unmarshal(nnConfigJSON, &nnConfig)
					if err != nil {
						fmt.Println("Error decoding JSON:", err)
						return
					}

					// Loop through dataMountedConfig
					for _, config := range dataMountedConfig {

						cols, okCols := config["cols"].([]interface{})

						rowCount, ok := config["rowCount"].(float64)
						if ok && okCols {
							inputCount := 0
							for _, col := range cols {
								colStr, ok := col.(string)
								if ok && strings.HasPrefix(colStr, "input") {
									inputCount++
								}
							}
							fmt.Println("Number of variables starting with 'input':", inputCount)
							for i := 0; i <= int(rowCount)-1; i++ {
								urlRow := "http://" + envSavedDataCachePort + "/row/?path=" + config["path"].(string) + "&index=" + strconv.Itoa(i)
								fmt.Println(urlRow)
								//http://192.168.0.228:4123/row/?path=./host/data.csv&index=0
								dataMountedRow, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/row/?path=" + config["path"].(string) + "&index=" + strconv.Itoa(i))
								dataMountedRowArray, okCheckRow := dataMountedRow.([]string)
								fmt.Println(dataMountedRowArray)
								if !okCheckRow {
									fmt.Println("Error: project items data is not an array of objects")
									return
								}
							}
						}
					}

					// Define some input values
					inputValues := map[string]float64{
						"1": 0.5,
						"2": 0.6,
						"3": 0.7,
					}

					// Feed the input values into the neural network
					outputs := feedforward(&nnConfig, inputValues)

					// Print the outputs
					fmt.Println(outputs)
				}
				break
			}
		}
	}
}
