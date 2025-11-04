package casbinauth

import (
	"fmt"
	"reflect"
	"strings"
)

func inScope(subject string, ctxCondition map[string]string, condition map[string]any) bool {
	fmt.Println(subject)
	fmt.Println(ctxCondition)
	fmt.Println(condition)

	if len(condition) == 0 {
		return true
	}

	for keyCondition, valCondition := range condition {
		switch keyCondition {
		case "and":
			subCondition, _ := valCondition.(map[string]any)
			if !inScope(subject, ctxCondition, subCondition) {
				return false
			}

		case "or":
			subCondition, _ := valCondition.(map[string]any)
			ok := false
			for subKeyCondition, subValCondition := range subCondition {
				if subKeyCondition == "and" || subKeyCondition == "or" {
					if inScope(subject, ctxCondition, map[string]any{subKeyCondition: subValCondition}) {
						ok = true
						break
					}
					continue
				}
				if isMatched(subject, ctxCondition, subKeyCondition, subValCondition) {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}

		default:
			if !isMatched(subject, ctxCondition, keyCondition, valCondition) {
				return false
			}
		}
	}

	return true
}

func isMatched(subject string, ctxCondition map[string]string, keyCondition string, valCondition any) bool {
	var op string
	field := keyCondition
	for _, suffix := range []string{"_eq", "_in"} {
		if strings.HasSuffix(keyCondition, suffix) {
			op = suffix
			field = strings.TrimSuffix(keyCondition, suffix)
			break
		}
	}

	ctxValCondition, ok := ctxCondition[field]
	if !ok || ctxValCondition == "" {
		return true
	}

	switch op {
	case "_eq":
		return compareEq(subject, ctxValCondition, valCondition)
	case "_in":
		return compareIn(ctxValCondition, valCondition)
	default:
		return compareEq(subject, ctxValCondition, valCondition)
	}
}

func compareEq(subject string, ctxValCondition string, valCondition any) bool {
	var valConditionStr string
	switch v := valCondition.(type) {
	case string:
		valConditionStr = v
	case fmt.Stringer:
		valConditionStr = v.String()
	default:
		valConditionStr = fmt.Sprintf("%v", v)
	}

	if valConditionStr == "owner_id" {
		return ctxValCondition == subject
	} else {
		return ctxValCondition == valConditionStr
	}
}

func compareIn(ctxValCondition string, valCondition any) bool {
	switch reflect.TypeOf(valCondition).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(valCondition)
		for i := 0; i < s.Len(); i++ {
			if fmt.Sprintf("%v", s.Index(i).Interface()) == ctxValCondition {
				return true
			}
		}
	}
	return false
}
