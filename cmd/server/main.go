package main

import (
	"api-gateway/internal/handlers"
	"api-gateway/internal/proxy"
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
		AllowOrigins:     []string{"http://localhost:5173", "https://www.eletrihub.com"}, // ajuste para produção
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Rota agregadora
	r.GET("/api/conversas-orcamentos", handlers.GetConversasOrcamentos)

	// Rotas simples
	r.Any("/user/login", proxy.Request("https://api-user-service.eletrihub.com/user/login", false))
	r.Any("/user/register", proxy.Request("https://api-user-service.eletrihub.com/user/register", false))
	r.Any("/user/list", proxy.Request("https://api-user-service.eletrihub.com/user/list", false))
	r.Any("/user/public/installers", proxy.Request("https://api-user-service.eletrihub.com/user/public/installers", false))
	r.Any("/user/public/installers/nearby", proxy.Request("https://api-user-service.eletrihub.com/user/public/installers/nearby", false))

	// Dinâmicas
	r.Any("/user/:id/password", proxy.Request("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id/photo", proxy.Request("https://api-user-service.eletrihub.com", true))
	r.Any("/user/:id", proxy.Request("https://api-user-service.eletrihub.com", true))

	// Rotas budget-service
	r.Any("/api/v1/budget/", proxy.Request("https://budget-service.api-castilho.com.br/api/v1/budget/", false))
	r.Any("/api/v1/budget/all", proxy.Request("https://budget-service.api-castilho.com.br/api/v1/budget/all", false))
	r.Any("/api/v1/budget/link", proxy.Request("https://budget-service.api-castilho.com.br/api/v1/budget/link", false))
	r.Any("/api/v1/budget/pagamento/webhook", proxy.Request("https://budget-service.api-castilho.com.br/api/v1/budget/pagamento/webhook", false))

	r.Any("/api/v1/budget/:id/value", proxy.Request("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/status", proxy.Request("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/dates", proxy.Request("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/payment", proxy.Request("https://budget-service.api-castilho.com.br", true))
	r.Any("/api/v1/budget/:id/confirm", proxy.Request("https://budget-service.api-castilho.com.br", true))

	// Rotas do payment-service
	r.Any("/criar-pagamento", proxy.Request("https://api-payments.eletrihub.com/criar-pagamento", false))
	r.Any("/webhook-mercado-pago", proxy.Request("https://api-payments.eletrihub.com/webhook-mercado-pago", false))

	// Rotas do notification-service
	r.Any("/notificar-cliente", proxy.Request("https://api-notification.eletrihub.com/notificar-cliente", false))
	r.Any("/notificar-instalador", proxy.Request("https://api-notification.eletrihub.com/notificar-instalador", false))

	r.Any("/chat/history", proxy.Request("https://api-chat-service.eletrihub.com/chat-history", false))

	log.Println("✅ API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

func Request(targetURL string, preservePath bool) gin.HandlerFunc {
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
