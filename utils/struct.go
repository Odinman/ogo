// Package utils provides ...
package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type StructColumn struct {
	Tag        string
	Name       string
	TagOptions TagOptions
	ExtTag     string
	ExtOptions TagOptions
	Type       reflect.Type
	Index      []int
}

/* {{{ func ReadStructColumns(i interface{}, tag string, underscore bool) (cols []string)
 * 从struct type中读取字段名
 * 默认从struct的FieldName读取, 如果tag里有db, 则以db为准
 */
func ReadStructColumns(i interface{}, underscore bool, tags ...string) (cols []StructColumn) {
	if t := toType(i); t.Kind() != reflect.Struct {
		return
	} else {
		return typeStructColumns(t, underscore, tags...)
	}
}

/* }}} */

/* {{{ func FieldByIndex(v reflect.Value, index []int) reflect.Value
 * 通过索引返回field
 */
func FieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

/* }}} */

/* {{{ func ImportVal(i interface{}, import map[string]string) (err error)
 * 将tag匹配的值导入结构
 */
func ImportValue(i interface{}, is map[string]string) (err error) {
	v := reflect.ValueOf(i)
	if cols := ReadStructColumns(i, true); cols != nil {
		for _, col := range cols {
			for tag, iv := range is {
				if col.TagOptions.Contains(tag) {
					fv := FieldByIndex(v, col.Index)
					switch fv.Type().String() {
					case "*string":
						fv.Set(reflect.ValueOf(&iv))
					case "string":
						fv.Set(reflect.ValueOf(iv))
					case "*int64":
						pv, _ := strconv.ParseInt(iv, 10, 64)
						fv.Set(reflect.ValueOf(&pv))
					case "int64":
						pv, _ := strconv.ParseInt(iv, 10, 64)
						fv.Set(reflect.ValueOf(pv))
					case "*int":
						tv, _ := strconv.ParseInt(iv, 10, 0)
						pv := int(tv)
						fv.Set(reflect.ValueOf(&pv))
					case "int":
						tv, _ := strconv.ParseInt(iv, 10, 0)
						pv := int(tv)
						fv.Set(reflect.ValueOf(pv))
					default:
						err = fmt.Errorf("field(%s) not support %s", col.Tag, fv.Kind().String())
					}
				}
			}
		}
	}
	return
}

/* }}} */

/* {{{ func typeStructColumns(t reflect.Type, underscore bool, tags ...string) (cols []StructColumn)
 * 从struct中读取字段名
 * 默认从struct的FieldName读取, 如果tag里有db, 则以db为准
 */
func typeStructColumns(t reflect.Type, underscore bool, tags ...string) (cols []StructColumn) {
	tag := "db"        // 默认tag是"db"
	extTag := "filter" // 默认扩展tag是filter
	if len(tags) > 0 {
		tag = tags[0]
	}
	if len(tags) > 1 {
		extTag = tags[1]
	}
	n := t.NumField()
	for i := 0; i < n; i++ {
		index := make([]int, 0)
		f := t.Field(i)
		index = append(index, i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct { //匿名struct , 也就是嵌套
			// Recursively add nested fields in embedded structs.
			subcols := typeStructColumns(f.Type, underscore, tags...)
			// 如果重名则不append, drop
			for _, subcol := range subcols {
				shouldAppend := true
				for _, col := range cols {
					if subcol.Tag == col.Tag {
						shouldAppend = false
						break
					}
				}
				if shouldAppend {
					for _, ii := range subcol.Index {
						subcol.Index = append(index, ii)
					}
					cols = append(cols, subcol)
				}
			}
		} else {
			// parse tag
			ts, tops := ParseTag(f.Tag.Get(tag))
			if ts == "" { //为空,则取字段名
				if underscore {
					ts = Underscore(f.Name)
				} else {
					ts = f.Name
				}
			}
			// parse exttag
			extTs, extTops := ParseTag(f.Tag.Get(extTag))
			// struct col
			sc := StructColumn{
				Tag:        ts,
				Name:       f.Name,
				TagOptions: tops,
				ExtTag:     extTs,
				ExtOptions: extTops,
				Type:       f.Type,
				Index:      index,
			}
			//检查同名,有则覆盖
			shouldAppend := true
			for index, col := range cols {
				if col.Tag == sc.Tag {
					cols[index] = sc
					shouldAppend = false
					break
				}
			}
			if shouldAppend {
				cols = append(cols, sc)
			}
		}
	}
	return
}

/* }}} */

/* {{{ func toType(i interface{}) reflect.Type
 * 如果是指针, 则调用Elem()至Type为止, 如果Type不是struct, 报错
 */
func toType(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)

	// If a Pointer to a type, follow
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	//if t.Kind() != reflect.Struct {
	//	return nil, fmt.Errorf("utils: Cannot SELECT into this type: %v", reflect.TypeOf(i))
	//}
	return t
}

/* }}} */

/* {{{ Underscore
 * 小程序, 把驼峰式转化为匈牙利式
 */
func Underscore(camelCaseWord string) string {
	underscoreWord := regexp.MustCompile("([A-Z]+)([A-Z][a-z])").ReplaceAllString(camelCaseWord, "${1}_${2}")
	underscoreWord = regexp.MustCompile("([a-z\\d])([A-Z])").ReplaceAllString(underscoreWord, "${1}_${2}")
	underscoreWord = strings.Replace(underscoreWord, "-", "_", 0)
	underscoreWord = strings.ToLower(underscoreWord)
	return underscoreWord
}

/* }}} */

/* {{{ func IsEmptyValue(v reflect.Value) bool
 *
 */
func IsEmptyValue(v reflect.Value) bool {
	const deref = false
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if deref {
			if v.IsNil() {
				return true
			}
			return IsEmptyValue(v.Elem())
		} else {
			return v.IsNil()
		}
	case reflect.Struct:
		// return true if all fields are empty. else return false.
		return v.Interface() == reflect.Zero(v.Type()).Interface()
		// for i, n := 0, v.NumField(); i < n; i++ {
		// 	if !isEmptyValue(v.Field(i), deref) {
		// 		return false
		// 	}
		// }
		// return true
	}
	return false
}

/* }}} */

/* {{{ func GetRealString(v reflect.Value) string
 *
 */
func GetRealString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', 2, 64)
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return GetRealString(v.Elem())
	default:
		//nothing
	}
	return ""
}

/* }}} */

/* {{{ func FieldExists(i interface{},f string) bool
 * 判断一个结构变量是否有某个字段
 */
func FieldExists(i interface{}, f string) bool {
	r := reflect.ValueOf(i)
	fv := reflect.Indirect(r).FieldByName(f)
	if !fv.IsValid() {
		return false
	} else {
		return true
	}
}

/* }}} */
