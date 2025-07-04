package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type StartRequest struct {
	Project string `json:"project" binding:"required"`
	Image   string `json:"image" binding:"required"`
}

func main() {
	loginScript := flag.String("login-script", "", "Cesta k docker login scriptu (sh)")
	composeDir := flag.String("compose-dir", "", "Cesta k adresáři s docker compose soubory")
	port := flag.String("port", "8080", "Port na kterém poběží API server")
	flag.Parse()

	if *loginScript == "" || *composeDir == "" {
		fmt.Println("Použití: go run main.go --login-script <cesta_k_scriptu> --compose-dir <cesta_k_adresari> [--port <port>]")
		os.Exit(1)
	}

	r := gin.Default()

	r.POST("/start", func(c *gin.Context) {
		var req StartRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 1. Spustit login script
		cmd := exec.Command("sh", *loginScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login script selhal", "detail": err.Error()})
			return
		}

		// 2. Najít compose soubor
		composeFile := filepath.Join(*composeDir, req.Project+".yaml")
		if _, err := os.Stat(composeFile); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Compose soubor nenalezen", "file": composeFile})
			return
		}

		// 3. Nahradit image v compose souboru
		input, err := os.ReadFile(composeFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Nelze číst compose soubor", "detail": err.Error()})
			return
		}
		lines := strings.Split(string(input), "\n")
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "image:") {
				indent := line[:strings.Index(line, "image:")]
				lines[i] = indent + "image: " + req.Image
			}
		}
		output := strings.Join(lines, "\n")
		if err := os.WriteFile(composeFile, []byte(output), 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Nelze zapsat compose soubor", "detail": err.Error()})
			return
		}

		// 4. Spustit docker compose up -d --force-recreate
		composeCmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d", "--force-recreate")
		composeCmd.Stdout = os.Stdout
		composeCmd.Stderr = os.Stderr
		if err := composeCmd.Run(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "docker compose selhal", "detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Docker compose úspěšně spuštěn",
			"project":      req.Project,
			"image":        req.Image,
			"compose_file": composeFile,
		})
	})

	fmt.Printf("API server running on :%s, ready for requests\n", *port)
	r.Run(":" + *port)
}