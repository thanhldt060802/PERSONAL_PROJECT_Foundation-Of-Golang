package main

import (
	"fmt"
	"log"
	"os"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const modelConf = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _   # <-- đây nè, thay vì g = _, _, dom

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

func main() {
	//----------------------------------------------------------
	// ⚙️ 1️⃣ Database connection (PostgreSQL)
	//----------------------------------------------------------
	// Example DSN (local Postgres)
	// make sure your database exists, e.g. `createdb casbin_demo`
	// user: postgres, password: mypassword, dbname: casbin_demo
	//----------------------------------------------------------
	dsn := "host=localhost user=postgres password=12345678 dbname=casbin_demo port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ failed to connect to Postgres: %v", err)
	}

	// Create adapter (Casbin ↔ Gorm)
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("❌ failed to create Casbin adapter: %v", err)
	}

	//----------------------------------------------------------
	// 🧩 2️⃣ Setup Enforcer
	//----------------------------------------------------------
	_ = os.WriteFile("model.conf", []byte(modelConf), 0644)

	e, err := casbin.NewEnforcer("model.conf", adapter)
	if err != nil {
		log.Fatalf("❌ failed to create Enforcer: %v", err)
	}

	_ = e.LoadPolicy()

	fmt.Println("🏗️  --- Casbin Authorization Demo (Multi-tenant RBAC with PostgreSQL) ---")

	//----------------------------------------------------------
	// 3️⃣ Add Policies (role → permission)
	//----------------------------------------------------------
	fmt.Println("\n👉 Add policies (role permissions):")

	e.AddPolicy("admin", "tenantA", "project", "read")
	e.AddPolicy("admin", "tenantA", "project", "write")
	e.AddPolicy("editor", "tenantA", "project", "read")
	e.AddPolicy("viewer", "tenantB", "dashboard", "view")

	e.SavePolicy()
	printPolicies(e)

	//----------------------------------------------------------
	// 4️⃣ Add Grouping (user → role)
	//----------------------------------------------------------
	fmt.Println("\n👉 Assign users to roles:")
	e.AddGroupingPolicy("alice", "admin", "tenantA")
	e.AddGroupingPolicy("bob", "editor", "tenantA")
	e.AddGroupingPolicy("charlie", "viewer", "tenantB")
	e.SavePolicy()
	printGroupings(e)

	//----------------------------------------------------------
	// 5️⃣ Enforce tests
	//----------------------------------------------------------
	fmt.Println("\n👉 Enforce checks:")
	testEnforce(e, "alice", "tenantA", "project", "write")    // ✅
	testEnforce(e, "bob", "tenantA", "project", "write")      // ❌
	testEnforce(e, "charlie", "tenantB", "dashboard", "view") // ✅
	testEnforce(e, "charlie", "tenantA", "project", "read")   // ❌
	testEnforce(e, "unknown", "tenantA", "project", "read")   // ❌

	//----------------------------------------------------------
	// 6️⃣ Update Policy
	//----------------------------------------------------------
	fmt.Println("\n👉 Update policy: (admin write → delete, add new)")
	e.RemovePolicy("admin", "tenantA", "project", "write")
	e.AddPolicy("admin", "tenantA", "project", "delete")
	e.SavePolicy()
	printPolicies(e)

	testEnforce(e, "alice", "tenantA", "project", "write")  // ❌
	testEnforce(e, "alice", "tenantA", "project", "delete") // ✅

	//----------------------------------------------------------
	// 7️⃣ Remove Grouping
	//----------------------------------------------------------
	fmt.Println("\n👉 Remove user from role (bob ← editor)")
	e.RemoveGroupingPolicy("bob", "editor", "tenantA")
	e.SavePolicy()
	printGroupings(e)

	testEnforce(e, "bob", "tenantA", "project", "read") // ❌

	//----------------------------------------------------------
	// 8️⃣ Remove Filtered Policy
	//----------------------------------------------------------
	fmt.Println("\n👉 Remove all policies for tenantA:")
	e.RemoveFilteredPolicy(1, "tenantA")
	e.SavePolicy()
	printPolicies(e)

	testEnforce(e, "alice", "tenantA", "project", "delete")   // ❌
	testEnforce(e, "charlie", "tenantB", "dashboard", "view") // ✅ still ok

	//----------------------------------------------------------
	fmt.Println("\n✅ All done — check your PostgreSQL 'casbin_rule' table to inspect stored rules.")
}

// ------------------------------------------------------------
// Helpers
// ------------------------------------------------------------

func testEnforce(e *casbin.Enforcer, sub, dom, obj, act string) {
	ok, err := e.Enforce(sub, dom, obj, act)
	if err != nil {
		fmt.Printf("⚠️  Enforce error: %v\n", err)
		return
	}
	status := "❌ DENY"
	if ok {
		status = "✅ ALLOW"
	}
	fmt.Printf("  %-7s | dom=%-8s | obj=%-10s | act=%-6s => %s\n", sub, dom, obj, act, status)
}

func printPolicies(e *casbin.Enforcer) {
	policies, err := e.GetPolicy()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Current Policies:")
	for _, p := range policies {
		fmt.Printf("  p, %v\n", p)
	}
}

func printGroupings(e *casbin.Enforcer) {
	groupings, err := e.GetGroupingPolicy()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Current Groupings:")
	for _, g := range groupings {
		fmt.Printf("  g, %v\n", g)
	}
}
