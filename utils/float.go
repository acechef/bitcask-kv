package utils

import "strconv"

func Float64FromBytes(val []byte) float64 {
	f, _ := strconv.ParseFloat(string(val), 64)
	return f
}

func Float64ToBytes(val float64) []byte {
	return []byte(strconv.FormatFloat(val, 'f', -1, 64))
}
