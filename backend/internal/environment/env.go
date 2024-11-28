package environment

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

func GetenvString(env string, def string) (string, error) {
	ret := os.Getenv(env)
	if ret == "" {
		return def, nil
	}
	return ret, nil
}

func MustGetenvString(env string, def string) string {
	ret, err := GetenvString(env, def)
	if err != nil {
		panic(err)
	}
	return ret
}

func GetenvInt(env string, def int) (int, error) {
	ret := os.Getenv(env)
	if ret == "" {
		return def, nil
	}
	return strconv.Atoi(ret)
}

func MustGetenvInt(env string, def int) int {
	ret, err := GetenvInt(env, def)
	if err != nil {
		panic(err)
	}
	return ret
}

func GetenvFloat64(env string, def float64) (float64, error) {
	ret := os.Getenv(env)
	if ret == "" {
		return def, nil
	}
	return strconv.ParseFloat(ret, 64)
}

func MustGetenvFloat64(env string, def float64) float64 {
	ret, err := GetenvFloat64(env, def)
	if err != nil {
		panic(err)
	}
	return ret
}

func GetenvBool(env string, def bool) (bool, error) {
	ret := os.Getenv(env)
	if ret == "" {
		return def, nil
	}
	return strconv.ParseBool(ret)
}

func MustGetenvBool(env string, def bool) bool {
	ret, err := GetenvBool(env, def)
	if err != nil {
		panic(err)
	}
	return ret
}

var isDebugOnce sync.Once
var isDebug bool

func IsDebug() bool {
	isDebugOnce.Do(func() {
		ret := os.Getenv("DEBUG")
		if ret == "" || ret == "0" || strings.EqualFold("false", ret) {
			isDebug = false
		} else {
			isDebug = true
		}
	})
	return isDebug
}
