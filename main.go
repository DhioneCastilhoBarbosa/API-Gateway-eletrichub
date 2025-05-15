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
	r.Any("/user/login", proxyRequest("https://api-user-service.eletrihub.com/user/login", false))
	r.Any("/user/register", proxyRequest("https://api-user-service.eletrihub.com/user/register", false))
	r.Any("/user/list", proxyRequest("https://api-user-service.eletrihub.com/user/list", false))
	r.Any("/user/public/installers", proxyRequest("https://api-user-service.eletrihub.com/user/public/installers", false))

	// Rotas dinâmicas (preserve o path original)
	r.Any("/user/:id/password", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id/photo", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("user/:id", proxyRequest("https://api-user-service.eletrihub.com", true))

	// WebSocket
	r.Any("/chat/ws", proxyRequest("http://localhost:8081/ws", false))

	log.Println("API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

// Função para redirecionar requisições para os microsserviços
func proxyRequest(targetURL string, preservePath bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := &http.Client{}

		// Decide se mantém o caminho original
		reqURL := targetURL
		if preservePath {
			reqURL += c.Request.URL.Path
		}

		if c.Request.URL.RawQuery != "" {
			reqURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequest(c.Request.Method, reqURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar requisição"})
			return
		}

		req.Header = make(http.Header)
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Erro ao conectar ao serviço"})
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler resposta do serviço"})
			return
		}

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
