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

	// CORS configurado corretamente
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://app.eletrihub.com"}, // ajuste para produção
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rotas simples
	r.Any("/user/login", proxyRequest("https://api-user-service.eletrihub.com/user/login", false))
	r.Any("/user/register", proxyRequest("https://api-user-service.eletrihub.com/user/register", false))
	r.Any("/user/list", proxyRequest("https://api-user-service.eletrihub.com/user/list", false))
	r.Any("/user/public/installers", proxyRequest("https://api-user-service.eletrihub.com/user/public/installers", false))
	r.Any("/user/public/installers/nearby", proxyRequest("https://api-user-service.eletrihub.com/user/public/installers/nearby", false))

	// Dinâmicas
	r.Any("/user/:id/password", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id/photo", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id", proxyRequest("https://api-user-service.eletrihub.com", true))

	// Rotas budget-service
	r.Any("/api/v1/budget/", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget/", false))
	r.Any("/api/v1/budget/all", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget/all", false))
	r.Any("/api/v1/budget/link", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget/link", false))

	r.Any("/api/v1/budget/:id/value", proxyRequest("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/status", proxyRequest("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/dates", proxyRequest("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/payment", proxyRequest("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/confirm", proxyRequest("https://budget-service.api-castilho.com.br", true))

	// WebSocket
	r.Any("/chat/ws", proxyRequest("http://localhost:8081/ws", false))

	log.Println("✅ API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

func proxyRequest(targetURL string, preservePath bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ✅ Trata preflight diretamente
		if c.Request.Method == http.MethodOptions {
			origin := c.GetHeader("Origin")
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Status(http.StatusOK)
			return
		}

		client := &http.Client{}
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

		// Copia os headers originais
		req.Header = make(http.Header)
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Faz a requisição
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Erro ao conectar ao serviço"})
			return
		}
		defer resp.Body.Close()

		// Copia headers da resposta
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler resposta do serviço"})
			return
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
