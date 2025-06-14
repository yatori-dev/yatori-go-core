package xuexitong

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/interfaces"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type XueXiTongUser struct {
	Account  string
	Password string
	UserID   string
	CacheMap map[string]any
	cookie   string //验证码用的session
}

// pad 确保数据长度是块大小的整数倍，以便符合块加密算法的要求
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

func (user *XueXiTongUser) Login() (map[string]any, error) {
	key := []byte("u2oh6Vu^HWe4_AES")
	block, err := aes.NewCipher(key)
	if err != nil {
		log2.Print(log2.DEBUG, "Error creating cipher:", err)
		return nil, err
	}
	// 加密电话号码
	phonePadded := pad([]byte(user.Account), block.BlockSize())
	phoneCipherText := make([]byte, len(phonePadded))
	mode := cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(phoneCipherText, phonePadded)
	phoneEncrypted := base64.StdEncoding.EncodeToString(phoneCipherText)

	// 加密密码
	passwdPadded := pad([]byte(user.Password), block.BlockSize())
	passwdCipherText := make([]byte, len(passwdPadded))
	mode = cipher.NewCBCEncrypter(block, key)
	mode.CryptBlocks(passwdCipherText, passwdPadded)
	passwdEncrypted := base64.StdEncoding.EncodeToString(passwdCipherText)

	resp, err := http.PostForm(xuexitong.ApiLoginWeb, url.Values{
		"fid":               {"-1"},
		"uname":             {phoneEncrypted},
		"password":          {passwdEncrypted},
		"t":                 {"true"},
		"forbidotherlogin":  {"0"},
		"validate":          {""},
		"doubleFactorLogin": {"0"},
		"independentId":     {"0"},
		"independentNameId": {"0"},
	})
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var jsonContent map[string]interface{}
	err = json.Unmarshal(body, &jsonContent)

	if status, ok := jsonContent["status"].(bool); !ok || !status {
		return nil, errors.New(string(body))
	}
	values := resp.Header.Values("Set-Cookie")
	for _, v := range values {
		user.cookie += strings.ReplaceAll(strings.ReplaceAll(v, "HttpOnly", ""), "Path=/", "")
		//if strings.Contains(v, "UUID=") {
		//
		//}
	}

	user.CacheMap = jsonContent

	log2.Print(log2.DEBUG, "["+user.Account+"]"+"登录成功", jsonContent, err)
	if gojsonq.New().JSONString(string(body)).Find("msg2") != nil {
		return nil, errors.New(gojsonq.New().JSONString(string(body)).Find("msg2").(string))
	} else {
		return jsonContent, nil
	}

}

func (user *XueXiTongUser) UserInfo() (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (user *XueXiTongUser) CacheData() (map[string]any, error) {
	return user.CacheMap, nil
}

func (user *XueXiTongUser) CourseList() ([]interfaces.ICourse, error) {
	fmt.Println(user.cookie)
	courses, err := user.CourseListApi()
	if err != nil {
		log2.Print(log2.INFO, "["+user.Account+"] "+" 拉取失败")
	}
	var xueXiTCourse entity.XueXiTCourseJson
	err = json.Unmarshal([]byte(courses), &xueXiTCourse)
	if err != nil {
		log2.Print(log2.INFO, "["+user.Account+"] "+" 解析失败")
		panic(err)
	}
	log2.Print(log2.INFO, "["+user.Account+"] "+" 课程数量："+strconv.Itoa(len(xueXiTCourse.ChannelList)))
	// log2.Print(log2.INFO, "["+cache.Name+"] "+courses)

	var courseList = make([]interfaces.ICourse, 0)
	for i, channel := range xueXiTCourse.ChannelList {
		var flag = false
		if channel.Content.Course.Data == nil && i >= 0 && i < len(xueXiTCourse.ChannelList) {
			xueXiTCourse.ChannelList = append(xueXiTCourse.ChannelList[:i], xueXiTCourse.ChannelList[i+1:]...)
			continue
		}
		var (
			teacher      string
			courseName   string
			courseDataID int
			classId      string
			courseID     string
			courseImage  string
		)

		for _, v := range channel.Content.Course.Data {
			teacher = v.Teacherfactor
			courseName = v.Name
			courseDataID = v.Id
			userID := strings.Split(v.CourseSquareUrl, "userId=")[1]
			user.UserID = userID
			classId = strings.Split(strings.Split(v.CourseSquareUrl, "classId=")[1], "&userId")[0]
			courseID = strings.Split(strings.Split(v.CourseSquareUrl, "courseId=")[1], "&personId")[0]
			courseImage = v.Imageurl
		}

		course := XueXiTCourse{
			Cpi:           channel.Cpi,
			Key:           classId,
			CourseID:      courseID,
			ChatID:        channel.Content.Chatid,
			CourseTeacher: teacher,
			CourseName:    courseName,
			CourseImage:   courseImage,
			CourseDataID:  courseDataID,
			ContentID:     channel.Content.Id,
		}
		for _, course := range courseList {
			if course.GetCourseID() == courseID {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		courseList = append(courseList, &course)
	}
	return courseList, nil
}
