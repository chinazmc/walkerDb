package utils

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func Max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
func IntToString(param interface{}) string {
	switch param.(type) {
	case int:
		return strconv.Itoa(param.(int))
	case uint:
		return strconv.FormatUint(uint64(param.(uint)), 10)
	case int8:
		return strconv.Itoa(int(param.(int8)))
	case uint8:
		return strconv.FormatUint(uint64(param.(uint8)), 10)
	case int16:
		return strconv.Itoa(int(param.(int16)))
	case uint16:
		return strconv.FormatUint(uint64(param.(uint16)), 10)
	case int32:
		return strconv.Itoa(int(param.(int32)))
	case uint32:
		return strconv.FormatUint(uint64(param.(uint32)), 10)
	case int64:
		return strconv.FormatInt(param.(int64), 10)
	case uint64:
		return strconv.FormatUint(uint64(param.(uint64)), 10)
	}
	return strconv.Itoa(param.(int))
}
func RandomString(n int) string {
	b := make([]rune, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range b {
		b[i] = defaultLetters[r.Intn(len(defaultLetters))]
	}
	return string(b)
}

// 找出list1中存在，list2中不存在的元素
func GetDiff(list1, list2 []interface{}) []interface{} {
	var b bool
	newList := make([]interface{}, 0)
	for _, obj1 := range list1 {
		b = false
		for _, obj2 := range list2 {
			if obj1 == obj2 {
				b = true
			}
		}
		if !b {
			newList = append(newList, obj1)
		}
	}
	return newList
}

func StructConvertMap(obj interface{}) map[string]string {
	var data = make(map[string]string)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	for i := 0; i < t.NumField(); i++ {
		tagName := t.Field(i).Tag.Get("json")
		if tagName != "" && tagName != "-" {
			if s := fmt.Sprint(v.Field(i).Interface()); s != "" {
				data[tagName] = s
			}
		}
	}
	return data
}

func IsExistsString(strs []string, key string) bool {
	if len(strs) == 0 || key == "" {
		return false
	}
	for _, str := range strs {
		if str == key {
			return true
		}
	}
	return false
}

func ContainsInt(slice []int, val int) bool {
	for _, v := range slice {
		if val == v {
			return true
		}
	}
	return false
}

func TimestampToStr(t int64) string {
	return time.Unix(t, 0).Format("2006-01-02 15:04:05")
}

// 判断当前
func IsBetweenTime(startTime, endTime int64) bool {
	now := time.Now().Unix() * 1000
	if now >= startTime && endTime >= now {
		return true
	}
	return false
}

func Atoi(s string) uint64 {
	u, _ := strconv.ParseUint(s, 10, 64)
	return u
}

// CommaString string split by comma
type CommaString string

func (c CommaString) IsNull() bool {
	return strings.TrimLeft(strings.TrimRight(string(c), ","), ",") == ""
}

func (c CommaString) Split() (list []string) {
	v := strings.TrimLeft(strings.TrimRight(string(c), ","), ",")
	if len(v) == 0 {
		return nil
	}
	return strings.Split(v, ",")
}

func (c CommaString) SplitAsInt() (list []uint64) {
	for _, v := range c.Split() {
		list = append(list, Atoi(v))
	}
	return list
}

func GoFunc() {}

func Parse(layout string, value string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	return time.ParseInLocation(layout, value, loc)
}

func TimestampToDateTime(t int64) string {
	tm := time.Unix(t, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// 判断当前时间是否在指定的一段时间范围之内
func NowTimeIsBetweenStartTimeAndEndTime(startTime, endTime int64) bool {
	now := time.Now().Unix()
	if startTime < now {
		if endTime > now {
			return true
		}
	}
	return false
}

func StringsToUint32(req ...string) ([]uint32, error) {
	resp := make([]uint32, len(req))
	for k, v := range req {
		atoi, err := strconv.Atoi(v)
		if err != nil {
			return resp, err
		}
		resp[k] = uint32(atoi)
	}

	return resp, nil
}

func GetArgsByUint32(list ...[]uint32) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GetArgsByUint16(list ...[]uint16) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GetArgsByUint64(list ...[]uint64) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GetArgsByInt32(list ...[]int32) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GetArgsByInt64(list ...[]int64) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GetArgsByInt(list ...[]int) []interface{} {
	var res []interface{}
	for _, v := range list {
		for _, v2 := range v {
			res = append(res, v2)
		}
	}
	return res
}

func GenSeq() string {
	str := time.Now().Format("20060102-150304")
	str = strings.Replace(str, "-", "", 1)
	for i := 0; i < 5; i++ {
		n := rand.Intn(10)
		str += strconv.Itoa(n)
	}
	return str
}
