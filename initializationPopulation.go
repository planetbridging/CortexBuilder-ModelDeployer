package main

import (
	"strings"
	"fmt"
	"sync"
	"github.com/google/uuid"
)


func checkAllKeysExist(js map[string]interface{}, keys []string) bool {
	for _, key := range keys {
		if _, ok := js[key]; !ok {
			return false
		}
	}
	return true
}


func initializationPopulation(j map[string]interface{}){
	//{"amount":500,"clientID":"127.0.0.1:60300","dataCache":"localhost:12345","path":"/path/p1","type":"initializationPopulation"}
	keysToCheck := []string{"amount", "clientID","dataCache","path"}
	allKeysExist := checkAllKeysExist(j, keysToCheck)
	if allKeysExist{
		amount, ok := j["amount"].(float64) // type assertion
		if !ok {
			fmt.Println("amount is not a float64")
			return
		}
		amountInt := int(amount)

		clientID := j["clientID"]
		dataCache, ok := j["dataCache"].(string) // type assertion
		if !ok {
			fmt.Println("dataCache is not a string")
			return
		}
		path, ok := j["path"].(string) // type assertion
		if !ok {
			fmt.Println("path is not a string")
			return
		}

		fmt.Println("amount",amount)
		fmt.Println("clientID",clientID)
		fmt.Println("dataCache",dataCache)
		fmt.Println("path",path)

		newPath := strings.ReplaceAll(dataCache, "12345", "4123")

		cleanPath := path
		if strings.HasPrefix(cleanPath, "/path/") {
			cleanPath = strings.TrimPrefix(cleanPath, "/path/")
		}

		url := "http://" + newPath + "/createfolder"
		postData := map[string]interface{}{
			"Path": "./" + cleanPath + "/init",
		}
		response, err := sendPostRequest(url, postData)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response)


		//single model creation example
		/*modelJSON := randomizeNetworkStaticTesting()
		modelName := uuid.New()
		modelPath := cleanPath + "/init/" + modelName.String() + ".json"
		modelCreatePath := "http://"+newPath+"/createfile"
		fmt.Println(modelName)

		// Prepare the data for the POST request
		postDataCreateModel := map[string]interface{}{
			"Path": modelPath, // replace with your actual file path
			"Data": modelJSON,
		}

		resCreateModel, err := sendPostRequest(modelCreatePath, postDataCreateModel)
		if err != nil {
			fmt.Println(err)
		}else{
			// Print the response
			fmt.Println(resCreateModel)
		}*/

		var wg sync.WaitGroup
		for i := 0; i < amountInt; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				modelJSON := randomizeNetworkStaticTesting()
				modelName := uuid.New()
				modelPath := cleanPath + "/init/" + modelName.String() + ".json"
				modelCreatePath := "http://"+newPath+"/createfile"
				fmt.Println(modelName)

				// Prepare the data for the POST request
				postDataCreateModel := map[string]interface{}{
					"Path": modelPath, // replace with your actual file path
					"Data": modelJSON,
				}

				resCreateModel, err := sendPostRequest(modelCreatePath, postDataCreateModel)
				if err != nil {
					fmt.Println(err)
				}else{
					// Print the response
					fmt.Println(resCreateModel)
				}
			}()
		}
		wg.Wait()
	
		fmt.Println("Created bulk init models")
		

	}
}