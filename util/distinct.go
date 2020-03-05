package util

//去重函数
func Distinct(tempItem []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(tempItem); i++ {
		repect := false
		for j := i + 1; j < len(tempItem); j++ {
			if tempItem[i] == tempItem[j] {
				repect = true
				break
			}
		}
		if !repect {
			newArr = append(newArr, tempItem[i])
		}
	}
	return
}
