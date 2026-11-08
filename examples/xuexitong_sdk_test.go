package examples

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yatori-dev/yatori-go-core/v3"
	"github.com/yatori-dev/yatori-go-core/v3/config"
	"github.com/yatori-dev/yatori-go-core/v3/global"
	"github.com/yatori-dev/yatori-go-core/v3/utils"
)

// xuexitongUserIndex 是 examples/TestData.yaml 里 users 列表的下标。
// 改成你自己的学习通账号即可（accountType 必须是 XUEXITONG）。
const xuexitongUserIndex = 78

// xuexitongCourseFilter 只处理课程名包含该关键字的课程；留空表示处理全部课程。
const xuexitongCourseFilter = ""

// xuexitongFastStudy 为 true 时视频直接提交 finish（不按普通模式步进观看）。
const xuexitongFastStudy = false

const (
	// 平台把看完的视频回报成 99.x%，留 1% 容差。
	xuexitongFinishedProgress = 99.0
	// 视频进度超过该值后只补剩余部分，不再走完整 start -> progress。
	xuexitongResumeProgress = 95.0
	// 普通模式每次推进的时长和两次提交之间的等待时间。
	xuexitongProgressStep     = 30 * time.Second
	xuexitongProgressInterval = 30 * time.Second
)

// TestXueXiTongAppSDKStudy 演示学习通 App 登录 -> 课程拉取 -> 章节树 -> 观看/学习。
//
// 账号来自 examples/TestData.yaml：
//
//	utils.YatoriCoreInit()
//	setup()
//	user := global.Config.Users[xuexitongUserIndex]
//
// App 登录会走本地 deviceInfo -> clientId -> cid/sc；服务端判定为新设备时
// 会触发短信验证，LoginCodeProvider 先让用户输入 1 确认，确认后才调用
// SendSMS 发送短信（发送前的图片验证码由 ddddocr 自动识别），再读取用户
// 收到的短信验证码。
//
// 运行：
//
//	go test ./examples -run TestXueXiTongAppSDKStudy -v
//	go test ./examples -run TestXueXiTongAppSDKStudy -v -short  # 只登录 + 拉课程
func TestXueXiTongAppSDKStudy(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := xuexitongTestUser(t, xuexitongUserIndex)

	client, err := yatori.New(yatori.Config{
		Platform: yatori.XueXiTong,
		Account: yatori.Account{
			Username: user.Account,
			Password: user.Password,
		},
		Options: yatori.Options{
			Timeout:           60 * time.Second,
			Logger:            xuexitongTestLogger{t: t},
			LoginMode:         yatori.LoginModeApp,
			LoginCodeProvider: xuexitongLoginCodeProvider(t),
		},
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// 普通模式下视频学习会持续很久，这里给 6 小时总时限。
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	t.Cleanup(cancel)

	t.Log("== 1/3 App 登录 ==")
	t.Log("如果服务端判定为新设备，会触发短信验证。SDK 不会自动发送短信：")
	t.Log("终端先提示输入 1，确认后才发送验证码；收到短信后再输入验证码。")
	if err := client.Login(ctx); err != nil {
		t.Fatalf("登录失败（App 登录失败时 SDK 会自动回退到 Web 登录）: %v", err)
	}
	session := client.Session()
	t.Logf("登录成功: account=%s userId=%s cookieLen=%d", session.Account, session.UserID, len(session.Cookie))

	t.Log("== 2/3 拉取课程 ==")
	courses := xuexitongCourses(t, ctx, client)

	if testing.Short() {
		t.Log("short 模式：只验证登录和课程拉取，跳过学习")
		return
	}

	xuexitongStudyCourses(t, ctx, client, courses)
}

// TestXueXiTongWebSDKStudy 演示学习通 Web 登录 -> 课程拉取 -> 章节树 -> 观看/学习。
//
// 与 App 登录示例共用账号来源和后续学习流程，区别只在 Options.LoginMode：
// Web 登录走 passport2 的 fanyalogin，不会返回 cid/sc。
//
// 运行：
//
//	go test ./examples -run TestXueXiTongWebSDKStudy -v
//	go test ./examples -run TestXueXiTongWebSDKStudy -v -short  # 只登录 + 拉课程
func TestXueXiTongWebSDKStudy(t *testing.T) {
	utils.YatoriCoreInit()
	setup()
	user := xuexitongTestUser(t, xuexitongUserIndex)

	client, err := yatori.New(yatori.Config{
		Platform: yatori.XueXiTong,
		Account: yatori.Account{
			Username: user.Account,
			Password: user.Password,
		},
		Options: yatori.Options{
			Timeout:           60 * time.Second,
			Logger:            xuexitongTestLogger{t: t},
			LoginMode:         yatori.LoginModeWeb,
			LoginCodeProvider: xuexitongLoginCodeProvider(t),
		},
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	// 普通模式下视频学习会持续很久，这里给 6 小时总时限。
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	t.Cleanup(cancel)

	t.Log("== 1/3 Web 登录 ==")
	if err := client.Login(ctx); err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	// 也可以复用浏览器里已有的会话：
	//   err = client.LoginWithCookie(ctx, "UID=...; vc3=...;")
	session := client.Session()
	t.Logf("登录成功: account=%s userId=%s cookieLen=%d", session.Account, session.UserID, len(session.Cookie))

	t.Log("== 2/3 拉取课程 ==")
	courses := xuexitongCourses(t, ctx, client)

	if testing.Short() {
		t.Log("short 模式：只验证登录和课程拉取，跳过学习")
		return
	}

	xuexitongStudyCourses(t, ctx, client, courses)
}

// xuexitongTestUser 从 global.Config.Users 里取学习通账号，并做基本校验。
func xuexitongTestUser(t *testing.T, index int) config.Users {
	t.Helper()
	if global.Config == nil {
		t.Fatal("global.Config 为空：请确认 examples/TestData.yaml 存在，且已执行 utils.YatoriCoreInit() + setup()")
	}
	if index < 0 || index >= len(global.Config.Users) {
		t.Fatalf("users 下标 %d 越界，TestData.yaml 里只有 %d 个账号", index, len(global.Config.Users))
	}
	user := global.Config.Users[index]
	if !strings.EqualFold(strings.TrimSpace(user.AccountType), "XUEXITONG") {
		t.Fatalf("users[%d].accountType = %q，不是 XUEXITONG", index, user.AccountType)
	}
	if strings.TrimSpace(user.Account) == "" || user.Password == "" {
		t.Fatalf("users[%d] 缺少 account/password", index)
	}
	return user
}

// xuexitongCourses 拉取课程列表并打印概要。
func xuexitongCourses(t *testing.T, ctx context.Context, client yatori.Client) []yatori.Course {
	t.Helper()
	courses, err := client.Courses(ctx)
	if err != nil {
		t.Fatalf("拉取课程失败: %v", err)
	}
	if len(courses) == 0 {
		t.Fatal("该账号没有可见课程")
	}
	t.Logf("共 %d 门课程:", len(courses))
	for _, course := range courses {
		t.Logf("  - [%s] %s（进度 %.1f%%，状态 %s）", course.ID, course.Name, course.Progress, course.Status)
	}
	return courses
}

// xuexitongStudyCourses 按课程过滤条件逐门学习。
func xuexitongStudyCourses(t *testing.T, ctx context.Context, client yatori.Client, courses []yatori.Course) {
	t.Helper()
	handled := 0
	for _, course := range courses {
		if xuexitongCourseFilter != "" && !strings.Contains(course.Name, xuexitongCourseFilter) {
			continue
		}
		handled++
		t.Logf("== 3/3 课程: %s ==", course.Name)
		xuexitongStudyCourse(t, ctx, client, course)
	}
	if handled == 0 {
		t.Logf("没有匹配 xuexitongCourseFilter=%q 的课程", xuexitongCourseFilter)
	}
}

// xuexitongStudyCourse 拉取章节树，并按深度优先顺序逐个学习叶子任务点。
func xuexitongStudyCourse(t *testing.T, ctx context.Context, client yatori.Client, course yatori.Course) {
	t.Helper()
	nodes, err := client.Tree(ctx, course)
	if err != nil {
		t.Errorf("课程 %s 拉取章节失败: %v", course.Name, err)
		return
	}
	leaves := yatori.Flatten(nodes)
	t.Logf("共 %d 个任务点", len(leaves))

	for i, node := range leaves {
		t.Logf("[%d/%d] %s（%s）", i+1, len(leaves), node.Name, node.Kind)
		xuexitongStudyNode(t, ctx, client, course, node)
	}
}

// xuexitongStudyNode 学习单个任务点：
//
//	视频/音频：start（拿真实时长）-> progress 步进 -> finish
//	文档/链接：一次 finish 提交
//	fast/已学大半：一次 finish 提交
func xuexitongStudyNode(t *testing.T, ctx context.Context, client yatori.Client, course yatori.Course, node yatori.Node) {
	t.Helper()
	switch node.Kind {
	case yatori.NodeChapter, yatori.NodeWork, yatori.NodeExam:
		t.Log("  非学习任务点，跳过（作业/考试走 Works/Exams）")
		return
	}
	if node.Completed || node.Progress >= xuexitongFinishedProgress {
		t.Logf("  已完成（%.1f%%），跳过", node.Progress)
		return
	}

	fast := xuexitongFastStudy
	mode := yatori.StudyModeNormal
	if fast {
		mode = yatori.StudyModeFast
	}

	isMedia := node.Kind == yatori.NodeVideo || node.Kind == yatori.NodeAudio
	resume := isMedia && node.Progress >= xuexitongResumeProgress

	// 非视频节点一次提交即可；视频在 fast 模式或已学大半时也直接 finish。
	if !isMedia || resume || fast {
		result, err := client.Study(ctx, yatori.StudyRequest{
			Course:   course,
			Node:     node,
			Action:   yatori.StudyActionFinish,
			Mode:     mode,
			Position: node.Duration,
		})
		if err != nil {
			xuexitongStudyError(t, node, err)
			return
		}
		if xuexitongCaptchaSkipped(t, result) {
			return
		}
		xuexitongLogStudyResult(t, result)
		return
	}

	// 普通模式：先 start，学习通会在这一步返回视频真实时长和播放会话。
	start, err := client.Study(ctx, yatori.StudyRequest{
		Course: course,
		Node:   node,
		Action: yatori.StudyActionStart,
		Mode:   yatori.StudyModeNormal,
	})
	if err != nil {
		xuexitongStudyError(t, node, err)
		return
	}
	if xuexitongCaptchaSkipped(t, start) {
		return
	}
	if node.Duration <= 0 && start != nil && start.Duration > 0 {
		node.Duration = start.Duration
	}
	t.Logf("  学习会话已创建，时长 %s，当前进度 %.1f%%", node.Duration, node.Progress)
	if node.Duration <= 0 {
		t.Errorf("%s 未获取到视频时长（可能平台还在转码），稍后重试", node.Name)
		return
	}

	position := time.Duration(float64(node.Duration) * xuexitongClamp(node.Progress, 0, 100) / 100)
	step := xuexitongProgressStep
	if step > node.Duration {
		step = node.Duration
	}

	for attempts := 0; attempts < 500 && position < node.Duration; attempts++ {
		position += step
		if position > node.Duration {
			position = node.Duration
		}
		result, err := client.Study(ctx, yatori.StudyRequest{
			Course:   course,
			Node:     node,
			Action:   yatori.StudyActionProgress,
			Mode:     yatori.StudyModeNormal,
			Position: position,
		})
		if err != nil {
			xuexitongStudyError(t, node, err)
			return
		}
		if xuexitongCaptchaSkipped(t, result) {
			return
		}
		xuexitongLogStudyResult(t, result)
		if result.Completed || result.Progress >= 99.9 || position >= node.Duration {
			break
		}
		// 学习通普通模式需要真实的学习间隔，提交太快会触发风控。
		time.Sleep(xuexitongProgressInterval)
	}

	result, err := client.Study(ctx, yatori.StudyRequest{
		Course:   course,
		Node:     node,
		Action:   yatori.StudyActionFinish,
		Mode:     yatori.StudyModeNormal,
		Position: node.Duration,
	})
	if err != nil {
		xuexitongStudyError(t, node, err)
		return
	}
	if xuexitongCaptchaSkipped(t, result) {
		return
	}
	xuexitongLogStudyResult(t, result)
}

// xuexitongStudyError 区分“平台不支持”和真正的学习失败。
func xuexitongStudyError(t *testing.T, node yatori.Node, err error) {
	t.Helper()
	if yatori.IsUnsupported(err) {
		t.Logf("  %s 该节点类型暂不支持自动学习，跳过", node.Name)
		return
	}
	t.Errorf("%s 学习失败: %v", node.Name, err)
}

var (
	xuexitongStdinMu   sync.Mutex
	xuexitongStdin     *bufio.Reader
	xuexitongStdinOnce sync.Once
)

// xuexitongLoginCodeProvider 返回 LoginCodeProvider：
//   - 图片验证码：直接用 ddddocr 自动识别，不打扰用户；
//   - 短信验证码：先提示输入 1，用户确认后才真正发送短信，再等待输入收到的验证码。
//
// 注意：core 只提供 SendSMS 这个“发送动作”，不会内嵌任何确认/输入逻辑。
// 下面的“输入 1 才发送、再输入短信验证码”都是测试代码里的调用方策略。
//
// 不设置 LoginCodeProvider 时，SDK 会用内置 ddddocr 识别图片验证码。
func xuexitongLoginCodeProvider(t *testing.T) yatori.LoginCodeProvider {
	t.Helper()
	return func(ctx context.Context, req yatori.LoginCodeRequest) (string, error) {
		xuexitongStdinMu.Lock()
		defer xuexitongStdinMu.Unlock()

		switch req.Kind {
		case "image":
			// 图片验证码不打扰用户：返回空串，SDK 会自动回退到内置 ddddocr。
			return "", nil
		case "sms":
			t.Logf("[短信验证] %s", strings.TrimSpace(req.Prompt))
			if req.Phone != "" {
				t.Logf("接收手机号: %s", req.Phone)
			}
			if req.SendSMS == nil {
				return "", fmt.Errorf("当前登录流程不支持主动发送短信验证码")
			}
			fmt.Print("输入 1 并回车发送短信验证码（直接回车取消）: ")
			confirm, err := xuexitongReadLine()
			if err != nil {
				return "", err
			}
			if confirm != "1" {
				return "", fmt.Errorf("用户取消发送短信验证码")
			}
			if err := req.SendSMS(ctx); err != nil {
				return "", err
			}
			t.Log("短信验证码已发送（图片验证码已由 ddddocr 自动识别），请查看手机短信")
			fmt.Print("请输入收到的短信验证码: ")
		default:
			t.Logf("[验证码] %s %s", req.Prompt, req.Phone)
			fmt.Print("请输入验证码: ")
		}

		code, err := xuexitongReadLine()
		if err != nil {
			return "", err
		}
		if code == "" {
			return "", fmt.Errorf("未输入验证码")
		}
		return code, nil
	}
}

// xuexitongReadLine 读取一行终端输入。
//
// go test 启动的测试进程默认拿不到可交互的 stdin（会被接到空设备），
// 所以这里优先复用管道输入；没有管道时再打开系统控制台（Windows 的
// CONIN$ / Unix 的 /dev/tty），保证在 cmd 里跑 go test -v 也能输入验证码。
func xuexitongReadLine() (string, error) {
	xuexitongStdinOnce.Do(func() {
		if info, err := os.Stdin.Stat(); err == nil && info.Mode()&os.ModeCharDevice == 0 {
			xuexitongStdin = bufio.NewReader(os.Stdin)
			return
		}
		if file, err := xuexitongOpenConsole(); err == nil {
			xuexitongStdin = bufio.NewReader(file)
			return
		}
		xuexitongStdin = bufio.NewReader(os.Stdin)
	})
	line, err := xuexitongStdin.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func xuexitongOpenConsole() (*os.File, error) {
	if runtime.GOOS == "windows" {
		return os.Open("CONIN$")
	}
	return os.Open("/dev/tty")
}

func xuexitongCaptchaSkipped(t *testing.T, result *yatori.StudyResult) bool {
	t.Helper()
	if result == nil {
		return false
	}
	if strings.Contains(result.Message, "验证码") {
		t.Logf("  %s", result.Message)
		return true
	}
	return false
}

func xuexitongLogStudyResult(t *testing.T, result *yatori.StudyResult) {
	t.Helper()
	if result == nil {
		return
	}
	state := "未提交"
	if result.Submitted {
		state = "已提交"
	}
	message := xuexitongShortMessage(result.Message)
	if message == "" {
		t.Logf("  -> %s，进度 %.1f%%，完成=%v", state, result.Progress, result.Completed)
		return
	}
	t.Logf("  -> %s，进度 %.1f%%，完成=%v，%s", state, result.Progress, result.Completed, message)
}

func xuexitongShortMessage(msg string) string {
	msg = strings.TrimSpace(strings.ReplaceAll(msg, "\n", " "))
	if msg == "" {
		return ""
	}
	runes := []rune(msg)
	if len(runes) > 120 {
		return string(runes[:120]) + "..."
	}
	return msg
}

func xuexitongClamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

// xuexitongTestLogger 把 SDK 的警告/错误转发到 testing 日志，
// 例如 App 登录失败回退 Web 登录时会输出一条 WARN。
type xuexitongTestLogger struct{ t *testing.T }

func (xuexitongTestLogger) Debugf(string, ...any) {}
func (xuexitongTestLogger) Infof(string, ...any)  {}

func (l xuexitongTestLogger) Warnf(format string, args ...any) {
	l.t.Logf("[WARN] "+format, args...)
}

func (l xuexitongTestLogger) Errorf(format string, args ...any) {
	l.t.Logf("[ERROR] "+format, args...)
}
