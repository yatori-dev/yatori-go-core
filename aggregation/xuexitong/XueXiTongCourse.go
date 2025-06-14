package xuexitong

import (
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/interfaces"
	"io/ioutil"
	"net/http"
)

type XueXiTCourse struct {
	Cpi           int    `json:"cpi"`      // 用户唯一标识
	Key           string `json:"key"`      // classID 在课程API中为key
	CourseID      string `json:"courseId"` // 课程ID
	ChatID        string `json:"chatId"`
	CourseTeacher string `json:"courseTeacher"` // 课程老师
	CourseName    string `json:"courseName"`    //课程名
	CourseImage   string `json:"courseImage"`
	// 两个标识 暂时不知道有什么用
	CourseDataID int `json:"courseDataId"`
	ContentID    int `json:"ContentID"`
}

// CourseListApi 拉取对应账号的课程数据
func (user *XueXiTongUser) CourseListApi() (string, error) {

	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, xuexitong.ApiPullCourses, nil)

	if err != nil {
		return "", err
	}
	req.Header.Add("Cookie", user.cookie)
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (course *XueXiTCourse) GetCourseName() string {
	return course.CourseName
}

func (course *XueXiTCourse) GetCourseID() string {
	return course.CourseID
}

func (course *XueXiTCourse) TaskList() []interfaces.ITask {
	panic("implement me")
}
