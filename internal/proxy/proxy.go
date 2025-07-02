package proxy

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
