package config

type Config struct {
	Setting Setting `json:"setting"`
	Users   []Users `json:"users"`
}

type Setting struct {
	BasicSetting BasicSetting `json:"basicSetting"`
	EmailInform  EmailInform  `json:"emailInform"`
	AiSetting    AiSetting    `json:"aiSetting"`
}

// CmpCourse 比较是否存在对应课程
func CmpCourse(course string, courseList []string) bool {
	for i := range courseList {
		if courseList[i] == course {
			return true
		}
	}
	return false
}
