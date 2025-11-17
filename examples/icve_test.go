package examples

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"fmt"
	log2 "log"
	"strings"
	"testing"

	action "github.com/yatori-dev/yatori-go-core/aggregation/icve"
	"github.com/yatori-dev/yatori-go-core/api/icve"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 测试登录
func TestIcveLogin(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[44]
	cache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//userCache.IcveLoginApi()
	err := action.IcveLoginAction(&cache)
	if err != nil {
		fmt.Println(err)
	}

}

// 测试拉取课程
func TestIcveCourseList(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := global.Config.Users[45]
	cache := icve.IcveUserCache{
		Account:  user.Account,
		Password: user.Password,
	}
	//err := action.IcveLoginAction(&cache)
	err := action.IcveCookieLogin(&cache)
	if err != nil {
		fmt.Println(err)
	}
	courseList, err := action.PullZYKCourseAction(&cache)
	if err != nil {
		fmt.Println(err)
	}
	for _, course := range courseList {
		fmt.Println(course)
		nodeList, err1 := action.PullZYKCourseNodeAction(&cache, course)
		if err1 != nil {
			fmt.Println(err1)
		}
		for _, node := range nodeList {
			if node.Speed >= 100 {
				continue
			}
			fmt.Println(node)
			result, err2 := action.SubmitZYKStudyTimeAction(&cache, node)
			if err2 != nil {
				fmt.Println(err2)
			}
			log2.Printf("(%s)学习状态：%s", node.Name, result)
		}
	}
	//userCache.CourseListApi()
}

// 智慧职教作业提交加密函数
func TestIcveExamSubmit(t *testing.T) {
	testTxt := `{
    "categoryId": "3",
    "courseId": "75eb7297-01f9-45ed-91f2-f38a084a48cc",
    "courseInfoId": "c58b5f83-25cc-467f-9309-bfbf1e7945f4",
    "examId": "00731616-ebc9-4a08-a0bb-dd42c6b7f5eb",
    "examTime": 73,
    "groupId": 0,
    "isLast": true,
    "status": "",
    "taskExamProblemRecordList": [
        {
            "questionNo": 0,
            "optionSort": "[{\"Content\":\"悲伤压抑\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"欢快活泼\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"神秘深沉\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"愤怒暴躁\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "0c4fc046-8224-4629-9b7b-2c0c47c6a638"
        },
        {
            "questionNo": 1,
            "optionSort": "[{\"Content\":\"金属铃铛\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"软质彩带\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"玻璃球\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"尖锐木棍\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "2f884822-f32c-4987-a8bd-544f2a256f16"
        },
        {
            "questionNo": 2,
            "optionSort": "[{\"Content\":\"只出现一次\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"多次重复\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"不断变化\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"与其他动作毫无关联\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "34edee22-6569-4466-9514-292a367c1e20"
        },
        {
            "questionNo": 3,
            "optionSort": "[{\"Content\":\"重复性动作\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"大幅度跳跃\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"模仿小动物\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"手部简单动作\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "3d03d086-9f5e-4c4d-911e-d6de85111468"
        },
        {
            "questionNo": 4,
            "optionSort": "[{\"Content\":\"复杂性\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"童趣性\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"技巧性\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"艺术性\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "406ac910-fb07-43d1-a0a4-9bb635adaa5a"
        },
        {
            "questionNo": 5,
            "optionSort": "[{\"Content\":\"家长喜好\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"教师个人风格\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"身体发展水平\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"舞蹈比赛要求\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "2",
            "paperId": "75f3bdde-9e03-441f-8c8a-015a614c8a15"
        },
        {
            "questionNo": 6,
            "optionSort": "[{\"Content\":\"复杂多样难以理解\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"简单易操作\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"频繁快速转换\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"一成不变\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "ae33a7b4-cd78-4356-a926-9cce32d826af"
        },
        {
            "questionNo": 7,
            "optionSort": "[{\"Content\":\"要求幼儿完全模仿\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"激发兴趣并引导动作\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"仅口头讲解不演示\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"强调动作标准化\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "ce1ae7c5-e0e6-4383-a66a-df0348afd956"
        },
        {
            "questionNo": 8,
            "optionSort": "[{\"Content\":\"仅使用地面动作\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"结合站立、蹲下、跳跃等不同高度动作\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"固定不变的位置\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"忽略空间变化\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "1",
            "paperId": "d4bbc288-2622-41bf-8bca-d109a1e4a5aa"
        },
        {
            "questionNo": 9,
            "optionSort": "[{\"Content\":\"模仿小动物走路\",\"Label\":0,\"SortOrder\":\"0\"},{\"Content\":\"简单的跳跃\",\"Label\":1,\"SortOrder\":\"1\"},{\"Content\":\"复杂的旋转技巧\",\"Label\":2,\"SortOrder\":\"2\"},{\"Content\":\"挥手打招呼\",\"Label\":3,\"SortOrder\":\"3\"}]",
            "answer": "2",
            "paperId": "f11f96b0-5cbe-4e7d-a165-5c360c5203e4"
        }
    ],
    "updateBy": "",
    "updateTime": "",
    "userId": "",
    "id": "793d2d3c-355e-4488-8249-2a899021e339",
    "resitId": "",
    "device": 1
}`
	//去除所有空格以及换行符并且转bytes
	encTxt := []byte(strings.Join(strings.Fields(testTxt), ""))
	ecb, err := aesEncryptECB(encTxt, []byte("djekiytolkijduey"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(ecb))
}

// PKCS7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// AES-ECB 加密
func aesEncryptECB(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	src = pkcs7Padding(src, bs)
	encrypted := make([]byte, len(src))
	for start := 0; start < len(src); start += bs {
		block.Encrypt(encrypted[start:start+bs], src[start:start+bs])
	}
	return encrypted, nil
}
