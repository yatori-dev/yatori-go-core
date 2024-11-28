package examples

import (
	"fmt"
	cqie "github.com/Yatori-Dev/yatori-go-core/aggregation/cqie"
	cqieApi "github.com/Yatori-Dev/yatori-go-core/api/cqie"
	"github.com/Yatori-Dev/yatori-go-core/global"
	"github.com/Yatori-Dev/yatori-go-core/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// 测试加密函数
func TestCqieEncrypted(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]
	// 调用函数进行加密
	accEncrypted := utils.CqieEncrypt(users.Account)
	passEncrypted := utils.CqieEncrypt(users.Password)
	// 输出加密后的数据
	fmt.Printf("Encrypted data: %x\n", accEncrypted)
	fmt.Printf("Encrypted data: %x\n", passEncrypted)
}

// 登录测试函数
func TestCqieLogin(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	users := global.Config.Users[4]
	cache := cqieApi.CqieUserCache{Account: users.Account, Password: users.Password}
	cqie.CqieLoginAction(&cache)
}

func TestCourse(t *testing.T) {
	startPos := 0
	stopPos := 3
	maxPos := 3
	for {
		if stopPos >= maxPos {
			maxPos = startPos + 3
		}
		fmt.Println(startPos, stopPos, maxPos)
		testSubmit(startPos, stopPos, maxPos)
		startPos = startPos + 3
		stopPos = stopPos + 3
		time.Sleep(3 * time.Second)
	}
}

func testSubmit(startPos, stopPos, maxPos int) {

	url := "https://study.cqie.edu.cn/gateway/system/orgStudent/updateStudyVideoPlan"
	method := "POST"

	payload := strings.NewReader(`{
   "id": "ce70b5cd1765f698788c84d0eec5a95a",
  "orgId": "1",
  "deptId": "f8955096610de9b61f0e89762fe44448",
  "majorId": "966c92d8abcec472a717ca5af8c24a49",
  "version": "ZSB_DSJ_24",
  "courseId": "1cbcd0b22030fdf75117cb02d2c766ec",
  "studentCourseId": "e256c4c26b3e0dc27901029755396bdd",
  "unitId": "84ed9e4674d811bc10a9f55390231f5f",
  "knowledgeId": null,
  "videoId": "65da2135311dc5cc461d53c6c827cc45",
  "studentId": "1e49b8b4ea23d07eb0724d0c927ed2bc",
  "studyTime": "2024-11-27 19:49:24",
  "startPos":` + strconv.Itoa(startPos) + `,
  "stopPos": ` + strconv.Itoa(stopPos) + `,
  "studyTime": "2024-11-27 19:49:24",
  "maxCurrentPos": ` + strconv.Itoa(maxPos) + `,
  "coursewareId": "558bf1e99f2727d493dbe2ed05983115"
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "eyJhbGciOiJIUzUxMiJ9.eyIwIjoiMSIsInVzZXJfaWQiOiJiODc5N2FkNjdhMGNmZDk2N2ViNGJhOWM4ODBkOWY5MCIsImFwcElkIjoiMjAyNDExMjcwMDk3ODc0MjIyMiIsInVzZXJfa2V5IjoiMTNlNDY2ZDUtMjZhZi00NzI2LTgyZmYtZmUzNjk0NTQ2YTViIiwidXNlcm5hbWUiOiLlrovlhYPlhbUifQ.zAtSLcc3xaLqvWpJjbmiXEgfgV-_JRS1B90iRBXfS_UT9OIELVplPKY3ZiR_0CPx5CxleDG76RVS93fItINu5A")
	req.Header.Add("Cookie", "JSESSIONID=D88401C497B92092F2A5F1D0F403A6D2")
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
