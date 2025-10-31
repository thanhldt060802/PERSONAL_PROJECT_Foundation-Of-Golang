package main

import (
	"context"
	"encoding/json"
	"thanhldt060802/casbinauth"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=12345678 dbname=casbin_demo port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	casbinauth.CasbinEnforcerInstance = casbinauth.NewCasbinEnforcer("config/hybrid_model.conf", db)

	// Add role
	policies := []casbinauth.Policy{
		{Domain: "domain_a", Object: "user", Action: "view", Condition: mapToString(map[string]any{
			"or": map[string]any{
				"team_id":          "team_1",
				"department_id_in": []string{"deparment_1", "deparment_2"},
				"and": map[string]any{
					"department_id":   "department_3",
					"department_type": "sale",
				},
			},
		})},
		{Domain: "domain_a", Object: "user", Action: "create", Condition: mapToString(map[string]any{
			"or": map[string]any{
				"team_id_id":    []string{"team_1", "team_2"},
				"department_id": "department_2",
			},
		})},
	}
	casbinauth.CasbinEnforcerInstance.AddPoliciesToGroup(context.Background(), "role_admin", &policies)
}

func mapToString(conditionMap map[string]any) string {
	b, err := json.Marshal(conditionMap)
	if err != nil {
		log.Errorf("Failed to marshal conditionMap: %v", err.Error())
		return ""
	}
	return string(b)
}
