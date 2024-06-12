package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

var dataArray []map[string]NetworkConfig
var envPWD string
var envSavedDataCache string

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	envPort := os.Getenv("PORT")

	if envPort == "" {
		envPort = "4125"
	}

	//testingEval()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Allows all domains
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type",
		AllowCredentials: false,
	}))

	app.Use(customCORSHandler)

	setupRoutes(app)

	envPWD = os.Getenv("SERVERPWD")
	if envPWD == "" {
		envPWD = "securepassword"
		//os.Setenv("FILE_PATH", envPath)
	}

	go startTcpServer(envPWD)

	err = app.Listen(":" + envPort)
	if err != nil {
		fmt.Println(err)
	}
}

func customCORSHandler(c *fiber.Ctx) error {
	c.Set("Access-Control-Allow-Origin", c.Get("Origin"))
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

	if c.Method() == "OPTIONS" {
		return c.SendStatus(fiber.StatusOK)
	}

	return c.Next()
}

func fetchData(url string) (NetworkConfig, error) {
	resp, err := http.Get(url)
	if err != nil {
		return NetworkConfig{}, err
	}
	defer resp.Body.Close()

	var data NetworkConfig
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return NetworkConfig{}, err
	}

	return data, nil
}
