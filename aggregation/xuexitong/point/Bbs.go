package point

import (
	"encoding/json"
	"errors"
	"fmt"
	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
	"github.com/thedevsaddam/gojsonq"
	ort "github.com/yalue/onnxruntime_go"
	xuexitong2 "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	"github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/config"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"log"
	"regexp"
)

type BBsTopic struct {
	Fid              int    `json:"fid"`
	UrlToken         string `json:"urlToken"`
	PraiseCount      int    `json:"praiseCount"`
	SPraisecount     string `json:"s_praisecount"`
	CodeTitle        string `json:"code_title"`
	Title            string `json:"title"`
	TopTime          int64  `json:"topTime"`
	Type             int    `json:"type"`
	Uuid             string `json:"uuid"`
	Content          string `json:"content"`
	UpdateTime       int64  `json:"update_time"`
	LastReplyUserId  int    `json:"last_reply_user_id"`
	Top              int    `json:"top"`
	ChapterId        int    `json:"chapterId"`
	Ispublic         int    `json:"ispublic"`
	SReplycount      string `json:"s_replycount"`
	ContentImgs      string `json:"content_imgs"`
	Id               int    `json:"id"`
	HasSensitive     int    `json:"hasSensitive"`
	LastReplyTime    int64  `json:"last_reply_time"`
	QuoteInfo        string `json:"quoteInfo"`
	Bbsid            string `json:"bbsid"`
	LockReply        int    `json:"lock_reply"`
	CreateTime       int64  `json:"create_time"`
	CreaterPuid      int    `json:"createrPuid"`
	IsRtf            int    `json:"isRtf"`
	ReplyCount       int    `json:"reply_count"`
	CreaterName      string `json:"createrName"`
	CreaterId        int    `json:"creater_id"`
	FolderId         int    `json:"folderId"`
	Tags             string `json:"tags"`
	TopicType        int    `json:"topicType"`
	CodeContent      string `json:"code_content"`
	SReadcount       string `json:"s_readcount"`
	PraiseCount1     int    `json:"praise_count"`
	LastReplyContent string `json:"last_reply_content"`
	LastReplyId      int64  `json:"last_reply_id"`
	CircleId         int    `json:"circleId"`
	ReadPersonCount  int    `json:"readPersonCount"`
	Choice           int    `json:"choice"`
}

// 拉取讨论任务点信息
func PullBbsInfoAction(cache *xuexitong.XueXiTUserCache, p *entity.PointBBsDto) (*BBsTopic, error) {
	utEnc, err2 := cache.PullUtEnc(p.CourseID, p.ClassID, fmt.Sprintf("%d", p.KnowledgeID), p.Enc)
	if err2.Error() == "触发验证码" {
		log2.Print(log2.DEBUG, utils.RunFuncName(), "触发验证码，正在进行AI智能识别绕过.....")
		for {
			codePath, err1 := cache.XueXiTVerificationCodeApi(5, nil)
			if err1 != nil {
				return nil, err1
			}
			if codePath == "" { //如果path为空，那么可能是账号问题
				return nil, errors.New("无法正常获取对应网站验证码，请检查对应url是否正常")
			}
			img, _ := utils.ReadImg(codePath) //读取验证码图片
			//codeResult := utils.AutoVerification(img, ort.NewShape(1, 23)) //自动识别
			codeResult := ddddocr.SemiOCRVerification(img, ort.NewShape(1, 23))
			utils.DeleteFile(codePath) //删除验证码文件
			status, err1 := cache.XueXiTPassVerificationCode(codeResult, 5, nil)
			//fmt.Println(codeResult)
			//fmt.Println(status)
			if status {
				break
			}
		}
		utEnc, err2 = cache.PullUtEnc(p.CourseID, p.ClassID, fmt.Sprintf("%d", p.KnowledgeID), p.Enc)
	}
	if err2 != nil {
		return nil, err2
	}
	//fmt.Println(utEnc)
	id1, id2, err := cache.PullBbsCircleIdApi(p.Mid, p.JobID, false, fmt.Sprintf("%d", p.KnowledgeID), "s", p.ClassID, p.Enc, utEnc, p.CourseID, p.IsJob)
	if err != nil {
		return nil, err
	}
	//fmt.Println(id1, id2)
	contentHtml, err2 := cache.PullBbsInfoApi(id1, id2, p.CourseID, p.ClassID, 3, nil)
	if err2 != nil {
		return nil, err2
	}
	bbsTopic := &BBsTopic{}

	var topicStr string
	topicCompile := regexp.MustCompile(`topic:([\w\W]+?)},[^c]+course:\{`)
	topicSubmatch := topicCompile.FindStringSubmatch(contentHtml)
	if len(topicSubmatch) > 1 {
		topicStr = topicSubmatch[1] + "}"
	}

	err2 = json.Unmarshal([]byte(topicStr), bbsTopic)
	if err2 != nil {
		return nil, err2
	}

	//截取urlToken
	urlTokenCompile := regexp.MustCompile(`urlToken:'([^']+?)'`)
	urlTokenSubmatch := urlTokenCompile.FindStringSubmatch(contentHtml)
	if len(urlTokenSubmatch) > 1 {
		bbsTopic.UrlToken = urlTokenSubmatch[1]
	}
	return bbsTopic, err2
}

// AI回复讨论
func (bbsTopic *BBsTopic) AIAnswer(cache *xuexitong.XueXiTUserCache, p *entity.PointBBsDto, aiUrl, model string, aiType ctype.AiType, apiKey string) (string, error) {
	que := entity.EssayQue{
		Type:         qtype.Essay,
		OpFromAnswer: make(map[string][]string),
	}
	que.Text = bbsTopic.Title + "\n" + bbsTopic.Content //将题目数据加入到题目中

	message := xuexitong2.AIProblemMessage(bbsTopic.Title, que.Type.String(), entity.ExamTurn{
		XueXEssayQue: que,
	})
	que.AnswerAIGet("", aiUrl, model, aiType, message, apiKey)
	for _, answer := range que.OpFromAnswer {
		answerResult, err := cache.AnswerBbsApi(bbsTopic.Uuid, p.CourseID, p.ClassID, answer[0], bbsTopic.UrlToken, bbsTopic.Bbsid, 3, nil)
		if err != nil {
			return "", err
		}
		statusJson := gojsonq.New().JSONString(answerResult).Find("status")
		if status, ok := statusJson.(bool); ok {
			if status {
				return answerResult, nil
			} else {
				return "", errors.New(answerResult)
			}
		} else {
			return "", errors.New(answerResult)
		}
	}
	return "", errors.New("AI未找到回复内容如")
}

// 测试用的讨论任务点函数
func ExecuteBbsTest(cache *xuexitong.XueXiTUserCache, p *entity.PointBBsDto, setting config.AiSetting) {
	bbsTopic, err := PullBbsInfoAction(cache, p)
	if err != nil {
		fmt.Println(err)
	}
	//AI回复讨论
	answer, err := bbsTopic.AIAnswer(cache, p, setting.AiUrl, setting.Model, setting.AiType, setting.APIKEY)
	if err != nil {
		fmt.Println(err)
	}
	status := gojsonq.New().JSONString(answer).Find("msg")
	log.Printf("ID.%d(%s)讨论任务点完成状态：%s\n", p.KnowledgeID, p.Title, status.(string))
}
