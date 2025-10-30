package casbinauth

import (
	"context"
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var CasbinEnforcerInstance ICasbinEnforcer

type ICasbinEnforcer interface {
	AddGroupingPolicy(ctx context.Context, groupingPolicy *GroupingPolicy) error
	AddGroupingPolicies(ctx context.Context, groupingPolicy *[]GroupingPolicy) error
	RemoveGroupingPolicy(ctx context.Context, groupingPolicyId string, domainId string) error
	RemoveGroupingPolicyByUser(ctx context.Context, groupingPolicyId string, userId string) error
	RemoveGroupingPolicyOfDomain(ctx context.Context, domainId string) error

	AddPolicies(ctx context.Context, policy *[]Policy) error
	RemovePolicyOfGroupingPolicy(ctx context.Context, groupingPolicyId string) error
	RemovePolicyOfDomain(ctx context.Context, domainId string) error

	Save(ctx context.Context) error
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(db *gorm.DB) ICasbinEnforcer {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("Failed to create Casbin adapter: %v", err.Error())
	}

	enforcer, err := casbin.NewEnforcer("model.conf", adapter)
	if err != nil {
		log.Fatalf("Failed to create Enforcer: %v", err.Error())
	}

	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("Failed to load Policy for Enforcer: %v", err.Error())
	}

	return &CasbinEnforcer{
		enforcer: enforcer,
	}
}

// AddGroupingPolicies implements ICasbinEnforcer.
func (c *CasbinEnforcer) AddGroupingPolicies(ctx context.Context, groupingPolicy *[]GroupingPolicy) error {
	panic("unimplemented")
}

// AddGroupingPolicy implements ICasbinEnforcer.
func (c *CasbinEnforcer) AddGroupingPolicy(ctx context.Context, groupingPolicy *GroupingPolicy) error {
	panic("unimplemented")
}

// AddPolicies implements ICasbinEnforcer.
func (c *CasbinEnforcer) AddPolicies(ctx context.Context, policy *[]Policy) error {
	panic("unimplemented")
}

// RemoveGroupingPolicy implements ICasbinEnforcer.
func (c *CasbinEnforcer) RemoveGroupingPolicy(ctx context.Context, groupingPolicyId string, domainId string) error {
	panic("unimplemented")
}

// RemoveGroupingPolicyByUser implements ICasbinEnforcer.
func (c *CasbinEnforcer) RemoveGroupingPolicyByUser(ctx context.Context, groupingPolicyId string, userId string) error {
	panic("unimplemented")
}

// RemoveGroupingPolicyOfDomain implements ICasbinEnforcer.
func (c *CasbinEnforcer) RemoveGroupingPolicyOfDomain(ctx context.Context, domainId string) error {
	panic("unimplemented")
}

// RemovePolicyOfDomain implements ICasbinEnforcer.
func (c *CasbinEnforcer) RemovePolicyOfDomain(ctx context.Context, domainId string) error {
	panic("unimplemented")
}

// RemovePolicyOfGroupingPolicy implements ICasbinEnforcer.
func (c *CasbinEnforcer) RemovePolicyOfGroupingPolicy(ctx context.Context, groupingPolicyId string) error {
	panic("unimplemented")
}

// Save implements ICasbinEnforcer.
func (c *CasbinEnforcer) Save(ctx context.Context) error {
	panic("unimplemented")
}
