package consistenthash

import (
	"testing"
)

func TestMap_Get(t *testing.T) {
	//hash := New(3, func(key []byte) uint32 {
	//	i,_ := strconv.Atoi(string(key))
	//	return uint32(i)
	//})
	hash := New(3, nil)

	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":    "2",
		"11":   "2",
		"23":   "4",
		"27":   "2",
		"127":  "2",
		"3327": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("k:v input, %s : %s, get %s", k, v, hash.Get(k))
		}
	}

	hash.Add("8")

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("k:v input, %s : %s, get %s", k, v, hash.Get(k))
		}
	}
}
