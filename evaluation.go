package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
							outputCount := 0
							for _, col := range cols {
								colStr, ok := col.(string)
								if ok && strings.HasPrefix(colStr, "input") {
									inputCount++
								}

								if ok && strings.HasPrefix(colStr, "output") {
									outputCount++
								}
							}
							fmt.Println("Number of variables starting with 'input':", inputCount)

							trainingPercent := 70.0
							predictionPercent := 30.0
							totalRows := int(rowCount)

							trainingRows, predictionRows := calculateRows(trainingPercent, predictionPercent, totalRows)
							fmt.Printf("Training rows: %d, Prediction rows: %d\n", trainingRows, predictionRows)

							start := time.Now()
							errorFloat, accFloat := multiThreadedFullData(0, envSavedDataCachePort, config["path"].(string), trainingRows, inputCount, outputCount, &nnConfig)

							fmt.Printf("Training Data Mean Absolute Percentage Error: %f%%\n", errorFloat)
							fmt.Printf("Training Data Accuracy: %f%%\n", accFloat)
							elapsed := time.Since(start)
							fmt.Printf("The section of code took with http %s to execute.\n", elapsed)

							start = time.Now()
							errorFloatPred, accFloatPred := multiThreadedFullData(int(rowCount)-predictionRows, envSavedDataCachePort, config["path"].(string), int(rowCount), inputCount, outputCount, &nnConfig)

							fmt.Printf("Prediction Data Mean Absolute Percentage Error: %f%%\n", errorFloatPred)
							fmt.Printf("Prediction Data Accuracy: %f%%\n", accFloatPred)
							elapsed = time.Since(start)
							fmt.Printf("The section of code took with http %s to execute.\n", elapsed)

							fmt.Println("ROW COUNT:", rowCount)
							//loopThroughData(config["path"].(string), int(rowCount))
						}
					}

				}
				break
			}
		}
	}
}

func singleThreadedFullData(envSavedDataCachePort string, configPath string, rowCount int, inputCount int, outputCount int, nnConfig *NetworkConfig) {
	fmt.Println(inputCount, outputCount)

	var totalPercentageDifference float64 = 0

	for i := 0; i <= rowCount-1; i++ {
		urlRow := "http://" + envSavedDataCachePort + "/row/?path=" + configPath + "&index=" + strconv.Itoa(i)
		fmt.Println(urlRow)
		//http://192.168.0.228:4123/row/?path=./host/data.csv&index=0
		dataMountedRow, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/row/?path=" + configPath + "&index=" + strconv.Itoa(i))
		dataMountedRowArray, okCheckRow := dataMountedRow.([]string)

		inputValues := map[string]float64{}

		for inVal := 0; inVal <= inputCount-1; inVal++ {
			strVal := strconv.Itoa(inVal + 1)
			floatVal, err := strconv.ParseFloat(dataMountedRowArray[inVal], 64)
			if err != nil {
				fmt.Println("Error parsing float:", err)
				return
			}
			inputValues[strVal] = floatVal
		}

		fmt.Println(dataMountedRowArray)
		if !okCheckRow {
			fmt.Println("Error: project items data is not an array of objects")
			return
		}

		// Feed the input values into the neural network
		outputs := feedforward(nnConfig, inputValues)

		// Print the outputs
		fmt.Println(outputs)

		// Compare the outputs to the actual values
		for outVal := inputCount; outVal < inputCount+outputCount; outVal++ {
			actualVal, err := strconv.ParseFloat(dataMountedRowArray[outVal], 64)
			if err != nil {
				fmt.Println("Error parsing float:", err)
				return
			}
			predictedVal := outputs[strconv.Itoa(outVal+1)]
			percentageDifference := math.Abs(actualVal-predictedVal) / actualVal
			totalPercentageDifference += percentageDifference
		}

		//break
	}

	// Calculate and print the Mean Absolute Percentage Error (MAPE)
	mape := totalPercentageDifference / float64(rowCount*outputCount) * 100
	fmt.Printf("Mean Absolute Percentage Error: %f%%\n", mape)

	// Calculate and print the accuracy as a percentage
	accuracy := 100 - mape
	fmt.Printf("Accuracy: %f%%\n", accuracy)
}

func multiThreadedFullData(startingRow int, envSavedDataCachePort string, configPath string, rowCount int, inputCount int, outputCount int, nnConfig *NetworkConfig) (float64, float64) {
	//fmt.Println(inputCount, outputCount)

	var totalPercentageDifference float64
	var mu sync.Mutex
	var wg sync.WaitGroup
	var totalRows int // New variable to keep track of the number of rows processed
	fmt.Println(startingRow, rowCount)
	for i := startingRow; i <= rowCount-1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			//urlRow := "http://" + envSavedDataCachePort + "/row/?path=" + configPath + "&index=" + strconv.Itoa(i)
			//fmt.Println(urlRow)
			dataMountedRow, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/row/?path=" + configPath + "&index=" + strconv.Itoa(i))
			dataMountedRowArray, okCheckRow := dataMountedRow.([]string)

			inputValues := map[string]float64{}

			for inVal := 0; inVal <= inputCount-1; inVal++ {
				strVal := strconv.Itoa(inVal + 1)
				floatVal, err := strconv.ParseFloat(dataMountedRowArray[inVal], 64)
				if err != nil {
					fmt.Println("Error parsing float:", err)
					return
				}
				inputValues[strVal] = floatVal
			}

			//fmt.Println(dataMountedRowArray)
			if !okCheckRow {
				fmt.Println("Error: project items data is not an array of objects")
				return
			}

			// Feed the input values into the neural network
			outputs := feedforward(nnConfig, inputValues)

			// Print the outputs
			//fmt.Println(outputs)

			// Compare the outputs to the actual values
			var rowPercentageDifference float64
			for outVal := inputCount; outVal < inputCount+outputCount; outVal++ {
				actualVal, err := strconv.ParseFloat(dataMountedRowArray[outVal], 64)
				if err != nil {
					fmt.Println("Error parsing float:", err)
					return
				}
				predictedVal := outputs[strconv.Itoa(outVal+1)]
				percentageDifference := math.Abs(actualVal-predictedVal) / actualVal
				rowPercentageDifference += percentageDifference
			}

			mu.Lock()
			totalPercentageDifference += rowPercentageDifference
			totalRows++ // Increment the total number of rows processed
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Calculate and print the Mean Absolute Percentage Error (MAPE)
	//mape := totalPercentageDifference / float64(rowCount*outputCount) * 100
	//fmt.Printf("Mean Absolute Percentage Error: %f%%\n", mape)

	var mape float64
	if startingRow == 0 { // If it's the training phase
		mape = totalPercentageDifference / float64(rowCount*outputCount) * 100
	} else { // If it's the prediction phase
		mape = totalPercentageDifference / float64(totalRows*outputCount) * 100
	}

	// Calculate and print the accuracy as a percentage
	accuracy := 100 - mape
	//fmt.Printf("Accuracy: %f%%\n", accuracy)

	return mape, accuracy
}

func loopThroughData(dataPath string, rowCount int) {

	conn, err := net.Dial("tcp", "192.168.0.228:8923")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	start := time.Now()
	for i := 0; i < rowCount-1; i++ {
		request := fmt.Sprintf("%s %d", dataPath, i)
		fmt.Fprintf(conn, "%s\n", request)

		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print("Received from server: " + message)
	}

	elapsed := time.Since(start)

	fmt.Printf("The section of code took direct tcp %s to execute.\n", elapsed)
}

func calculateRows(trainingPercent float64, predictionPercent float64, totalRows int) (int, int) {
	trainingRows := int(math.Ceil(trainingPercent / 100 * float64(totalRows)))
	predictionRows := int(math.Ceil(predictionPercent / 100 * float64(totalRows)))

	// Adjust if the sum of trainingRows and predictionRows is greater than totalRows
	if trainingRows+predictionRows > totalRows {
		predictionRows = totalRows - trainingRows
	}

	return trainingRows, predictionRows
}

func startEval(js map[string]interface{}) {

	selectedComputer, sCOk := js["selectedComputer"].(string)
	clientID, cIOk := js["clientID"].(string)
	aiPod, aPOk := js["aiPod"].(string)
	selectedDataPath, sdpOk := js["selectedDataPath"].(string)
	selectedProject, spOk := js["selectedProject"].(string)
	testing, testOk := js["testing"].(float64)
	training, trainOk := js["training"].(float64)

	//selectedProject:/path/testing

	if sCOk && cIOk && aPOk && sdpOk && spOk && testOk && trainOk {
		selectedComputerDataCache := selectedComputer + ":4123"

		dataConfigInterface, _ := parseJSONFromURL("http://" + selectedComputerDataCache + "/files/config.json")
		dataConfig, okDataConfig := dataConfigInterface.(map[string]interface{})

		dataMounted, _ := parseJSONFromURL("http://" + selectedComputerDataCache + "/mounted")
		dataMountedConfig, dataMountedConfigOk := dataMounted.([]map[string]interface{})

		if okDataConfig && dataMountedConfigOk {
			fmt.Println(dataConfig)
			fmt.Println(dataMountedConfig)

			dataProjectItemsInterface, err := parseJSONFromURL("http://" + selectedComputerDataCache + selectedProject)
			//fmt.Println("http://" + selectedComputerDataCache + selectedProject)
			if err != nil {
				fmt.Println("Error fetching data:", err)
			} else {
				dataProjectItems, dataProjectItemsOk := dataProjectItemsInterface.([]map[string]interface{})
				if dataProjectItemsOk {
					fmt.Println(dataProjectItems)
				} else {
					fmt.Println("Failed to get folders in project")
				}
			}

		}
	} else {
		fmt.Println("Failed to extract information from json")
	}

	fmt.Println(selectedComputer, sCOk)
	fmt.Println(clientID, cIOk)
	fmt.Println(aiPod, aPOk)
	fmt.Println(selectedDataPath, sdpOk)
	fmt.Println(selectedProject, spOk)
	fmt.Println(testing, testOk)
	fmt.Println(training, trainOk)
}
