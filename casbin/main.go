package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	// testSetupRole()
	// testPrintRole()
}

func testSetupRole() {
	// SETUP FOR DOMAIN_1

	// Add domain_1_role_1
	{
		policies := []casbinauth.Policy{
			{
				SubjectGroup: "domain_1_role_1",
				Domain:       "domain_1",
				Object:       "user",
				Action:       "view",
				Condition:    "*",
			},
			{
				SubjectGroup: "domain_1_role_1",
				Domain:       "domain_1",
				Object:       "user",
				Action:       "create",
				Condition: mapToString(map[string]any{
					"or": map[string]any{
						"team_id_in": []string{"domain_1_team_1", "domain_1_team_2"},
						"and": map[string]any{
							"team_id_eq":       "domain_1_team_3",
							"department_id_in": []string{"domain_1_department_1", "domain_1_department_2"},
						},
					},
				}),
			},
		}
		casbinauth.CasbinEnforcerInstance.AddPoliciesToGroup(context.Background(), &policies)
	}

	// Add domain_1_user_1 and domain_1_user_2 to domain_1_role_1
	{
		groupingPolicies := []casbinauth.GroupingPolicy{
			{
				Subject:      "domain_1_user_1",
				SubjectGroup: "domain_1_role_1",
				Domain:       "domain_1",
			},
			{
				Subject:      "domain_1_user_2",
				SubjectGroup: "domain_1_role_1",
				Domain:       "domain_1",
			},
		}
		casbinauth.CasbinEnforcerInstance.AddGroupingPoliciesToGroup(context.Background(), &groupingPolicies)

	}

	// Add domain_1_role_2
	{
		policies := []casbinauth.Policy{
			{
				SubjectGroup: "domain_1_role_2",
				Domain:       "domain_1",
				Object:       "user",
				Action:       "view",
				Condition: mapToString(map[string]any{
					"and": map[string]any{
						"user_id_eq": "owner_id",
					},
				}),
			},
			{
				SubjectGroup: "domain_1_role_2",
				Domain:       "domain_1",
				Object:       "user",
				Action:       "update",
				Condition: mapToString(map[string]any{
					"and": map[string]any{
						"user_id_eq": "owner_id",
					},
				}),
			},
		}
		casbinauth.CasbinEnforcerInstance.AddPoliciesToGroup(context.Background(), &policies)
	}

	// Add domain_1_user_3 and domain_1_user_4 to domain_1_role_2
	{
		groupingPolicies := []casbinauth.GroupingPolicy{
			{
				Subject:      "domain_1_user_3",
				SubjectGroup: "domain_1_role_2",
				Domain:       "domain_1",
			},
			{
				Subject:      "domain_1_user_4",
				SubjectGroup: "domain_1_role_2",
				Domain:       "domain_1",
			},
		}
		casbinauth.CasbinEnforcerInstance.AddGroupingPoliciesToGroup(context.Background(), &groupingPolicies)
	}

	// SETUP FOR DOMAIN_2 SAME DOMAIN_1

	// Add domain_2_role_1
	{
		policies := []casbinauth.Policy{
			{
				SubjectGroup: "domain_2_role_1",
				Domain:       "domain_2",
				Object:       "user",
				Action:       "view",
				Condition:    "*",
			},
			{
				SubjectGroup: "domain_2_role_1",
				Domain:       "domain_2",
				Object:       "user",
				Action:       "create",
				Condition: mapToString(map[string]any{
					"or": map[string]any{
						"team_id_in": []string{"domain_2_team_1", "domain_2_team_2"},
						"and": map[string]any{
							"team_id_eq":       "domain_2_team_3",
							"department_id_in": []string{"domain_2_department_1", "domain_2_department_2"},
						},
					},
				}),
			},
		}
		casbinauth.CasbinEnforcerInstance.AddPoliciesToGroup(context.Background(), &policies)
	}

	// Add domain_2_user_1 and domain_2_user_2 to domain_2_role_1
	{
		groupingPolicies := []casbinauth.GroupingPolicy{
			{
				Subject:      "domain_2_user_1",
				SubjectGroup: "domain_2_role_1",
				Domain:       "domain_2",
			},
			{
				Subject:      "domain_2_user_2",
				SubjectGroup: "domain_2_role_1",
				Domain:       "domain_2",
			},
		}
		casbinauth.CasbinEnforcerInstance.AddGroupingPoliciesToGroup(context.Background(), &groupingPolicies)
	}

	// Add domain_2_role_2
	{
		policies := []casbinauth.Policy{
			{
				SubjectGroup: "domain_2_role_2",
				Domain:       "domain_2",
				Object:       "user",
				Action:       "view",
				Condition: mapToString(map[string]any{
					"and": map[string]any{
						"user_id_eq": "owner_id",
					},
				}),
			},
			{
				SubjectGroup: "domain_2_role_2",
				Domain:       "domain_2",
				Object:       "user",
				Action:       "update",
				Condition: mapToString(map[string]any{
					"and": map[string]any{
						"user_id_eq": "owner_id",
					},
				}),
			},
		}
		casbinauth.CasbinEnforcerInstance.AddPoliciesToGroup(context.Background(), &policies)
	}

	// Add domain_2_user_3 and domain_2_user_4 to domain_2_role_2
	{
		groupingPolicies := []casbinauth.GroupingPolicy{
			{
				Subject:      "domain_2_user_3",
				SubjectGroup: "domain_2_role_2",
				Domain:       "domain_2",
			},
			{
				Subject:      "domain_2_user_4",
				SubjectGroup: "domain_2_role_2",
				Domain:       "domain_2",
			},
		}
		casbinauth.CasbinEnforcerInstance.AddGroupingPoliciesToGroup(context.Background(), &groupingPolicies)
	}

	casbinauth.CasbinEnforcerInstance.Save(context.Background())
}

func testPrintRole() {
	// VISUALIZE FOR DOMAIN_1
	fmt.Println("DOMAIN_1 ------------------------------------------------")

	// Print all policies of domain_1
	fmt.Println("All policies of domain_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfDomain(context.Background(), "domain_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all policies of domain_1_role_1
	fmt.Println("All policies of domain_1_role_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfGroup(context.Background(), "domain_1_role_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all policies of domain_1_role_2
	fmt.Println("All policies of domain_1_role_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfGroup(context.Background(), "domain_1_role_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_1
	fmt.Println("All grouping policies of domain_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfDomain(context.Background(), "domain_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_1_role_1
	fmt.Println("All grouping policies of domain_1_role_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfGroup(context.Background(), "domain_1_role_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_1_role_2
	fmt.Println("All grouping policies of domain_1_role_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfGroup(context.Background(), "domain_1_role_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// VISUALIZE FOR DOMAIN_2
	fmt.Println("DOMAIN_2 ------------------------------------------------")

	// Print all policies of domain_2
	fmt.Println("All policies of domain_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfDomain(context.Background(), "domain_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all policies of domain_2_role_1
	fmt.Println("All policies of domain_2_role_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfGroup(context.Background(), "domain_2_role_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all policies of domain_2_role_2
	fmt.Println("All policies of domain_2_role_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetPoliciesOfGroup(context.Background(), "domain_2_role_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_2
	fmt.Println("All grouping policies of domain_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfDomain(context.Background(), "domain_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_2_role_1
	fmt.Println("All grouping policies of domain_2_role_1")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfGroup(context.Background(), "domain_2_role_1")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}

	// Print all grouping policies of domain_2_role_2
	fmt.Println("All grouping policies of domain_2_role_2")
	{
		policies, _ := casbinauth.CasbinEnforcerInstance.GetGroupingPoliciesOfGroup(context.Background(), "domain_2_role_2")
		for _, policy := range *policies {
			fmt.Println(policy)
		}
	}
}

func testEnforce() {
	// ENFORCE FOR DOMAIN_1
	fmt.Println(casbinauth.CasbinEnforcerInstance.Enforce(context.Background(), casbinauth.Request{
		Subject: "domain_1_user_1",
		Domain:  "domain_1",
		Object:  "user",
		Action:  "view",
		CtxCondition: map[string]string{
			"team_id": "domain_1_team_5",
		},
	})) // 1
	fmt.Println(casbinauth.CasbinEnforcerInstance.Enforce(context.Background(), casbinauth.Request{
		Subject: "domain_1_user_1",
		Domain:  "domain_1",
		Object:  "user",
		Action:  "view",
		CtxCondition: map[string]string{
			"team_id":       "domain_1_team_4",
			"department_id": "domain_1_department_3",
		},
	})) // 2
	fmt.Println(casbinauth.CasbinEnforcerInstance.Enforce(context.Background(), casbinauth.Request{
		Subject: "domain_1_user_1",
		Domain:  "domain_1",
		Object:  "user",
		Action:  "create",
		CtxCondition: map[string]string{
			"team_id":       "domain_1_team_1",
			"department_id": "domain_1_department_1",
		},
	})) // 3
	fmt.Println(casbinauth.CasbinEnforcerInstance.Enforce(context.Background(), casbinauth.Request{
		Subject: "domain_1_user_1",
		Domain:  "domain_1",
		Object:  "user",
		Action:  "create",
		CtxCondition: map[string]string{
			"team_id":       "domain_1_team_3",
			"department_id": "domain_1_department_1",
		},
	})) // 4
}

func mapToString(conditionMap map[string]any) string {
	b, err := json.Marshal(conditionMap)
	if err != nil {
		log.Errorf("Failed to marshal conditionMap: %v", err.Error())
		return ""
	}
	return string(b)
}
