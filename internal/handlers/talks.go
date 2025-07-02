package handlers

import (
	"api-gateway/internal/services"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetConversasOrcamentos(c *gin.Context) {
	token := c.GetHeader("Authorization")
	id := c.Query("id")
	tipo := c.Query("tipo") // "cliente" ou "instalador"

	//log.Println("🔍 Header Authorization recebido:", token)

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token não fornecido"})
		return
	}

	if id == "" || tipo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros 'id' e 'tipo' são obrigatórios"})
		return
	}

	// 🔄 Busca todos os orçamentos
	budgets, err := services.FetchBudgets(token)
	if err != nil {
		log.Println("❌ Erro ao buscar orçamentos:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar orçamentos"})
		return
	}

	// 🔍 Filtra de acordo com o tipo
	var filteredBudgets []services.Budget
	for _, b := range budgets {
		status := strings.ToLower(strings.TrimSpace(b.Status))
		if status == "concluido" || status == "cancelado" {
			continue
		}

		if (tipo == "cliente" && b.UserID == id) || (tipo == "instalador" && b.InstallerID == id) {
			filteredBudgets = append(filteredBudgets, b)
		}
	}

	//log.Printf("📊 Orçamentos encontrados para %s ID %s: %d\n", tipo, id, len(filteredBudgets))

	if len(filteredBudgets) == 0 {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	// 👥 Busca dados dos usuários
	ids := services.ExtractUserIDs(filteredBudgets)
	users, err := services.FetchUsersBatch(ids, token)
	if err != nil {
		log.Println("❌ Erro ao buscar usuários:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
		return
	}

	// 🧩 Monta a resposta
	var result []gin.H
	for _, b := range filteredBudgets {
		result = append(result, gin.H{
			"budget_id":  b.ID,
			"status":     b.Status,
			"cliente":    users[b.UserID],
			"instalador": users[b.InstallerID],
		})
	}

	c.JSON(http.StatusOK, result)
}
