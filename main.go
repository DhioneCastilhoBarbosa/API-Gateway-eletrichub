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
	r.Any("/user/:id/password", proxyRequest("https://api-user-service.eletrihub.com"))
	r.Any("/user/list", proxyRequest("https://api-user-service.eletrihub.com/user/list"))
	r.Any("/chat/ws", proxyRequest("http://localhost:8081/ws"))
	r.Any("/user/:id/photo", proxyRequest("https://api-user-service.eletrihub.com"))

	log.Println("API Gateway rodando na porta 8086...")
	r.Run(":8086")
}

// Função para redirecionar requisições para os microsserviços
func proxyRequest(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := &http.Client{}

		// Constrói a URL completa (inclui path e query string)
		reqURL := targetURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			reqURL += "?" + c.Request.URL.RawQuery
		}

		// Cria nova requisição com o mesmo método e corpo
		req, err := http.NewRequest(c.Request.Method, reqURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar requisição"})
			return
		}

		// Copia todos os headers (necessário para multipart/form-data e Authorization)
		req.Header = make(http.Header)
		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// Faz a requisição para o serviço de destino
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Erro ao conectar ao serviço"})
			return
		}
		defer resp.Body.Close()

		// Copia o corpo da resposta para o cliente original
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler resposta do serviço"})
			return
		}

		// Encaminha status, headers e corpo da resposta
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
