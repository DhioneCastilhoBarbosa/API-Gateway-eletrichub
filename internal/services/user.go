package services

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type UserInfo struct {
	ID    string `json:"id"`
	Name  string `json:"username"` // <- era "nome", mas o JSON retorna "username"
	Photo string `json:"photo"`    // <- certinho
}

func FetchUsersBatch(ids []string, token string) (map[string]UserInfo, error) {
	userMap := make(map[string]UserInfo)

	for _, id := range ids {
		url := "https://api.eletrihub.com/user/list?id=" + id

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		log.Printf("📦 JSON do usuário [%s]: %s\n", id, string(body))

		var users []UserInfo
		err = json.Unmarshal(body, &users)
		if err != nil {
			log.Printf("❌ Erro ao decodificar usuário [%s]: %v", id, err)
			continue
		}

		if len(users) > 0 {
			userMap[id] = users[0]
		}
	}

	return userMap, nil
}
