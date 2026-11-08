# 学习通：App / Web 登录到课程学习示例

示例文件：`examples/xuexitong_sdk_test.go`，和仓库现有示例一样使用 `_test.go` 形式：

| 测试函数 | 登录方式 | 说明 |
| --- | --- | --- |
| `TestXueXiTongAppSDKStudy` | `yatori.LoginModeApp` | 本地 deviceInfo → `clientId` → `cid/sc`，短信验证码由用户确认后发送并输入 |
| `TestXueXiTongWebSDKStudy` | `yatori.LoginModeWeb` | passport `fanyalogin`，也可以改用 `LoginWithCookie` |

账号统一从 `examples/TestData.yaml` 读取：

```go
utils.YatoriCoreInit()
setup()
user := global.Config.Users[xuexitongUserIndex]
```

`xuexitongUserIndex` 默认是 `78`，改成自己的学习通账号下标即可，要求
`accountType: XUEXITONG`。`TestData.yaml` 已在 `.gitignore` 中，不会入库。

## 运行

```bash
# App 登录：登录 -> 课程 -> 章节 -> 学习
go test ./examples -run TestXueXiTongAppSDKStudy -v

# Web 登录：同样的流程
go test ./examples -run TestXueXiTongWebSDKStudy -v

# 只验证登录和课程拉取，不执行学习
go test ./examples -run TestXueXiTongWebSDKStudy -v -short
```

## 职责边界

core 只负责学习通协议本身：

- 解析 `tocheckmessage` 页面参数；
- 提供 `LoginCodeRequest.SendSMS` 这个发送入口；
- `SendSMS` 内部完成 `/num/booleanCode` 预检、图片验证码 ddddocr 识别、`/num/phonecode` 发送；
- 提交 `/v11/checkmessage`，继续 `clientId` / `cid/sc`。

测试或业务代码负责交互策略：

- 是否提示“输入 1 发送验证码”；
- 什么时候调用 `req.SendSMS`；
- 读取用户输入的短信验证码并返回给 SDK。

这些交互逻辑没有内嵌到 core。

## 核心调用

### 1. 创建客户端（App 登录）

```go
user := global.Config.Users[xuexitongUserIndex]

client, err := yatori.New(yatori.Config{
    Platform: yatori.XueXiTong,
    Account: yatori.Account{
        Username: user.Account,
        Password: user.Password,
    },
    Options: yatori.Options{
        LoginMode:         yatori.LoginModeApp,
        LoginCodeProvider: xuexitongLoginCodeProvider(t), // 图片/短信验证码
    },
})
if err != nil {
    t.Fatalf("创建客户端失败: %v", err)
}
t.Cleanup(func() { _ = client.Close() })
```

Web 登录只需要把 `LoginMode` 换成 `yatori.LoginModeWeb`。

### 2. App 短信登录

SDK 不会自动发短信。触发短信验证时，`LoginCodeRequest` 里的 `SendSMS`
是调用方显式触发发送的入口：

```go
func promptLoginCode(ctx context.Context, req yatori.LoginCodeRequest) (string, error) {
    if req.Kind == "image" {
        // 实际不会走到：图片验证码由 SDK 内部 ddddocr 识别并持续重试。
        return "", nil
    }

    // 1. 等用户确认，避免一登录就发短信
    fmt.Print("输入 1 并回车发送短信验证码: ")
    if readFromConsole() != "1" {
        return "", fmt.Errorf("用户取消发送短信验证码")
    }

    // 2. 真正的发送动作。发送前的图片验证码由 SDK 内置
    //    ddddocr 自动识别并重试，不需要用户处理。
    if err := req.SendSMS(ctx); err != nil {
        return "", err
    }

    // 3. 等用户输入手机收到的短信验证码
    fmt.Print("请输入收到的短信验证码: ")
    return readFromConsole(), nil
}
```

示例里的 `xuexitongLoginCodeProvider` 就是这么实现的。因为 `go test`
启动的测试进程默认拿不到可交互的 stdin，示例会主动打开系统控制台
（Windows `CONIN$` / Unix `/dev/tty`），所以在 cmd 里直接
`go test -v` 也能正常输入。

### 3. 登录

```go
ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
t.Cleanup(cancel)

if err := client.Login(ctx); err != nil {
    t.Fatalf("登录失败: %v", err)
}
```

已有浏览器 Cookie 时：

```go
err = client.LoginWithCookie(ctx, "UID=...; vc3=...;")
```

### 4. 拉课程 + 章节树

```go
courses, err := client.Courses(ctx)

nodes, err := client.Tree(ctx, course)
for _, node := range yatori.Flatten(nodes) {
    // node.Kind: video/audio/document/link/work/exam/...
}
```

### 5. 学习单个节点

普通模式（真实观看节奏）：

```go
// 1) 创建会话，拿真实时长
start, err := client.Study(ctx, yatori.StudyRequest{
    Course: course,
    Node:   node,
    Action: yatori.StudyActionStart,
    Mode:   yatori.StudyModeNormal,
})

// 2) 按步进提交进度
result, err := client.Study(ctx, yatori.StudyRequest{
    Course:   course,
    Node:     node,
    Action:   yatori.StudyActionProgress,
    Mode:     yatori.StudyModeNormal,
    Position: 60 * time.Second,
})

// 3) 收尾
result, err = client.Study(ctx, yatori.StudyRequest{
    Course:   course,
    Node:     node,
    Action:   yatori.StudyActionFinish,
    Mode:     yatori.StudyModeNormal,
    Position: node.Duration,
})
```

快速模式（直接提交完成）：

```go
client.Study(ctx, yatori.StudyRequest{
    Course: course,
    Node:   node,
    Action: yatori.StudyActionFinish,
    Mode:   yatori.StudyModeFast,
})
```

文档、链接等非视频节点直接 `StudyActionFinish` 即可。

## 注意事项

1. **App 登录可能触发短信验证。**
   Go 的 TLS/HTTP 指纹和真实 App 不一致时，服务端会判定为新设备并返回 `type=2`。
   SDK 不会自动发送短信：`LoginCodeProvider` 先让用户输入 `1` 确认，确认后
   调用 `SendSMS` 触发发送；发送前的图片验证码由内置 ddddocr 自动识别并重试，
   最后只需输入手机收到的短信验证码。重试登录不是可靠的绕过方式。
2. **App 登录失败会自动回退 Web 登录。**
   例如取消发送短信、短信验证码错误时会回退，并在日志里输出一条 `WARN`。
   回退后没有 `cid/sc`，需要人脸签名的课程要额外配置 `FaceSignature`。
3. **普通模式不要提交太快。**
   学习通对学习间隔有校验，示例默认每 30 秒提交一次，可调整
   `xuexitongProgressStep` / `xuexitongProgressInterval`。
4. **卡片验证码。**
   章节卡片拉取过密会触发图片验证码。`Study` 会返回 `Message` 含“验证码”的空结果，
   示例会跳过该节点而不是反复重试。
5. **完成度判断。**
   平台常把看完的视频回报成 99.x%，示例用 `99.0` 作为完成阈值。
6. **设备信息按账号持久化，并且每个账号各不相同。**
   App 登录会为每个账号分配独立的机型、`deviceUniqueId` / `mediaDrmId` /
   `android_id` 以及 UA 里的 `schild` / `(@Kalimdor)` 令牌，写入
   `assets/xxt_device/<account>.json`。同一账号后续登录复用同一份设备信息，
   不同账号之间的设备指纹不会相同；`time_stamp` 仍然每次重新生成。
   删除该文件后会重新生成一套设备信息。
7. **图片验证码不会转人工。**
   登录过程中的图片验证码始终由内置 ddddocr 识别；识别为空或被平台拒绝时，
   会自动换一张继续识别，直到通过或上下文取消，不会调用 `LoginCodeProvider`
   要求用户输入图片验证码。
