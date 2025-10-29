package util

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

func BuildQuery(query *gorm.DB, filter interface{}, refModel interface{}) *gorm.DB {
	filterFieldValueMap := detectFieldValueMapFromQueryTag(filter)
	refModelFieldValueMap := detectFieldValueMapFromGormTag(refModel)

	for filterFieldName, filterValue := range filterFieldValueMap {
		if !isValidField(filterValue) {
			continue
		}

		filterFieldName, op := parseOperator(filterFieldName, refModelFieldValueMap)
		if op == "" {
			continue
		}

		switch op {
		case "LIKE", "ILIKE":
			query = query.Where(fmt.Sprintf("%s %s ?", filterFieldName, op), "%"+filterValue.String()+"%")
		default:
			query = query.Where(fmt.Sprintf("%s %s ?", filterFieldName, op), filterValue.Interface())
		}
	}

	return query
}

func detectFieldValueMapFromQueryTag(model interface{}) map[string]reflect.Value {
	fieldValueMap := make(map[string]reflect.Value)

	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagConfig := field.Tag.Get("query")
		if tagConfig == "" {
			continue
		}
		fieldName := tagConfig
		value := v.Field(i)
		fieldValueMap[fieldName] = value
	}

	return fieldValueMap
}

func detectFieldValueMapFromGormTag(model interface{}) map[string]reflect.Value {
	fieldValueMap := make(map[string]reflect.Value)

	v := reflect.ValueOf(model).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagConfig := field.Tag.Get("gorm")
		if tagConfig == "" {
			continue
		}
		columnName := strings.Split(tagConfig, ";")[0] // If tag has many config
		fieldName := columnName[strings.Index(columnName, ":")+1:]
		value := v.Field(i)
		fieldValueMap[fieldName] = value
	}

	return fieldValueMap
}

func isValidField(fieldValue reflect.Value) bool {
	if !fieldValue.IsValid() {
		return false
	}
	if fieldValue.Kind() == reflect.String && fieldValue.String() == "" {
		return false
	}
	if fieldValue.Kind() == reflect.Slice && fieldValue.Len() == 0 {
		return false
	}
	return true
}

func parseOperator(filterFieldName string, refModelFieldValueMap map[string]reflect.Value) (string, string) {
	op := ""

	switch {
	case strings.HasSuffix(filterFieldName, "_eq"):
		op = "="
		filterFieldName = strings.TrimSuffix(filterFieldName, "_eq")

	case strings.HasSuffix(filterFieldName, "_gte"):
		op = ">="
		filterFieldName = strings.TrimSuffix(filterFieldName, "_gte")
	case strings.HasSuffix(filterFieldName, "_lte"):
		op = "<="
		filterFieldName = strings.TrimSuffix(filterFieldName, "_lte")

	case strings.HasSuffix(filterFieldName, "_like"):
		op = "LIKE"
		filterFieldName = strings.TrimSuffix(filterFieldName, "_like")
	case strings.HasSuffix(filterFieldName, "_ilike"):
		op = "ILIKE"
		filterFieldName = strings.TrimSuffix(filterFieldName, "_ilike")

	case strings.HasSuffix(filterFieldName, "_in"):
		op = "IN"
		filterFieldName = strings.TrimSuffix(filterFieldName, "_in")

	default:
		refModelValue, ok := refModelFieldValueMap[filterFieldName]
		if !ok {
			return "", ""
		}

		switch refModelValue.Kind() {
		case reflect.String:
			op = "ILIKE"
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Bool:
			op = "="
		default:
			return "", ""
		}
	}

	return filterFieldName, op
}
