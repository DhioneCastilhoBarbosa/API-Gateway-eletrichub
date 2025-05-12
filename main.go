package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configuração do CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Você pode restringir, ex: http://localhost:3000
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rotas do API Gateway
	r.Any("/user/login", proxyRequest("https://api-user-service.eletrihub.com/user/login"))
	r.Any("/user/register", proxyRequest("https://api-user-service.eletrihub.com/user/register"))
	r.Any("/user/list", proxyRequest("https://api-user-service.eletrihub.com/user/list"))
	r.Any("/chat/ws", proxyRequest("http://localhost:8081/ws"))

	log.Println("API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

// Função para redirecionar requisições para os microsserviços
func proxyRequest(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := &http.Client{}

		// Adiciona a query string original
		reqURL := targetURL
		if c.Request.URL.RawQuery != "" {
			reqURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequest(c.Request.Method, reqURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar requisição"})
			return
		}

		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao conectar ao serviço"})
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}
