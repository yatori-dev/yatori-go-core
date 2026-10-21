package haiqikeji

// ============================================================================
// 海旗科技（HQKJ）自动答题（作业 + 考试）核心逻辑
//
// 提供两种答案来源：
//   - HqkjAutoAnswerAction：优先使用平台内置标答，缺失时调用 AI 补答；
//   - HqkjAutoFillBuiltinAnswerAction：仅使用平台内置标答自动填写提交，不依赖 AI。
//
// 本文件原先位于 yatori-go-console 的 logic/haiqikeji/HqkjWorkExamAction.go，
// 现按框架分层迁移到核心库，console 层只保留薄调用。整体仿照英华(yinghua)
// StartWorkAction / StartExamAction 的范式：
//
//   work_list / exam_list（HqkjApi.go 已有）→ 拿 workId / examId
//        └→ work_detail / exam_detail → 取 paperId（部分作业绑定固定试卷，start 时需带上）
//        └→ work_start / exam_start（HqkjWorkExamApi.go）→ 题面 + 选项
//             └→ qentity.Question → aiq 算答案 → qutils 匹配回选项 idx → SubmitWork/ExamAnswerApi 提交
//
// ⚠️ 以下为「运行时假设」，已用日志暴露，实跑一次即可校准：
//   1. 提交答案的 recordId 采用 work_start/exam_start 返回中的 recordId；
//   2. 题型编码 mapHqkjType 目前仅确认 type=1 为单选，其余为推测；
//   3. 逐题 AnswerApi 提交后即落库，考试需再调 yee_exam_answer_finish 才算交卷；
//   4. consult_list 需该作业/考试存在答题记录方能返回题目（首次可能需网页进入一次）。
// ============================================================================

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/config"
	hqkjApi "github.com/yatori-dev/yatori-go-core/internal/api/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
)

// HqkjAnswerSetting 答题运行设置（AI 与内置标答模式共用）。
// 由上层（console）根据自己的配置结构组装后传入，核心库不直接依赖上层 config。
type HqkjAnswerSetting struct {
	AccountType    string           // 平台展示名（日志前缀，如“海旗科技”）
	Account        string           // 账号展示名（可传备注名）
	AutoExam       int              // 0=不答题，1=AI答题（内置标答模式忽略此字段）
	ExamAutoSubmit int              // 0=只答题不交卷，1=自动交卷
	AiSetting      config.AiSetting // AI 配置（内置标答模式忽略）
}

// HqkjQuizAnswerResult 单个作业/考试的答题结果，供外部对接直接取用。
type HqkjQuizAnswerResult struct {
	Kind      string `json:"kind"`   // work / exam
	QuizId    string `json:"quizId"` // workId / examId
	Title     string `json:"title"`
	Total     int    `json:"total"`     // 拉到的题目数
	Submitted int    `json:"submitted"` // 成功提交数
	Skipped   int    `json:"skipped"`   // 没有内置标答而跳过数（仅内置标答模式）
	Failed    int    `json:"failed"`    // 提交失败数
	Finished  bool   `json:"finished"`  // 考试是否已交卷
	Message   string `json:"message,omitempty"`
}

// hqkjAnswerSource 答案来源。
type hqkjAnswerSource int

const (
	hqkjAnswerByAI hqkjAnswerSource = iota
	hqkjAnswerByBuiltin
)

func hqkjAnswerTag(setting HqkjAnswerSetting, source hqkjAnswerSource) string {
	if source == hqkjAnswerByBuiltin {
		return "<内置标答>"
	}
	return fmt.Sprintf("<%s>", setting.AiSetting.AiType)
}

// hqkjQuiz 作业/考试条目（仅取答题所需字段）
type hqkjQuiz struct {
	Id            string // workId 或 examId
	Title         string
	Frequency     int     // 已作答/开考次数（exam_list 的 frequency；开考后会自增，不能单独作为完成依据）
	Status        int     // 平台状态：考试 examStatus（0未开始 1进行中 2已完成），作业 workStatus
	AchievedScore float64 // 实际得分（从 consult_list 的 studentResult[0].score 获取，仅已完成时有效）
}

// completed 判断作业/考试是否已经完成。
// 注意：海旗 exam_start 之后 frequency 会从 0 变 1，但 examStatus 仍为 1（进行中），
// 因此不能只用 frequency>0 判断完成，否则会漏掉可继续作答的考试。
func (q hqkjQuiz) completed() bool {
	switch q.Status {
	case 2, 3: // 已完成/已交卷/已批改
		return true
	case 1: // 进行中
		return false
	}
	// 平台未返回状态时（部分作业列表）沿用 frequency>0 的旧判断。
	return q.Frequency > 0
}

// hqkjTopic 单道题目
type hqkjTopic struct {
	TopicId    string           // 题目 id（提交用）
	RecordId   string           // work_start / exam_start 的整卷记录 id
	WrId       string           // consult 的作答记录 id
	WaId       string           // consult 的单题答案槽 id
	CateBid    string           // consult 题目大分类 id
	CateMid    string           // consult 题目中分类 id
	TypeInt    int              // 海旗原始题型编码（提交用）
	Question   qentity.Question // 交给 AI 的标准题目结构
	OptionIdx  []string         // 与 Question.Options 一一对应的选项 idx（A/B/C/D）
	CorrectIdx []string         // scale>=100 的选项 idx（平台返回的标准答案标记；为空时回退 AI）
}

// HqkjAutoAnswerAction 海旗 AI 自动答题总入口（作业 + 考试）。
// 在视频处理之后调用；仅当 setting.AutoExam==1（AI 答题）时生效。
func HqkjAutoAnswerAction(cache *hqkjApi.HqkjUserCache, course HqkjCourse, nodeList []HqkjNode, setting HqkjAnswerSetting) {
	if setting.AutoExam != 1 { // 0=不答题，1=AI答题（海旗目前仅接入 AI 答题）
		return
	}
	// 检测 AI 可用性：不可用则跳过答题，但不影响已完成的视频刷课（故不退出进程）
	if err := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType); err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), lg.BoldRed, fmt.Sprintf("<%s>", setting.AiSetting.AiType), "AI不可用，已跳过该课程AI答题，错误信息："+err.Error())
		return
	}
	_ = hqkjRunAnswerAction(cache, course, nodeList, setting, hqkjAnswerByAI)
}

// HqkjAutoFillBuiltinAnswerAction 仅使用海旗拉题返回的内置标准答案（option.scale>=100）
// 自动填写并提交，不调用 AI、不依赖 AI 配置，供外部直接对接。
//
// 说明：
//   - 作业、考试都会处理；同一 workId/examId 自动去重。
//   - 已完成的作业/考试会跳过，避免重复提交；进行中的考试可继续作答。
//   - 考试是否交卷由 setting.ExamAutoSubmit 控制（1=自动交卷）。
//   - 没有内置标答的题目会跳过并计入 Skipped；返回结果可直接序列化给调用方。
func HqkjAutoFillBuiltinAnswerAction(cache *hqkjApi.HqkjUserCache, course HqkjCourse, nodeList []HqkjNode, setting HqkjAnswerSetting) []HqkjQuizAnswerResult {
	if setting.AccountType == "" {
		setting.AccountType = "海旗科技"
	}
	return hqkjRunAnswerAction(cache, course, nodeList, setting, hqkjAnswerByBuiltin)
}

func hqkjRunAnswerAction(cache *hqkjApi.HqkjUserCache, course HqkjCourse, nodeList []HqkjNode, setting HqkjAnswerSetting, source hqkjAnswerSource) []HqkjQuizAnswerResult {
	seen := map[string]bool{} // 同一作业可能挂在多个 node 下被重复拉取，按 workId 去重，避免重复作答/提交
	results := make([]HqkjQuizAnswerResult, 0)
	for _, node := range nodeList {
		results = append(results, workAction(cache, course, node, setting, source, seen)...)
		results = append(results, examAction(cache, course, node, setting, source, seen)...)
	}
	return results
}

// workAction 作业答题。seen 按 workId 去重，避免同一作业在多个 node 下被重复作答。
func workAction(cache *hqkjApi.HqkjUserCache, course HqkjCourse, node HqkjNode, setting HqkjAnswerSetting, source hqkjAnswerSource, seen map[string]bool) []HqkjQuizAnswerResult {
	results := make([]HqkjQuizAnswerResult, 0)
	if node.TabWork <= 0 { // 过滤非作业节点
		return results
	}
	listResult, err := cache.PullWorkInfoApi(course.Id, node.Id, 10, nil)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "拉取作业列表失败：", err.Error())
		return append(results, HqkjQuizAnswerResult{Kind: "work", Message: "拉取作业列表失败：" + err.Error()})
	}
	actionName := "正在AI自动写作业："
	if source == hqkjAnswerByBuiltin {
		actionName = "正在使用内置标答自动写作业："
	}
	for _, q := range parseQuizList(listResult, "workInfo", "workId") {
		if seen[q.Id] { // 已处理过的作业跳过
			continue
		}
		seen[q.Id] = true
		// 已完成的作业调 consult_list 获取实际得分
		if q.completed() {
			if consultResult, cerr := pullConsultApi(cache, "work", q.Id, course.Id); cerr == nil {
				q.AchievedScore = parseConsultScore(consultResult)
			}
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", hqkjAnswerTag(setting, source), "【", course.Name, "】", "【", node.Name, "】 ", lg.Yellow, actionName, q.Title)
		results = append(results, answerQuiz(cache, course, node, setting, source, "work", q))
	}
	return results
}

// examAction 考试答题。seen 按 examId 去重（加 exam: 前缀，避免与作业 workId 数值碰撞）。
func examAction(cache *hqkjApi.HqkjUserCache, course HqkjCourse, node HqkjNode, setting HqkjAnswerSetting, source hqkjAnswerSource, seen map[string]bool) []HqkjQuizAnswerResult {
	results := make([]HqkjQuizAnswerResult, 0)
	if node.TabExam <= 0 { // 过滤非考试节点
		return results
	}
	listResult, err := cache.PullExamListApi(course.Id, node.Id, 10, nil)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "拉取考试列表失败：", err.Error())
		return append(results, HqkjQuizAnswerResult{Kind: "exam", Message: "拉取考试列表失败：" + err.Error()})
	}
	actionName := "正在AI自动考试："
	if source == hqkjAnswerByBuiltin {
		actionName = "正在使用内置标答自动考试："
	}
	for _, q := range parseQuizList(listResult, "examInfo", "examId") {
		if seen["exam:"+q.Id] { // 加 exam: 前缀，避免与作业 workId 数值碰撞被误去重
			continue
		}
		seen["exam:"+q.Id] = true
		// 已完成的考试调 consult_list 获取实际得分
		if q.completed() {
			if consultResult, cerr := pullConsultApi(cache, "exam", q.Id, course.Id); cerr == nil {
				q.AchievedScore = parseConsultScore(consultResult)
			}
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", hqkjAnswerTag(setting, source), "【", course.Name, "】", "【", node.Name, "】 ", lg.Yellow, actionName, q.Title, " (examId=", q.Id, " frequency=", strconv.Itoa(q.Frequency), ")")
		results = append(results, answerQuiz(cache, course, node, setting, source, "exam", q))
	}
	return results
}

// answerQuiz 对单个作业/考试执行：拉题 → 取内置标答或 AI 作答 → 逐题提交
func answerQuiz(cache *hqkjApi.HqkjUserCache, course HqkjCourse, node HqkjNode, setting HqkjAnswerSetting, source hqkjAnswerSource, kind string, quiz hqkjQuiz) HqkjQuizAnswerResult {
	res := HqkjQuizAnswerResult{Kind: kind, QuizId: quiz.Id, Title: quiz.Title}

	// 已完成的作业/考试直接跳过，避免再调 start 造成无意义的 500 或重复提交。
	if quiz.completed() {
		kindStr := "作业"
		if kind == "exam" {
			kindStr = "考试"
		}
		if quiz.AchievedScore > 0 {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", "【", node.Name, "】", "【", quiz.Title, "】 ", lg.Yellow, kindStr, "已完成，不可再次提交（已作答 ", strconv.Itoa(quiz.Frequency), " 次，得分 ", fmt.Sprintf("%.0f", quiz.AchievedScore), " 分）")
			res.Message = fmt.Sprintf("已完成，已作答 %d 次，得分 %.0f 分", quiz.Frequency, quiz.AchievedScore)
		} else {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", "【", node.Name, "】", "【", quiz.Title, "】 ", lg.Yellow, kindStr, "已完成，不可再次提交（已作答 ", strconv.Itoa(quiz.Frequency), " 次）")
			res.Message = fmt.Sprintf("已完成，已作答 %d 次", quiz.Frequency)
		}
		return res
	}

	// 第一步：拉详情，取记录列表（含每条记录的 id / paperId / classId）
	detailResult, derr := pullDetailApi(cache, kind, quiz.Id, course.Id, quiz.Title)
	if derr != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "拉取作业详情失败：", derr.Error())
		res.Message = "拉取详情失败：" + derr.Error()
		return res
	}
	records := parseDetailRecords(detailResult)
	var paperId, classId string
	for _, r := range records {
		if classId == "" {
			classId = r.ClassId
		}
		if paperId == "" {
			paperId = r.PaperId
		}
	}
	// 第二步：拉题。统一走 work_start / exam_start（正常作答流程：它会创建可写作答会话并返回 recordId）。
	// paperId=0 的老式作业显式传 "0" 试着建会话（传空会被海旗拒"缺少必要参数: paperId"）。
	startPaperId := paperId
	if startPaperId == "" {
		startPaperId = "0"
	}
	rawResp, err := pullStartApi(cache, kind, quiz.Id, course.Id, startPaperId, classId)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "拉取题目失败：", err.Error())
		res.Message = "拉取题目失败：" + err.Error()
		return res
	}
	topics := parseStartQuestions(rawResp)
	// work_start 没出题时（老式作业可能不吃 paperId=0），回退 consult_list 拉题
	if len(topics) == 0 {
		for _, r := range records {
			consultResult, cerr := pullConsultApi(cache, kind, r.Id, course.Id)
			if cerr != nil {
				continue
			}
			rawResp = consultResult
			if t2 := parseStartQuestions(consultResult); len(t2) > 0 {
				topics = t2
				break
			}
		}
	}
	if len(topics) == 0 {
		// 没拉到题：智能判断原因（未开始/已完成/已截止）并打印人性化提示
		var startTime, endTime string
		var frequency int
		for _, r := range records {
			if r.StartTime != "" {
				startTime = r.StartTime
			}
			if r.EndTime != "" {
				endTime = r.EndTime
			}
		}
		frequency = quiz.Frequency // 从 exam_list 继承的 frequency
		now := time.Now()
		var reason string
		// 优先判断完成状态。frequency>0 也可能只是开考过（进行中），不能直接当成已完成。
		if quiz.completed() {
			if quiz.AchievedScore > 0 {
				reason = fmt.Sprintf("已完成，不可再次提交（已作答 %d 次，得分 %.0f 分）", frequency, quiz.AchievedScore)
			} else {
				reason = fmt.Sprintf("已完成，不可再次提交（已作答 %d 次）", frequency)
			}
		} else if frequency > 0 {
			reason = fmt.Sprintf("已有作答记录但尚未完成（frequency=%d），可继续作答，请稍后重试", frequency)
		}
		// 次优判断开始/截止时间（使用本地时区解析，避免 UTC 时区误判）
		if reason == "" && startTime != "" {
			st, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
			if err == nil && now.Before(st) {
				reason = fmt.Sprintf("尚未开始，开始时间：%s", startTime)
			}
		}
		if reason == "" && endTime != "" {
			et, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
			if err == nil && now.After(et) {
				reason = fmt.Sprintf("已截止，截止时间：%s", endTime)
			}
		}
		// 兜底：拉不到题但时间正常、frequency=0 → 可能是服务端其他限制
		if reason == "" {
			reason = fmt.Sprintf("未拉到客观题（frequency=%d, 当前时间=%s, 开始时间=%s, 截止时间=%s）", frequency, now.Format("2006-01-02 15:04:05"), startTime, endTime)
		}
		kindStr := "作业"
		if kind == "exam" {
			kindStr = "考试"
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", "【", node.Name, "】", "【", quiz.Title, "】 ", lg.Yellow, kindStr, reason)
		res.Message = reason
		return res
	}
	res.Total = len(topics)
	for _, t := range topics {
		// 答案来源：开考接口若已带标答(scale=100→CorrectIdx)则直接采用、跳过 AI；
		// 内置标答模式下没有标答的题目直接跳过，不调用 AI。
		answers := hqkjBuiltinAnswers(t)
		if len(answers) > 0 {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Green, "采用平台内置标答(scale=100)：topicId=", t.TopicId, " 答案=", fmt.Sprintf("%v", answers))
		} else if source == hqkjAnswerByBuiltin {
			res.Skipped++
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Yellow, "题目 ", t.TopicId, " 没有内置标答，已跳过")
			continue
		} else {
			answers = aiAnswerTopic(setting, t)
		}
		var subResult string
		if kind == "work" { // 作业走 work_answer_add（8 字段不带 wrId/waId）
			subResult, err = cache.SubmitWorkAnswerApi(course.Id, quiz.Id, t.TopicId, t.RecordId, strconv.Itoa(t.TypeInt), answers, 10, nil)
		} else { // 考试走 exam_answer_add（同样 8 字段，不带 wrId/waId）
			if examDryRun { // 【考试 dry-run】只打印将提交的 payload，绝不真正提交；有开放考试核对无误后把 examDryRun 改为 false 放开
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ",
					lg.Yellow, "【考试DRY-RUN·未提交】", examAnswerPayload(cache, course.Id, quiz.Id, t.TopicId, t.RecordId, strconv.Itoa(t.TypeInt), answers))
				continue
			}
			subResult, err = cache.SubmitExamAnswerApi(course.Id, quiz.Id, t.TopicId, t.RecordId, strconv.Itoa(t.TypeInt), answers, 10, nil)
		}
		if err != nil {
			res.Failed++
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "提交答案异常：", err.Error())
			continue
		}
		if code, ok := gojsonq.New().JSONString(subResult).Find("code").(float64); ok && int(code) == 200 {
			res.Submitted++
			kindStr := "作业"
			if kind == "exam" {
				kindStr = "考试"
			}
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Green, kindStr, "题目 ", t.TopicId, " 提交成功（", strconv.Itoa(res.Submitted), "/", strconv.Itoa(len(topics)), "）")
		} else {
			res.Failed++
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Red, "提交答案返回异常：", subResult, " workId：", quiz.Id, " topicId：", t.TopicId, " wrId：", t.WrId, " waId：", t.WaId, " 题型：", strconv.Itoa(t.TypeInt), " 答案：", fmt.Sprintf("%v", answers))
		}
	}
	doneName := "AI答题"
	if source == hqkjAnswerByBuiltin {
		doneName = "内置标答答题"
	}
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Green, doneName, "完毕，成功提交 ", strconv.Itoa(res.Submitted), "/", strconv.Itoa(res.Total), " 题，跳过 ", strconv.Itoa(res.Skipped), " 题，失败 ", strconv.Itoa(res.Failed), " 题")

	// 考试交卷：逐题 yee_exam_answer_add 仅暂存，需 yee_exam_answer_finish(state=2) 才正式交卷出成绩。
	// 双重保险：仅 kind==exam 且 setting.ExamAutoSubmit==1 且至少成功提交一题才交卷；examDryRun 下只打印不真正交卷。
	if kind == "exam" && setting.ExamAutoSubmit == 1 && res.Submitted > 0 {
		recordId := topics[0].RecordId // 整卷同一 recordId（开考接口返回）
		if examDryRun {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Yellow, "【考试DRY-RUN】跳过交卷(yee_exam_answer_finish)，recordId=", recordId)
		} else {
			finRes, ferr := cache.FinishExamAnswerApi(course.Id, quiz.Id, recordId, 10, nil)
			if ferr != nil {
				res.Message = "考试交卷异常：" + ferr.Error()
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "考试交卷异常：", ferr.Error())
			} else if code, ok := gojsonq.New().JSONString(finRes).Find("code").(float64); ok && int(code) == 200 {
				res.Finished = true
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Green, "考试已正式交卷(state=2)，examId=", quiz.Id)
			} else {
				res.Message = "考试交卷返回异常：" + finRes
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.Red, "考试交卷返回异常：", finRes)
			}
		}
	}
	return res
}

// hqkjBuiltinAnswers 取海旗题目自带的标准答案，并归一化为提交格式（选项 idx）。
// 单选/判断题只取第一个正确项，多选题保留全部正确项并去重。
func hqkjBuiltinAnswers(t hqkjTopic) []string {
	if len(t.CorrectIdx) == 0 {
		return nil
	}
	qt := qtype.Index(t.Question.Type)
	if qt == qtype.SingleChoice || qt == qtype.TrueOrFalse {
		if t.CorrectIdx[0] == "" {
			return nil
		}
		return []string{t.CorrectIdx[0]}
	}

	seen := make(map[string]bool, len(t.CorrectIdx))
	answers := make([]string, 0, len(t.CorrectIdx))
	for _, idx := range t.CorrectIdx {
		if idx == "" || seen[idx] {
			continue
		}
		seen[idx] = true
		answers = append(answers, idx)
	}
	return answers
}

// aiAnswerTopic 调 AI 作答单题，并把答案转换为提交格式（选择/判断→选项 idx；填空/简答→文本）
func aiAnswerTopic(setting HqkjAnswerSetting, t hqkjTopic) []string {
	aiAnswer, err := aiq.AggregationAIApi(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.AiType, aiq.BuildAiQuestionMessage(t.Question), setting.AiSetting.APIKEY)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "AI回答异常：", err.Error())
	}
	var contents []string
	_ = json.Unmarshal([]byte(aiAnswer), &contents)

	qt := qtype.Index(t.Question.Type)
	var answers []string
	switch qt {
	case qtype.SingleChoice, qtype.MultipleChoice, qtype.TrueOrFalse:
		for _, c := range contents {
			if idx := matchOptionIdx(c, t); idx != "" {
				answers = append(answers, idx)
			}
		}
	default: // 填空、简答、名词解释、论述等：直接用文本作为答案
		answers = append(answers, contents...)
	}

	// 兜底：AI 无法解析时给默认答案，避免漏题（仿英华策略）
	if len(answers) == 0 {
		switch qt {
		case qtype.SingleChoice, qtype.TrueOrFalse:
			if len(t.OptionIdx) > 0 {
				answers = []string{t.OptionIdx[0]}
			}
		case qtype.MultipleChoice:
			if len(t.OptionIdx) >= 2 {
				answers = []string{t.OptionIdx[0], t.OptionIdx[1]}
			} else if len(t.OptionIdx) > 0 {
				answers = []string{t.OptionIdx[0]}
			}
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", setting.AccountType), "[", lg.Green, setting.Account, lg.Default, "] ", lg.BoldRed, "AI回答无法解析，题型：", t.Question.Type, "，已采用兜底答案。AI原始返回：", aiAnswer)
	}
	return answers
}

// matchOptionIdx 把 AI 返回的「选项内容」匹配回海旗选项 idx（A/B/C/D）。
// 用编辑距离相似度找最匹配的选项下标，再取其真实 idx（避免依赖 qutils 内部字母表顺序）。
func matchOptionIdx(content string, t hqkjTopic) string {
	if len(t.Question.Options) == 0 {
		return ""
	}
	best, bestScore := 0, -1.0
	for i, opt := range t.Question.Options {
		if s := qutils.Similarity(opt, content); s > bestScore {
			bestScore = s
			best = i
		}
	}
	if best < len(t.OptionIdx) {
		return t.OptionIdx[best]
	}
	return ""
}

// mapHqkjType 海旗题型编码(int) → qtype 中文题型（用于构造喂给 AI 的题目）。
// ⚠️假设2：目前仅确认 type=1 为单选；2/3/4/5 为按常见编码的推测，遇未知类型默认按单选处理。
func mapHqkjType(t int) string {
	switch t {
	case 1:
		return qtype.SingleChoice.String() // 单选题
	case 2:
		return qtype.MultipleChoice.String() // 多选题
	case 3:
		return qtype.TrueOrFalse.String() // 判断题
	case 4:
		return qtype.FillInTheBlank.String() // 填空题
	case 5:
		return qtype.ShortAnswer.String() // 简答题
	default:
		return qtype.SingleChoice.String()
	}
}

// parseQuizList 解析作业/考试列表，提取 id 与标题。
// 作业：arrKey=workInfo、idKey=workId；考试：arrKey=examInfo、idKey=examId。
func parseQuizList(jsonStr, arrKey, idKey string) []hqkjQuiz {
	var res []hqkjQuiz
	arr, ok := gojsonq.New().JSONString(jsonStr).Find("data." + arrKey).([]any)
	if !ok {
		return res
	}
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		var q hqkjQuiz
		if v, ok := m[idKey].(float64); ok {
			q.Id = strconv.Itoa(int(v))
		}
		if v, ok := m["title"].(string); ok {
			q.Title = v
		}
		if v, ok := m["frequency"].(float64); ok { // 已作答次数（考试列表特有；作业列表无此字段则为 0）
			q.Frequency = int(v)
		}
		// 考试状态：0未开始 1进行中 2已完成；作业列表字段名为 workStatus。
		if v, ok := m["examStatus"].(float64); ok {
			q.Status = int(v)
		}
		if v, ok := m["workStatus"].(float64); ok {
			q.Status = int(v)
		}
		// 列表里的 score：未完成时通常是试卷总分，完成后为实际得分，作为 consult 的兜底。
		if v, ok := m["score"].(float64); ok {
			q.AchievedScore = v
		}
		if q.Id != "" {
			res = append(res, q)
		}
	}
	return res
}

// parseStartQuestions 解析 work_start/exam_start 返回的题目。
// 实测题目数组为 data.workTopics[]，每题字段：id(题目id→提交用 topicId)/recordId(整卷记录id→提交用)/
// type(题型 1单选 3判断…)/topic(题面,含HTML标签)/option[]{idx,answer,scale}。
// scale(=100 即正确答案标记)会被收集进 CorrectIdx，由 answerQuiz 优先采用（拿不到时回退 AI 答题）。
func parseStartQuestions(jsonStr string) []hqkjTopic {
	var topics []hqkjTopic
	var arr []any
	for _, key := range []string{"data.workTopics", "data.examTopics", "data.workResult"} {
		if v, ok := gojsonq.New().JSONString(jsonStr).Find(key).([]any); ok {
			arr = v
			break
		}
	}
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		var t hqkjTopic
		if v, ok := m["id"].(float64); ok { // work_start: 题目 id → 提交用 topicId
			t.TopicId = strconv.Itoa(int(v))
		}
		if v, ok := m["topicId"].(float64); ok { // consult_list: topicId → 提交用 topicId
			t.TopicId = strconv.Itoa(int(v))
		}
		if v, ok := m["recordId"].(float64); ok { // work_start: 整卷记录 id
			t.RecordId = strconv.Itoa(int(v))
		}
		if v, ok := m["wrId"].(float64); ok { // consult_list: 作答记录 id
			t.WrId = strconv.Itoa(int(v))
		}
		if v, ok := m["waId"].(float64); ok { // consult_list: 单题答案槽 id
			t.WaId = strconv.Itoa(int(v))
		}
		if v, ok := m["cateBid"].(float64); ok { // consult_list: 题目大分类 id
			t.CateBid = strconv.Itoa(int(v))
		}
		if v, ok := m["cateMid"].(float64); ok { // consult_list: 题目中分类 id
			t.CateMid = strconv.Itoa(int(v))
		}
		if v, ok := m["type"].(float64); ok {
			t.TypeInt = int(v)
		}
		if v, ok := m["topic"].(string); ok {
			t.Question.Content = stripHTML(v)
		}
		t.Question.Type = mapHqkjType(t.TypeInt)
		if opts, ok := m["option"].([]any); ok {
			for _, o := range opts {
				om, ok := o.(map[string]any)
				if !ok {
					continue
				}
				ans, _ := om["answer"].(string)
				idx, _ := om["idx"].(string)
				t.Question.Options = append(t.Question.Options, ans)
				t.OptionIdx = append(t.OptionIdx, idx)
				// 平台在 option.scale 标注正确项：scale>=100 即满分正确答案。
				// 收集所有满分项 idx 作为标准答案（多选可能多个），供 answerQuiz 优先采用、跳过 AI。
				if scale, ok := om["scale"].(float64); ok && scale >= 100 && idx != "" {
					t.CorrectIdx = append(t.CorrectIdx, idx)
				}
			}
		}
		if t.TopicId != "" && len(t.Question.Options) > 0 {
			topics = append(topics, t)
		}
	}
	return topics
}

// htmlTagRe 匹配 HTML 标签，用于清洗题面。
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// stripHTML 去掉题面里的 HTML 标签与零宽/实体字符，得到纯文本喂给 AI。
func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "Feff", "")
	return strings.TrimSpace(s)
}

// pullDetailApi 发起作业/考试「详情(拉题)」查询：work_detail / exam_detail。
// 注意：consult_list 是「查看已答记录」接口（未作答返回空 workResult），真正拉题需用 detail 接口。
func pullDetailApi(cache *hqkjApi.HqkjUserCache, kind, quizId, courseId, title string) (string, error) {
	if kind == "work" {
		return cache.PullWorkDetailApi(quizId, courseId, title, 10, nil)
	}
	return cache.PullExamDetailApi(quizId, courseId, title, 10, nil)
}

// pullStartApi 调海旗「开始作业/考试」接口正式拉题（work_start / exam_start）。
// paperId=0 的老式作业也显式传 "0"（传空会被拒"缺少必要参数: paperId"）。
func pullStartApi(cache *hqkjApi.HqkjUserCache, kind, quizId, courseId, paperId, classId string) (string, error) {
	if kind == "work" {
		return cache.PullWorkStartApi(quizId, courseId, paperId, classId, 10, nil)
	}
	return cache.PullExamStartApi(quizId, courseId, paperId, classId, 10, nil)
}

// hqkjRecord work_detail/exam_detail 里的一条答题记录。
type hqkjRecord struct {
	Id        string // 记录 id（paperId=0 的老式作业用它走 consult_list 拉题）
	PaperId   string // 固定试卷 id（>0 时走 work_start 拉题）
	ClassId   string // 班级 id（work_start 必填）
	StartTime string // 开始时间（格式 "2026-06-14 00:00:00"）
	EndTime   string // 截止时间（格式 "2026-06-21 23:55:55"）
	Frequency int    // 已作答次数（从 exam_list 的 frequency 继承，detail 里无此字段）
}

// parseDetailRecords 解析 work_detail/exam_detail 的记录列表（id / paperId>0 / classId>0）。
func parseDetailRecords(jsonStr string) []hqkjRecord {
	var recs []hqkjRecord
	for _, arrKey := range []string{"data.workDetail", "data.examDetail"} {
		arr, ok := gojsonq.New().JSONString(jsonStr).Find(arrKey).([]any)
		if !ok {
			continue
		}
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			var r hqkjRecord
			if v, ok := m["id"].(float64); ok {
				r.Id = strconv.Itoa(int(v))
			}
			if v, ok := m["paperId"].(float64); ok && int(v) > 0 {
				r.PaperId = strconv.Itoa(int(v))
			}
			if v, ok := m["classId"].(float64); ok && int(v) > 0 {
				r.ClassId = strconv.Itoa(int(v))
			}
			if v, ok := m["startTime"].(string); ok {
				r.StartTime = v
			}
			if v, ok := m["endTime"].(string); ok {
				r.EndTime = v
			}
			if r.Id != "" {
				recs = append(recs, r)
			}
		}
	}
	return recs
}

// pullConsultApi 调海旗「答题记录查看」接口 consult_list 拉题（用于 paperId=0 的老式作业/考试）。
// 参数位上的 workId/examId 实际传 work_detail 里的记录 id；返回题目在 data.workResult[] 下。
func pullConsultApi(cache *hqkjApi.HqkjUserCache, kind, recordId, courseId string) (string, error) {
	if kind == "work" {
		return cache.PullWorkConsultApi(recordId, courseId, 10, nil)
	}
	return cache.PullExamConsultApi(recordId, courseId, 10, nil)
}

// examDryRun 考试专用 dry-run 开关：true=考试只打印将提交的 payload、绝不真正发请求（验证用）；false=真实提交。
// 作业已实测打通、恒真实提交，不受此开关影响。payload 与交卷接口已用抓包逐字节验证无误，故放开为 false 真实提交。
// 交卷(yee_exam_answer_finish)另受 setting.ExamAutoSubmit 控制（默认 0=只答题不交卷）。
const examDryRun = false

// examAnswerPayload 构造考试提交 payload。字段与真实抓包(yee_exam_answer_add)完全一致：
// {schoolId,courseId,userId,examId,topicId,answer,recordId,type}——考试不带 wrId/waId（那是作业 work_answer_add 才需要的）。
func examAnswerPayload(cache *hqkjApi.HqkjUserCache, courseId, examId, topicId, recordId, qType string, answers []string) string {
	if recordId == "" { // 兜底，避免拼出非法 JSON（正常 recordId 来自 exam_start 响应的 examTopics[].recordId）
		recordId = "0"
	}
	answersData, _ := json.Marshal(answers)
	return `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"examId":` + examId + `,"topicId":` + topicId + `,"answer":` + string(answersData) +
		`,"recordId":` + recordId + `,"type":` + qType + `}`
}

// parseConsultScore 从 consult_list 返回的 JSON 中提取实际得分（studentResult[0].score）。
func parseConsultScore(jsonStr string) float64 {
	scoreVal, ok := gojsonq.New().JSONString(jsonStr).Find("data.studentResult.[0].score").(float64)
	if !ok {
		return 0
	}
	return scoreVal
}
