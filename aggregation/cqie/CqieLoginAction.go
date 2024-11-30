package cqie

import (
	"errors"
	"time"

	cqieApi "github.com/Yatori-Dev/yatori-go-core/api/cqie"
	"github.com/Yatori-Dev/yatori-go-core/utils"
	"github.com/Yatori-Dev/yatori-go-core/utils/log"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
)

type CqieCourse struct {
	Id         string    //这个是课程的id也就是courseId
	CourseName string    //课程名称
	SumUnit    int       //一共多少单元
	HaveUnit   int       //以学单元
	SumTime    time.Time //总时长
	HaveTime   time.Time //以学时长
	Learned    string    //已学进度

}

type CqieVideo struct {
	VideoId    string //视屏Id
	CourseId   string //课程Id
	UnitId     string //单元Id
	VideoName  string //视屏名称
	TimeLength int    //视屏时长
}

// CqieLoginAction 登录API聚合整理
// {"refresh_code":1,"status":false,"msg":"账号密码不正确"}
// {"_code": 1, "status": false,"msg": "账号登录超时，请重新登录", "result": {}}
func CqieLoginAction(cache *cqieApi.CqieUserCache) error {
	for {
		path, cookie := cache.VerificationCodeApi() //获取验证码
		cache.SetCookie(cookie)
		img, _ := utils.ReadImg(path)                                  //读取验证码图片
		codeResult := utils.AutoVerification(img, ort.NewShape(1, 26)) //自动识别
		utils.DeleteFile(path)                                         //删除验证码文件
		cache.SetVerCode(codeResult)                                   //填写验证码
		jsonStr, _ := cache.LoginApi()                                 //执行登录
		log.Print(log.DEBUG, "["+cache.Account+"] "+"LoginAction---"+jsonStr)
		if gojsonq.New().JSONString(jsonStr).Find("msg") == "验证码有误！" {
			continue
		} else if int(gojsonq.New().JSONString(jsonStr).Find("code").(float64)) != 200 {
			return errors.New(gojsonq.New().JSONString(jsonStr).Find("msg").(string))
		}
		cache.SetAccess_Token(gojsonq.New().JSONString(jsonStr).Find("data.access_token").(string))
		cache.SetToken(gojsonq.New().JSONString(jsonStr).Find("data.user.token").(string))
		cache.SetUserId(gojsonq.New().JSONString(jsonStr).Find("data.user.userId").(string))
		cache.SetAppId(gojsonq.New().JSONString(jsonStr).Find("data.user.appId").(string))
		cache.SetIpaddr(gojsonq.New().JSONString(jsonStr).Find("data.user.ipaddr").(string))
		cache.SetDeptId(gojsonq.New().JSONString(jsonStr).Find("data.user.deptId").(string))
		userJson, err := cache.UserDetailsApi(8, nil) //获取用户信息
		if err != nil {
			return err
		}
		cache.SetStudentId(gojsonq.New().JSONString(userJson).Find("data.id").(string))
		cache.SetUserName(gojsonq.New().JSONString(userJson).Find("data.userName").(string))
		cache.SetOrgId(gojsonq.New().JSONString(userJson).Find("data.orgId").(string))
		cache.SetUserId(gojsonq.New().JSONString(userJson).Find("data.userId").(string))
		cache.SetMobile(gojsonq.New().JSONString(userJson).Find("data.mobile").(string))
		cache.SetOrgMajorId(gojsonq.New().JSONString(userJson).Find("data.orgMajorId").(string))

		log.Print(log.INFO, "["+cache.Account+"] "+" 登录成功")
		break
	}
	return nil
}

// CqiePullCourseListAction 拉取课程列表信息
func CqiePullCourseListAction(cache *cqieApi.CqieUserCache) ([]CqieCourse, error) {
	var courseList []CqieCourse
	courseApi, err := cache.PullCourseList(5, nil)
	if err != nil {
	}
	if gojsonq.New().JSONString(courseApi).Find("msg") != "操作成功" {
		return courseList, errors.New("获取数据失败：" + courseApi)
	}
	jsonList := gojsonq.New().JSONString(courseApi).Find("data.records")
	if items, ok := jsonList.([]interface{}); ok {
		for _, item := range items {
			// 每个 item 是 map[string]interface{} 类型
			if obj, ok := item.(map[string]interface{}); ok {
				sumTime, _ := time.Parse("15:04")
				haveTime, _ := time.Parse("15:04")
				courseList = append(courseList, CqieCourse{
					Id:         obj["id"].(string),
					CourseName: obj["name"].(string),
					SumUnit:    int(obj["sumUnit"].(float64)),
					HaveUnit:   int(obj["haveUnit"].(float64)),
					SumTime:    sumTime,
					HaveTime:   haveTime,
					Learned:    obj["learned"].(string),
				})
			}
		}
	}
	return courseList, nil
}

// 拉取对应课程的所有视屏
func PullCourseVideoListAction(cache *cqieApi.CqieUserCache, course *CqieCourse) {

}
