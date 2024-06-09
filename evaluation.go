package main

import (
	"fmt"
	"os"
)

func testingEval() {
	envSavedDataCache := os.Getenv("envSavedDataCache")

	if envSavedDataCache != "" {
		fmt.Println("Running testing eval:", envSavedDataCache)
		envSavedDataCachePort := envSavedDataCache + ":4123"

		dataConfig, _ := parseJSONFromURL("http://" + envSavedDataCachePort + "/files/config.json")

		//http://192.168.0.228:4123/mounted

		//map[setProjectPath:/path/p1]
		fmt.Println(dataConfig)
	}
}
