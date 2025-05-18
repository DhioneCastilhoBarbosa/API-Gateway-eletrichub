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
		AllowOrigins:     []string{"*"}, // Restrinja para produção
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rotas simples (endpoint fixo)
	r.Any("/user/login", proxyRequest("https://api-user-service.eletrihub.com/user/login", false))
	r.Any("/user/register", proxyRequest("https://api-user-service.eletrihub.com/user/register", false))
	r.Any("/user/list", proxyRequest("https://api-user-service.eletrihub.com/user/list", false))
	r.Any("/user/public/installers", proxyRequest("https://api-user-service.eletrihub.com/user/public/installers", false))

	// Rota com query params (lat/lng) - ajustada
	r.Any("/user/public/installers/nearby", proxyRequest("https://api-user-service.eletrihub.com/user/public/installers/nearby", false))

	// Rotas dinâmicas (com ID no path)
	r.Any("/user/:id/password", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id/photo", proxyRequest("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id", proxyRequest("https://api-user-service.eletrihub.com", true))

	// Rotas para o budget-service
	r.Any("/api/v1/budget", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget", false))
	r.Any("/api/v1/budget/", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget", false))          // POST, GET
	r.Any("/api/v1/budget/all", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget/all", false))   // GET
	r.Any("/api/v1/budget/link", proxyRequest("https://budget-service.api-castilho.com.br/api/v1/budget/link", false)) // PUT

	r.Any("/api/v1/budget/:id/value", proxyRequest("https://budget-service.api-castilho.com.br", true))   // PUT (protegido)
	r.Any("/api/v1/budget/:id/status", proxyRequest("https://budget-service.api-castilho.com.br", true))  // PUT (protegido)
	r.Any("/api/v1/budget/:id/dates", proxyRequest("https://budget-service.api-castilho.com.br", true))   // PUT (protegido)
	r.Any("/api/v1/budget/:id/payment", proxyRequest("https://budget-service.api-castilho.com.br", true)) // PUT (protegido)
	r.Any("/api/v1/budget/:id/confirm", proxyRequest("https://budget-service.api-castilho.com.br", true)) // PUT (protegido)

	// WebSocket
	r.Any("/chat/ws", proxyRequest("http://localhost:8081/ws", false))

	log.Println("API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

// Função para redirecionar requisições para os microsserviços
func proxyRequest(targetURL string, preservePath bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := &http.Client{}

		// Monta a URL de destino
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

		// Copia os headers
		req.Header = make(http.Header)
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Executa a requisição
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

		// Copia o corpo da resposta
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler resposta do serviço"})
			return
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
