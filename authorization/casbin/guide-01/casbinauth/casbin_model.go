package casbinauth

type CustomCasbinRule struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Ptype string `gorm:"size:100"`
	V0    string `gorm:"size:100"`
	V1    string `gorm:"size:100"`
	V2    string `gorm:"size:100"`
	V3    string `gorm:"size:100"`
	V4    string `gorm:"size:text"`
	V5    string `gorm:"size:100"`
}

func (CustomCasbinRule) TableName() string {
	return "custom_casbin_rule"
}

type Request struct {
	Subject      string
	Domain       string
	Object       string
	Action       string
	CtxCondition string
}

type Policy struct {
	Domain    string
	Object    string
	Action    string
	Condition string
}

type GroupingPolicy struct {
	Subject string
	Domain  string
}
