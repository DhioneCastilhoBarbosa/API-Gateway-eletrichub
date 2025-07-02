package services

import (
	"encoding/json"
	"io"
	"log"
	"strings"

	"net/http"
)

type Budget struct {
	ID          int    `json:"id"`
	UserID      string `json:"user_id"`
	InstallerID string `json:"installer_id"`
	Status      string `json:"status"`
}

func FetchBudgets(token string) ([]Budget, error) {
	url := "https://api.eletrihub.com/api/v1/budget/all"

	log.Println("🛡️ Token enviado:", token)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("❌ Erro ao fazer requisição:", err)
		return nil, err
	}
	defer resp.Body.Close()

	//log.Println("🔄 Status da resposta:", resp.Status)

	body, _ := io.ReadAll(resp.Body)
	//log.Println("📦 Corpo da resposta:", string(body))

	var budgets []Budget
	err = json.Unmarshal(body, &budgets)
	if err != nil {
		log.Println("❌ Erro ao decodificar JSON:", err)
		return nil, err
	}

	return budgets, nil
}

func ExtractUserIDs(budgets []Budget) []string {
	idMap := make(map[string]bool)
	for _, b := range budgets {
		if strings.TrimSpace(b.UserID) != "" {
			idMap[b.UserID] = true
		}
		if strings.TrimSpace(b.InstallerID) != "" {
			idMap[b.InstallerID] = true
		}
	}
	var ids []string
	for id := range idMap {
		ids = append(ids, id)
	}
	return ids
}
