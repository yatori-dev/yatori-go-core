package icve

import (
	"errors"
	"fmt"
	"image"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/api/icve"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 账号密码登录模块
func IcveLoginAction(cache *icve.IcveUserCache) error {
	for {
		data, err := cache.PullVerDataApi()
		instruction := data.Data.DynShowInfo.Instruction
		for {
			if instruction != "" {
				break
			}
			data, err = cache.PullVerDataApi()
			instruction = data.Data.DynShowInfo.Instruction
		}
		//fmt.Println(instruction)
		split := strings.Split(strings.Split(instruction, "：")[1], " ")
		posWords := split[0] + split[1] + split[2]
		//fmt.Println(posWords)
		if err != nil {
			return err
		}
		rand.Seed(time.Now().UnixNano())
		//randNum := rand.Intn(85000) + 5000
		randNum := 1
		img, err := icve.PullCapImgApi(data)
		//拉取vmjs
		vmjs, err := cache.PullVMApi(data, "./assets/tencentVM"+fmt.Sprintf("%d", randNum)+".js")
		if err != nil {
			return err
		}
		err1 := utils.SaveTextToFile("./assets/tencentVM"+fmt.Sprintf("%d", randNum)+".js", vmjs, false, 0644)
		if err1 != nil {
			return err1
		}
		posData := getDetectionData(posWords, img)
		//拉取eks
		var eks string
		compile := regexp.MustCompile(`(?s)window\.([A-Za-z0-9_$]+)\s*=\s*(?:'([^']*)'|"([^"]*)")`)
		submatch := compile.FindStringSubmatch(vmjs)
		if len(submatch) != 0 {
			eks = submatch[2]
		}
		//fmt.Println(posData)
		//拉取collect
		collect := icve.GetCollect(`assets\tencentVM` + fmt.Sprintf("%d", randNum) + `.js`)
		result, err2 := cache.SubmitVerApi(data, collect, eks, posData)
		if err2 != nil {
			return err2
		}
		errorCode := gojsonq.New().JSONString(result).Find("errorCode")
		if errorCode == nil {
			return errors.New(result)
		}
		//如果单纯是失败而已,则再来一次
		if errorCode == "50" {
			continue
		}
		//过验证码失败
		if errorCode.(string) != "0" {
			log.Println(errorCode.(string))
			return errors.New(result)
		}
		cache.VerCodeRandStr = gojsonq.New().JSONString(result).Find("randstr").(string)
		cache.VerCodeTicket = gojsonq.New().JSONString(result).Find("ticket").(string)
		break
	}
	//登录
	loginResult, err := cache.IcveLoginApi()
	if err != nil {
		log.Fatal(err)
	}
	loginCode := gojsonq.New().JSONString(loginResult).Find("code")
	if loginCode == nil {
		return errors.New(loginResult)
	}
	//如果登录失败
	if int(loginCode.(float64)) != 200 {
		//log.Fatal(gojsonq.New().JSONString(loginResult).Find("msg"))
		return errors.New(loginResult)
	}
	//登录成功直接赋值token
	cache.Token = gojsonq.New().JSONString(loginResult).Find("data.token").(string)
	_, err1 := cache.IcveUserEncryptApi()
	if err1 != nil {
		return err1
	}
	//主页的token获取
	//accessToken, err := cache.IcveAccessTokenApi()
	//if err != nil {
	//	return err
	//}
	//tokenCode := gojsonq.New().JSONString(accessToken).Find("code")
	//if tokenCode == nil {
	//	return errors.New(accessToken)
	//}
	////拉取鉴权token失败的话
	//if int(tokenCode.(float64)) != 200 {
	//	return errors.New(accessToken)
	//}
	//cache.AccessToken = gojsonq.New().JSONString(accessToken).Find("data").(string)

	err2 := icveZYKAction(cache)
	if err2 != nil {
		return err2
	}
	return nil
}

// Cookie登录模块
func IcveCookieLogin(cache *icve.IcveUserCache) error {
	cache.Cookies = utils.TurnCookiesFromString(cache.Password)
	for _, cookie := range cache.Cookies {
		if cookie.Name == "token" {
			cache.Token = cookie.Value
		}
		if cookie.Name == "zhzj-Token" {
			cache.AccessToken = cookie.Value
		}
	}
	err2 := icveZYKAction(cache)
	if err2 != nil {
		return err2
	}
	return nil
}

// 资源库登录初始化模块
func icveZYKAction(cache *icve.IcveUserCache) error {
	//资源库token获取----------------------------
	accessToken, err := cache.IcveZYKAccessTokenApi()
	if err != nil {
		return err
	}
	tokenCode := gojsonq.New().JSONString(accessToken).Find("code")
	if tokenCode == nil {
		return errors.New(accessToken)
	}
	//拉取鉴权token失败的话
	if int(tokenCode.(float64)) != 200 {
		return errors.New(accessToken)
	}
	cache.ZYKAccessToken = gojsonq.New().JSONString(accessToken).Find("data.access_token").(string)
	//加载个人信息------------------------------
	resultInfo, err := cache.IcveZYKPullUserInfoApi()
	if err != nil {
		return err
	}
	statusCode := gojsonq.New().JSONString(accessToken).Find("code")
	if statusCode == nil {
		return errors.New(resultInfo)
	}
	if int(statusCode.(float64)) != 200 {
		return errors.New(resultInfo)
	}
	userId := gojsonq.New().JSONString(resultInfo).Find("user.userId")
	if userId != nil {
		cache.UserId = userId.(string)
	}
	nickName := gojsonq.New().JSONString(resultInfo).Find("user.nickName")
	if nickName != nil {
		cache.NickName = nickName.(string)
	}

	phoneNumber := gojsonq.New().JSONString(resultInfo).Find("user.phonenumber")
	if phoneNumber != nil {
		cache.PhoneNumber = phoneNumber.(string)
	}

	sex := gojsonq.New().JSONString(resultInfo).Find("user.sex")
	if sex != nil {
		cache.Sex = sex.(string)
	}
	return nil
}

// 获取验证码识别数据
func getDetectionData(posWords string, img image.Image) string {
	dets, err := ddddocr.AutoDetectionForTencent(img, 3)
	if err != nil {
		log.Fatal(err)
	}
	result := "["
	rs := []rune(posWords)
	for i, word := range rs {
		for _, det := range dets {
			if det.Describe != string(word) {
				continue
			}
			x := det.BBox.Min.X + (det.BBox.Max.X-det.BBox.Min.X)/2 - 10
			y := det.BBox.Min.Y + (det.BBox.Max.Y-det.BBox.Min.Y)/2 - 8
			result += `{"elem_id":` + strconv.Itoa(i+1) + `,"type":"DynAnswerType_POS","data":"` + fmt.Sprintf("%d,%d", x, y) + `"}`
		}
		if i != len(dets)-1 {
			result += ","
		} else {
			result += "]"
		}
	}

	return result
}
