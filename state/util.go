package state

// fill slice by 0.
func ZeroClear(slice []int64) []int64 {
	return FillNumber(slice, 0)
}

// fill slice for given number.
func FillNumber(slice []int64, num int64) []int64 {
	for i, _ := range slice {
		slice[i] = num
	}
	return slice
}

// fill slice for empty string.
func StrClear(slice []string) []string {
	for i, _ := range slice {
		slice[i] = ""
	}
	return slice
}
