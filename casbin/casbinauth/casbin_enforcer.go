package casbinauth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var CasbinEnforcerInstance ICasbinEnforcer

type ICasbinEnforcer interface {
	GetPoliciesOfGroup(ctx context.Context, groupId string) (*[]Policy, error)
	GetPoliciesOfDomain(ctx context.Context, domainId string) (*[]Policy, error)
	AddPoliciesToGroup(ctx context.Context, policies *[]Policy) error
	UpdatePoliciesForGroup(ctx context.Context, groupId string, policies *[]Policy) error
	RemovePoliciesFromGroup(ctx context.Context, groupId string) error
	RemovePoliciesFromDomain(ctx context.Context, domainId string) error

	GetGroupingPoliciesOfGroup(ctx context.Context, groupId string) (*[]GroupingPolicy, error)
	GetGroupingPoliciesOfDomain(ctx context.Context, domainId string) (*[]GroupingPolicy, error)
	AddGroupingPolicyToGroup(ctx context.Context, groupingPolicy *GroupingPolicy) error
	AddGroupingPoliciesToGroup(ctx context.Context, groupingPolicies *[]GroupingPolicy) error
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
	enforcer.EnableAutoSave(false)

	if err := enforcer.LoadPolicy(); err != nil {
		log.Fatalf("Failed to load Policy for Enforcer: %v", err.Error())
	}

	casbinEnf := &CasbinEnforcer{
		enforcer: enforcer,
	}
	casbinEnf.enforcer.AddFunction("inScope", casbinEnf.inScope)

	return casbinEnf
}

func (casbinEnf *CasbinEnforcer) GetPoliciesOfGroup(ctx context.Context, groupId string) (*[]Policy, error) {
	rawPolicies, err := casbinEnf.enforcer.GetFilteredPolicy(0, groupId)
	if err != nil {
		return nil, err
	}

	policies := make([]Policy, 0)
	for _, rawPolicy := range rawPolicies {
		policies = append(policies, Policy{
			SubjectGroup: rawPolicy[0],
			Domain:       rawPolicy[1],
			Object:       rawPolicy[2],
			Action:       rawPolicy[3],
			Condition:    rawPolicy[4],
		})
	}

	return &policies, nil
}

func (casbinEnf *CasbinEnforcer) GetPoliciesOfDomain(ctx context.Context, domainId string) (*[]Policy, error) {
	rawPolicies, err := casbinEnf.enforcer.GetFilteredPolicy(1, domainId)
	if err != nil {
		return nil, err
	}

	policies := make([]Policy, 0)
	for _, rawPolicy := range rawPolicies {
		policies = append(policies, Policy{
			SubjectGroup: rawPolicy[0],
			Domain:       rawPolicy[1],
			Object:       rawPolicy[2],
			Action:       rawPolicy[3],
			Condition:    rawPolicy[4],
		})
	}

	return &policies, nil
}

func (casbinEnf *CasbinEnforcer) AddPoliciesToGroup(ctx context.Context, policies *[]Policy) error {
	for _, policy := range *policies {
		if _, err := casbinEnf.enforcer.AddPolicy(policy.SubjectGroup, policy.Domain, policy.Object, policy.Action, policy.Condition); err != nil {
			return err
		}
	}
	return nil
}

func (casbinEnf *CasbinEnforcer) UpdatePoliciesForGroup(ctx context.Context, groupId string, policies *[]Policy) error {
	if err := casbinEnf.RemovePoliciesFromGroup(ctx, groupId); err != nil {
		return err
	}
	if err := casbinEnf.AddPoliciesToGroup(ctx, policies); err != nil {
		return err
	}
	return nil
}

func (casbinEnf *CasbinEnforcer) RemovePoliciesFromGroup(ctx context.Context, groupId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredPolicy(0, groupId)
	return err
}

func (casbinEnf *CasbinEnforcer) RemovePoliciesFromDomain(ctx context.Context, domainId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredPolicy(1, domainId)
	return err
}

func (casbinEnf *CasbinEnforcer) GetGroupingPoliciesOfGroup(ctx context.Context, groupId string) (*[]GroupingPolicy, error) {
	rawGroupingPolicies, err := casbinEnf.enforcer.GetFilteredGroupingPolicy(1, groupId)
	if err != nil {
		return nil, err
	}

	groupingPolicies := make([]GroupingPolicy, 0)
	for _, rawGroupingPolicy := range rawGroupingPolicies {
		groupingPolicies = append(groupingPolicies, GroupingPolicy{
			Subject:      rawGroupingPolicy[0],
			SubjectGroup: rawGroupingPolicy[1],
			Domain:       rawGroupingPolicy[2],
		})
	}

	return &groupingPolicies, nil
}

func (casbinEnf *CasbinEnforcer) GetGroupingPoliciesOfDomain(ctx context.Context, domainId string) (*[]GroupingPolicy, error) {
	rawGroupingPolicies, err := casbinEnf.enforcer.GetFilteredGroupingPolicy(2, domainId)
	if err != nil {
		return nil, err
	}

	groupingPolicies := make([]GroupingPolicy, 0)
	for _, rawGroupingPolicy := range rawGroupingPolicies {
		groupingPolicies = append(groupingPolicies, GroupingPolicy{
			Subject:      rawGroupingPolicy[0],
			SubjectGroup: rawGroupingPolicy[1],
			Domain:       rawGroupingPolicy[2],
		})
	}

	return &groupingPolicies, nil
}

func (casbinEnf *CasbinEnforcer) AddGroupingPolicyToGroup(ctx context.Context, groupingPolicy *GroupingPolicy) error {
	_, err := casbinEnf.enforcer.AddGroupingPolicy(groupingPolicy.Subject, groupingPolicy.SubjectGroup, groupingPolicy.Domain)
	return err
}

func (casbinEnf *CasbinEnforcer) AddGroupingPoliciesToGroup(ctx context.Context, groupingPolicies *[]GroupingPolicy) error {
	for _, groupingPolicy := range *groupingPolicies {
		if _, err := casbinEnf.enforcer.AddGroupingPolicy(groupingPolicy.Subject, groupingPolicy.SubjectGroup, groupingPolicy.Domain); err != nil {
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

func (casbinEnf *CasbinEnforcer) RemoveGroupingPoliciesFromDomain(ctx context.Context, domainId string) error {
	_, err := casbinEnf.enforcer.RemoveFilteredGroupingPolicy(2, domainId)
	return err
}

func (casbinEnf *CasbinEnforcer) Enforce(ctx context.Context, request Request) (bool, error) {
	return casbinEnf.enforcer.Enforce(request.Subject, request.Domain, request.Object, request.Action, request.CtxCondition)
}

func (casbinEnf *CasbinEnforcer) Save(ctx context.Context) error {
	return casbinEnf.enforcer.SavePolicy()
}

func (casbinEnf *CasbinEnforcer) inScope(args ...interface{}) (interface{}, error) {
	subject, ok := args[0].(string)
	if !ok {
		return false, fmt.Errorf("failed to parse subject")
	}

	ctxCondition, ok := args[1].(map[string]string)
	if !ok {
		return false, fmt.Errorf("failed to ctxCondition subject")
	}

	rawCondition, ok := args[2].(string)
	if !ok {
		return false, fmt.Errorf("failed to condition subject")
	}

	var condition map[string]any
	if rawCondition != "*" {
		if err := json.Unmarshal([]byte(rawCondition), &condition); err != nil {
			return false, fmt.Errorf("failed to unmarshal condition")
		}
	}

	return inScope(subject, ctxCondition, condition), nil
}
