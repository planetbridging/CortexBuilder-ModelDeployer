package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/data/:index/:uuid", func(c *fiber.Ctx) error {
		return handleGetData(c)
	})

	app.Post("/data/add", func(c *fiber.Ctx) error {
		return handleAddData(c)
	})

	app.Get("/status", func(c *fiber.Ctx) error {
		return handleStatus(c)
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				break
			}

			var result = make(map[string]interface{})
			result["ip"] = c.RemoteAddr().String()

			var m map[string]string
			err = json.Unmarshal(msg, &m)
			if err != nil {
				fmt.Println("Invalid JSON:", err)
				break
			}

			var jsonData []byte

			switch m["action"] {
			case "get":
				index, _ := strconv.Atoi(m["index"])
				uuid := m["uuid"]
				data, err := getData(index, uuid)
				if err != nil {
					result["error"] = err.Error()
				} else {
					result["data"] = data
				}
				jsonData, _ = json.Marshal(result)

			case "add":
				index, _ := strconv.Atoi(m["index"])
				uuid := m["uuid"]
				url := m["url"]
				err := addData(index, uuid, url)
				if err != nil {
					result["error"] = err.Error()
				} else {
					result["message"] = "Data added successfully"
				}
				jsonData, _ = json.Marshal(result)

			case "status":
				status := getStatus()
				jsonData, _ = json.Marshal(status)

			default:
				result["error"] = "Unknown action"
				jsonData, _ = json.Marshal(result)
			}

			c.WriteMessage(websocket.TextMessage, jsonData)
		}
	}))
}

func handleGetData(c *fiber.Ctx) error {
	indexParam := c.Params("index")
	uuid := c.Params("uuid")

	index, err := strconv.Atoi(indexParam)
	if err != nil || index < 0 || index >= len(dataArray) {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid index")
	}

	data, err := getData(index, uuid)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString(err.Error())
	}

	return c.JSON(data)
}

func handleAddData(c *fiber.Ctx) error {
	indexParam := c.Query("index")
	uuid := c.Query("uuid")
	url := c.Query("url")

	if indexParam == "" || uuid == "" || url == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Index, UUID, and URL parameters are required")
	}

	index, err := strconv.Atoi(indexParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid index parameter")
	}

	err = addData(index, uuid, url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("Data added successfully")
}

func handleStatus(c *fiber.Ctx) error {
	status := getStatus()
	return c.JSON(status)
}

func getData(index int, uuid string) (interface{}, error) {
	if index < 0 || index >= len(dataArray) {
		return nil, fmt.Errorf("invalid index")
	}

	dataMap := dataArray[index]
	value, exists := dataMap[uuid]
	if !exists {
		return nil, fmt.Errorf("UUID not found in map")
	}

	return value, nil
}

func addData(index int, uuid, url string) error {
	data, err := fetchData(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %v", err)
	}

	// Ensure dataArray has enough capacity
	for len(dataArray) <= index {
		dataArray = append(dataArray, nil)
	}

	if dataArray[index] == nil {
		dataArray[index] = make(map[string]interface{})
	}

	dataArray[index][uuid] = data
	return nil
}

func getStatus() map[string]int {
	status := make(map[string]int)
	status["total_arrays"] = len(dataArray)
	for i, dataMap := range dataArray {
		status["array_"+strconv.Itoa(i)] = len(dataMap)
	}
	return status
}
