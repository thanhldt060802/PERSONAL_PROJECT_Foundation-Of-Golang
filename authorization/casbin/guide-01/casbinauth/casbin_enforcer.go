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
	GetPoliciesOfGroup(ctx context.Context, groupId string) (*[]Policy, error)
	GetPoliciesOfDomain(ctx context.Context, domainId string) (*[]Policy, error)
	AddPoliciesToGroup(ctx context.Context, groupId string, policies *[]Policy) error
	UpdatePoliciesForGroup(ctx context.Context, groupId string, policies *[]Policy) error
	RemovePoliciesFromGroup(ctx context.Context, groupId string) error
	RemovePoliciesFromDomain(ctx context.Context, domainId string) error

	GetGroupingPoliciesOfGroup(ctx context.Context, groupId string) (*[]GroupingPolicy, error)
	GetGroupingPoliciesOfDomain(ctx context.Context, domainId string) (*[]GroupingPolicy, error)
	AddGroupingPolicyToGroup(ctx context.Context, groupId string, groupingPolicy *GroupingPolicy) error
	AddGroupingPoliciesToGroup(ctx context.Context, groupId string, groupingPolicies *[]GroupingPolicy) error
	RemoveGroupingPolicyFromGroup(ctx context.Context, groupId string, subjectId string) error
	RemoveGroupingPoliciesFromGroup(ctx context.Context, groupId string) error
	RemoveGroupingPoliciesFromDomain(ctx context.Context, domainId string) error

	Enforce(ctx context.Context, request Request) (bool, error)

	Save(ctx context.Context) error
}

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(configFile string, db *gorm.DB) ICasbinEnforcer {
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(db, &CustomCasbinRule{})
	if err != nil {
		log.Fatalf("Failed to create Casbin adapter: %v", err.Error())
	}

	enforcer, err := casbin.NewEnforcer(configFile, adapter)
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

func (casbinEnf *CasbinEnforcer) AddPoliciesToGroup(ctx context.Context, groupId string, policies *[]Policy) error {
	for _, policy := range *policies {
		if _, err := casbinEnf.enforcer.AddPolicy(groupId, policy.Domain, policy.Object, policy.Action, policy.Condition); err != nil {
			return err
		}
	}
	return nil
}

func (casbinEnf *CasbinEnforcer) UpdatePoliciesForGroup(ctx context.Context, groupId string, policies *[]Policy) error {
	if err := casbinEnf.RemovePoliciesFromGroup(ctx, groupId); err != nil {
		return err
	}
	if err := casbinEnf.AddPoliciesToGroup(ctx, groupId, policies); err != nil {
		return err
	}
	return nil
}

func (casbinEnf *CasbinEnforcer) RemovePoliciesFromGroup(ctx context.Context, groupId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredPolicy(0, groupId)
	return err
}

func (casbinEnf *CasbinEnforcer) AddGroupingPolicyToGroup(ctx context.Context, groupId string, groupingPolicy *GroupingPolicy) error {
	_, err := casbinEnf.enforcer.AddGroupingPolicy(groupingPolicy.Subject, groupId, groupingPolicy.Domain)
	return err
}

func (casbinEnf *CasbinEnforcer) AddGroupingPoliciesToGroup(ctx context.Context, groupId string, groupingPolicies *[]GroupingPolicy) error {
	for _, groupingPolicy := range *groupingPolicies {
		if _, err := casbinEnf.enforcer.AddGroupingPolicy(groupingPolicy.Subject, groupId, groupingPolicy.Domain); err != nil {
			return err
		}
	}
	return nil
}

func (casbinEnf *CasbinEnforcer) RemoveGroupingPolicyFromGroup(ctx context.Context, groupId string, subjectId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredGroupingPolicy(0, subjectId, groupId)
	return err
}

func (casbinEnf *CasbinEnforcer) RemoveGroupingPoliciesFromGroup(ctx context.Context, groupId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredGroupingPolicy(1, groupId)
	return err
}

func (casbinEnf *CasbinEnforcer) Save(ctx context.Context) error {
	return casbinEnf.enforcer.SavePolicy()
}
