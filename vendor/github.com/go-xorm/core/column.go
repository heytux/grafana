package core

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	TWOSIDES = iota + 1
	ONLYTODB
	ONLYFROMDB
)

// database column
type Column struct {
	Name            string
	FieldName       string
	SQLType         SQLType
	Length          int
	Length2         int
	Nullable        bool
	Default         string
	Indexes         map[string]bool
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MapType         int
	IsCreated       bool
	IsUpdated       bool
	IsDeleted       bool
	IsCascade       bool
	IsVersion       bool
	fieldPath       []string
	DefaultIsEmpty  bool
	EnumOptions     map[string]int
	SetOptions      map[string]int
	DisableTimeZone bool
	TimeZone        *time.Location // column specified time zone
}

func NewColumn(name, fieldName string, sqlType SQLType, len1, len2 int, nullable bool) *Column {
	return &Column{
		Name:            name,
		FieldName:       fieldName,
		SQLType:         sqlType,
		Length:          len1,
		Length2:         len2,
		Nullable:        nullable,
		Default:         "",
		Indexes:         make(map[string]bool),
		IsPrimaryKey:    false,
		IsAutoIncrement: false,
		MapType:         TWOSIDES,
		IsCreated:       false,
		IsUpdated:       false,
		IsDeleted:       false,
		IsCascade:       false,
		IsVersion:       false,
		fieldPath:       nil,
		DefaultIsEmpty:  false,
		EnumOptions:     make(map[string]int),
	}
}

// generate column description string according dialect
func (col *Column) String(d Dialect) string {
	sql := d.QuoteStr() + col.Name + d.QuoteStr() + " "

	sql += d.SqlType(col) + " "

	if col.IsPrimaryKey {
		sql += "PRIMARY KEY "
		if col.IsAutoIncrement {
			sql += d.AutoIncrStr() + " "
		}
	}

	if d.ShowCreateNull() {
		if col.Nullable {
			sql += "NULL "
		} else {
			sql += "NOT NULL "
		}
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}

	return sql
}

func (col *Column) StringNoPk(d Dialect) string {
	sql := d.QuoteStr() + col.Name + d.QuoteStr() + " "

	sql += d.SqlType(col) + " "

	if d.ShowCreateNull() {
		if col.Nullable {
			sql += "NULL "
		} else {
			sql += "NOT NULL "
		}
	}

	if col.Default != "" {
		sql += "DEFAULT " + col.Default + " "
	}

	return sql
}

// return col's filed of struct's value
func (col *Column) ValueOf(bean interface{}) (*reflect.Value, error) {
	dataStruct := reflect.Indirect(reflect.ValueOf(bean))
	return col.ValueOfV(&dataStruct)
}

func (col *Column) ValueOfV(dataStruct *reflect.Value) (*reflect.Value, error) {
	var fieldValue reflect.Value
	if col.fieldPath == nil {
		col.fieldPath = strings.Split(col.FieldName, ".")
	}

	if dataStruct.Type().Kind() == reflect.Map {
		keyValue := reflect.ValueOf(col.fieldPath[len(col.fieldPath)-1])
		fieldValue = dataStruct.MapIndex(keyValue)
		return &fieldValue, nil
	} else if dataStruct.Type().Kind() == reflect.Interface {
		structValue := reflect.ValueOf(dataStruct.Interface())
		dataStruct = &structValue
	}

	level := len(col.fieldPath)
	fieldValue = dataStruct.FieldByName(col.fieldPath[0])
	for i := 0; i < level-1; i++ {
		if !fieldValue.IsValid() {
			break
		}
		if fieldValue.Kind() == reflect.Struct {
			fieldValue = fieldValue.FieldByName(col.fieldPath[i+1])
		} else if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			fieldValue = fieldValue.Elem().FieldByName(col.fieldPath[i+1])
		} else {
			return nil, fmt.Errorf("field  %v is not valid", col.FieldName)
		}
	}

	if !fieldValue.IsValid() {
		return nil, fmt.Errorf("field  %v is not valid", col.FieldName)
	}

	return &fieldValue, nil
}