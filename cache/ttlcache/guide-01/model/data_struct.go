package model

import "time"

type DataStruct struct {
	Field1 string        `json:"field1"`
	Field2 int32         `json:"field2"`
	Field3 int64         `json:"field3"`
	Field4 float32       `json:"field4"`
	Field5 float64       `json:"field5"`
	Field6 time.Time     `json:"field6"`
	Field7 SubDataStruct `json:"field7"`
}

type SubDataStruct struct {
	Field1 string `json:"field1"`
	Field2 int32  `json:"field2"`
	Field3 int64  `json:"field3"`
}
