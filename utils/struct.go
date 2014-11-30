// Package utils provides ...
package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/* {{{ func readStructField
 * 从struct中读取字段名
 * 默认从struct的FieldName读取, 如果tag里有db, 则以db为准
 */
func ReadStructColumns(i interface{}, tag string, underscore bool) (cols []string) {
	t, err := toType(i)
	if err != nil {
		return
	}
	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct { //匿名struct , 也就是嵌套
			// Recursively add nested fields in embedded structs.
			subcols := ReadStructColumns(f.Type, tag, underscore)
			// 如果重名则不append, drop
			for _, subcol := range subcols {
				shouldAppend := true
				for _, col := range cols {
					if subcol == col {
						shouldAppend = false
						break
					}
				}
				if shouldAppend {
					cols = append(cols, subcol)
				}
			}
		} else {
			columnName := f.Tag.Get(tag)
			if columnName == "" {
				if underscore {
					columnName = Underscore(f.Name)
				} else {
					columnName = f.Name
				}
			}
			//检查同名,有则覆盖
			shouldAppend := true
			for index, col := range cols {
				if col == columnName {
					cols[index] = columnName
					shouldAppend = false
					break
				}
			}
			if shouldAppend {
				cols = append(cols, columnName)
			}
		}
	}
	return
}

/* }}} */

/* {{{ toType(i interface{}) (reflect.Type, error)
 * 如果是指针, 则调用Elem()至Type为止, 如果Type不是struct, 报错
 */
func toType(i interface{}) (reflect.Type, error) {
	t := reflect.TypeOf(i)

	// If a Pointer to a type, follow
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("utils: Cannot SELECT into this type: %v", reflect.TypeOf(i))
	}
	return t, nil
}

/* }}} */

/* {{{
 *小程序, 把驼峰式转化为匈牙利式
 */
func Underscore(camelCaseWord string) string {
	underscoreWord := regexp.MustCompile("([A-Z]+)([A-Z][a-z])").ReplaceAllString(camelCaseWord, "${1}_${2}")
	underscoreWord = regexp.MustCompile("([a-z\\d])([A-Z])").ReplaceAllString(underscoreWord, "${1}_${2}")
	underscoreWord = strings.Replace(underscoreWord, "-", "_", 0)
	underscoreWord = strings.ToLower(underscoreWord)
	return underscoreWord
}

/* }}} */
