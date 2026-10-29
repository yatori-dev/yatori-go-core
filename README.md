![yatori-go-core](https://socialify.git.ci/yatori-dev/yatori-go-core/image?font=Source%20Code%20Pro&forks=1&language=1&logo=https%3A%2F%2Fyatori-dev.github.io%2Fyatori-docs%2Fimg%2Flogo.png&name=1&owner=1&pattern=Floating%20Cogs&pulls=1&stargazers=1&theme=Dark)

<div align="center"><h2>Yatori-go-core</h2></div>

<div align="center"><img width="125px" src="https://img.shields.io/badge/GO1.27-building-r.svg?logo=go"></img> <img width="80px" src="https://img.shields.io/github/stars/yatori-dev/yatori-go-core.svg"></img> <img width="90px" src="https://img.shields.io/github/downloads/yatori-dev/yatori-go-core/total.svg"></img> <img width="70px" src="https://img.shields.io/github/license/yatori-dev/yatori-go-core.svg"></div>

Yatori-go-core 是 yatori 系列的 Go 核心库。它把多个在线学习平台的登录、课程、章节、学习、作业和考试能力收敛到同一套接口，第三方项目可以直接导入并做二次开发。

> 文档站：[使用教程](https://yatori-dev.github.io/yatori-docs/yatori-go-core/docs.html) · [开发手册](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/) · [新增平台教程](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/new-platform.html)

## 官方文档

完整文档站：**[https://yatori-dev.github.io/yatori-docs/](https://yatori-dev.github.io/yatori-docs/)**

使用文档：

- [V3 总览](https://yatori-dev.github.io/yatori-docs/yatori-go-core/docs.html)
- [快速开始](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/getting-started.html)
- [Client API](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/client-api.html)
- [平台说明与账号参数](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/platforms.html)
- [完整示例](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/examples.html)
- [常见问题](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/faq.html)

开发文档：

- [开发手册总览](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/)
- [架构与分层](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/architecture.html)
- [新增平台完整教程](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/new-platform.html)
- [Adapter API 规范](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/adapter-api.html)
- [数据映射规范](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/data-mapping.html)
- [测试与验收](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/testing.html)
- [代码风格与 PR](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/style-and-pr.html)

## 安装

```bash
go get github.com/yatori-dev/yatori-go-core
```

要求 Go 1.27+ 和 CGO（部分平台依赖 SQLite / ONNX Runtime）。

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/yatori-dev/yatori-go-core"
)

func main() {
	client, err := yatori.New(yatori.Config{
		Platform: yatori.XueXiTong,
		Account: yatori.Account{
			Username: "ACCOUNT",
			Password: "PASSWORD",
			// 英华、海旗等平台需要填写 URL：
			// URL: "https://example.com",
		},
		Options: yatori.Options{
			// ProxyURL: "http://127.0.0.1:7890",
			// AI: &yatori.AIConfig{
			// 	Type:   yatori.AIDeepSeek,
			// 	Model:  "MODEL",
			// 	APIKey: "KEY",
			// },
		},
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Login(ctx); err != nil {
		panic(err)
	}

	courses, err := client.Courses(ctx)
	if err != nil {
		panic(err)
	}
	for _, course := range courses {
		fmt.Println(course.Name, course.Progress)
		nodes, err := client.Tree(ctx, course)
		if err != nil {
			continue
		}
		for _, node := range yatori.Flatten(nodes) {
			if node.Completed || node.Kind == yatori.NodeChapter {
				continue
			}
			// Study 是“提交一次学习进度”的原语：
			// 调用方自行控制循环、间隔和风控节奏。
			_, _ = client.Study(ctx, yatori.StudyRequest{
				Course:   course,
				Node:     node,
				Action:   yatori.StudyActionProgress,
				Position: 60 * time.Second,
				Mode:     yatori.StudyModeNormal,
			})
		}
	}
}
```

> Study 是进度原语：不同平台粒度不同。学习通/英华/CQIE/TTCDW 接受 `Position` 进度；职教云/课堂行/WeLearn 会直接完成整个节点；青书/海旗按平台会话提交。请以返回的 `StudyResult` 为准，不要把某个平台的循环节奏套到所有平台。

更多配置、错误处理和完整流程见 [快速开始文档](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/getting-started.html)。

## 统一接口

所有平台都实现同一个 `yatori.Client`：

```go
type Client interface {
	Platform() Platform
	Capabilities() Capability

	Login(ctx context.Context) error
	LoginWithCookie(ctx context.Context, cookie string) error
	Session() Session

	Courses(ctx context.Context) ([]Course, error)
	Tree(ctx context.Context, course Course) ([]Node, error)

	Study(ctx context.Context, req StudyRequest) (*StudyResult, error)

	Works(ctx context.Context, course Course) ([]Work, error)
	DoWork(ctx context.Context, req WorkRequest) (*AnswerResult, error)

	Exams(ctx context.Context, course Course) ([]Exam, error)
	DoExam(ctx context.Context, req ExamRequest) (*AnswerResult, error)

	SignIn(ctx context.Context, req SignInRequest) (*SignInResult, error)

	Call(ctx context.Context, action Action, params map[string]any) (any, error)
	Close() error
}
```

平台不支持的操作返回 `yatori.ErrUnsupported`，调用前可用 `client.Capabilities()` 判断：

```go
	if yatori.Supports(client, yatori.CapWorks) {
	works, _ := client.Works(ctx, course)
	for _, work := range works {
		_, _ = client.DoWork(ctx, yatori.WorkRequest{
			Course: course,
			Work:   work,
			Source: yatori.AnswerSourceAI,
			Submit: true,
		})
	}
}
```

逐方法说明见 [Client API 文档](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/client-api.html)，能力位的完整矩阵见 [能力与错误](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/capabilities.html)。

## 平台能力

| 平台 | 登录 | Cookie/Token | 课程 | 节点树 | 学习 | 作业 | 考试 |
|---|---|---|---|---|---|---|---|
| XUEXITONG | ✅ | ✅ | ✅ | ✅ | ✅ 视频/音频/文档/外链 | ✅ | ✅ |
| YINGHUA | ✅ | — | ✅ | ✅ | ✅ | ✅ | ✅ |
| ICVE | ✅ | ✅ | ✅ | ✅ | ✅ | — | — |
| CQIE | ✅ | Token | ✅ | ✅ | ✅ | 内嵌于视频段 | — |
| ENAAE | ✅ | ✅ | ✅ | ✅ | ✅ | — | 列表 |
| WELEARN | ✅ | ✅ | ✅ | ✅ | ✅ | — | — |
| QSXT | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — |
| TTCDW | ✅ | ✅ | ✅ | ✅ | ✅ | — | — |
| KETANGX | ✅ | — | ✅ | ✅ | ✅ | — | — |
| ZXCPKS | ✅ | — | ✅ | ✅ | ✅ | ✅ | ✅ |
| WEIBAN | ✅ | — | — | — | — | — | — |
| CELA | ✅ | — | — | — | — | — | — |
| MOOC | ✅ | — | — | — | — | — | — |
| ZHIHUISHU | 扫码 | — | — | — | — | — | — |
| GONGXUE | 签到 | — | — | — | — | — | — |

### 账号参数

| 平台 | 必填字段 | 说明 |
|---|---|---|
| XUEXITONG | Username, Password | 也支持把 Cookie 填到 `Account.Cookie` |
| YINGHUA | URL, Username, Password | URL 必须是登录后的站点地址 |
| ICVE | Username, Password | 也支持 Cookie |
| CQIE | Username, Password | 也支持把 access_token 填到 `Account.Cookie` |
| ENAAE | Username, Password | 也支持 Cookie |
| WELEARN | Username, Password | 也支持 Cookie |
| QSXT | Username, Password | 也支持把 token 填到 `Account.Cookie` |
| TTCDW | Username, Password | 也支持 Cookie |
| KETANGX | Username, Password | — |
| ZXCPKS | URL, Username, Password | URL 为学校站点 |
| WEIBAN | School, Username, Password | `School` 为学校名称 |
| CELA / MOOC | Username, Password | 登录验证码自动识别 |
| ZHIHUISHU | 无 | 使用 `QRCodeLogin` 扫码 |
| GONGXUE | Username, Password | `SignInRequest.Extra` 传国家/省/市/经纬度等 |

智慧树使用扫码登录：

```go
qr, ok := client.(yatori.QRCodeLogin)
if ok {
	imageBase64, token, _ := qr.QRCode(ctx)
	// 把 imageBase64 转成图片交给用户扫描
	session, err := qr.WaitQRCode(ctx, token)
	_ = session
	_ = err
}
```

更详细的账号参数、平台限制和扫码流程见 [平台说明文档](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/platforms.html)。

## 动态调用

对于配置驱动或 Web 后端场景，可以用 `Call` 免去类型分支：

```go
result, err := client.Call(ctx, yatori.ActionCourses, nil)

result, err = client.Call(ctx, yatori.ActionStudy, map[string]any{
	"courseId": "COURSE_ID",
	"nodeId":   "NODE_ID",
	"mode":     "normal",
	"position": "60s",
})
```

完整的 Action、参数和返回类型表见 [动态调用文档](https://yatori-dev.github.io/yatori-docs/yatori-go-core/v3/dynamic-call.html)。

## 目录结构

```
yatori-go-core
├─ yatori.go / types.go / errors.go   # 唯一公开入口
├─ core/                              # 接口、领域模型、错误、能力位、动态分发
├─ internal/driver/                   # 各平台适配器（外部不可导入）
│  ├─ xuexitong/ yinghua/ icve/ ...
│  └─ driverutil/                     # 适配器共享工具
├─ internal/api/                      # 底层协议实现（原 api/）
├─ internal/aggregation/              # 平台业务动作（原 aggregation/）
├─ que-core/                          # AI 题库与标准题目模型
├─ utils/ models/ config/ global/     # 兼容保留
├─ cmd/hello/                         # 原根目录 main.go
└─ examples/                          # 示例与回归测试
```

> 迁移提示：原来的 `api/*` 与 `aggregation/*` 已移动到 `internal/`，外部项目不能再直接导入。请改用 `yatori.New(...)` 和 `Client` 接口。`config`、`global`、`common` 为兼容保留，新代码不需要它们。

> 运行目录提示：部分平台仍沿用旧实现的相对路径（验证码图片、职教云辅助 exe、学习通人脸素材），默认写入当前工作目录下的 `./assets`。请确保运行目录可写；OCR 模型只在首次登录相关平台时释放。

## 开发

贡献代码前请先阅读 [开发手册总览](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/)。新增平台请按 [新增平台完整教程](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/new-platform.html) 操作，并遵循 [Adapter API 规范](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/adapter-api.html)、[数据映射规范](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/data-mapping.html) 和 [代码风格与 PR 清单](https://yatori-dev.github.io/yatori-docs/yatori-go-core/dev/style-and-pr.html)。

```bash
go build ./...
go vet ./...
go test ./...
```

`examples/` 中的旧测试依赖 `examples/TestData.yaml` 与真实账号，默认不参与普通单元测试。新代码的回归入口是：

```bash
go test ./core ./internal/driver/... .
```

## License

见 [LICENSE](LICENSE)。
