package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func asJson(data []byte, err error) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err = json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func limit(str string, limit int) string {
	return str[0:min(limit, len(str))]
}

func m(data interface{}, keys ...string) interface{} {
	result := data
	for _, key := range keys {
		switch v := result.(type) {
		case map[string]interface{}:
			result = v[key]
		case map[interface{}]interface{}:
			result = v[key]
		}
		if result == nil {
			return result
		}
	}
	return result
}
func ms(data interface{}, keys ...string) string {
	return fmt.Sprintf("%s", m(data, keys...))
}

func mns(data interface{}, keys ...string) string {
	return strconv.Itoa(int(m(data, keys...).(float64)))

}

func mn(data interface{}, keys ...string) int {
	return int(m(data, keys...).(float64))

}

func l(data interface{}) []interface{} {
	if data == nil {
		return make([]interface{}, 0)
	}
	return data.([]interface{})
}

func nilsafe(data interface{}) interface{} {
	if data == nil {
		return ""
	}
	return data;
}
