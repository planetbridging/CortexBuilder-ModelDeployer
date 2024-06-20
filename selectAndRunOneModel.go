package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func selectAndTestOneModel() {
	getModelLink := "http://localhost:4123/files/testmodel.json"

	nnConfigJSON, err := getRequest(getModelLink)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}

	var nnConfig NetworkConfig
	err = json.Unmarshal(nnConfigJSON, &nnConfig)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	csvFilePath := "./data.csv"
	file, err := os.Create(csvFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"input1", "input2", "input3", "output1", "output2", "output3", "modelOutput1", "modelOutput2", "modelOutput3"}
	writer.Write(headers)

	numRows := 10000
	for i := 1; i <= numRows; i++ {
		inputs := []string{strconv.Itoa(i), strconv.Itoa(i+1), strconv.Itoa(i+2)}
		outputs := []string{strconv.Itoa(2*i), strconv.Itoa(2*(i+1)), strconv.Itoa(2*(i+2))}

		// Compute model outputs (replace with actual logic)
		modelOutputs := feedforward(&nnConfig, map[string]float64{
			"1": float64(i),
			"2": float64(i + 1),
			"3": float64(i + 2),
		})

		fmt.Println(modelOutputs)

		// Append model outputs to existing outputs
		outputs = append(outputs,
			fmt.Sprintf("%.6f", modelOutputs["5"]), // Adjust the key here
			fmt.Sprintf("%.6f", modelOutputs["6"]), // Adjust the key here
			fmt.Sprintf("%.6f", modelOutputs["7"]), // Adjust the key here
		)

		fmt.Printf("Input: %v %v %v | Output: %v %v %v | Model Output: %v %v %v\n",
			inputs[0], inputs[1], inputs[2],
			outputs[0], outputs[1], outputs[2],
			modelOutputs["5"], modelOutputs["6"], modelOutputs["7"])


		data := append(inputs, outputs...)
		writer.Write(data)
	}

	fmt.Println("Data saved to", csvFilePath)
}
