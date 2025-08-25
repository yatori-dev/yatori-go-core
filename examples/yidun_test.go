package examples

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)
import "github.com/dop251/goja"

// Dun163 结构体
type Dun163 struct {
	id        string
	referer   string
	fpH       string
	ua        string
	client    *http.Client
	jsRuntime *goja.Runtime
	fp        string
}

// NewDun163 初始化
func NewDun163(id, referer, fpH, ua string) *Dun163 {
	return &Dun163{
		id:      id,
		referer: referer,
		fpH:     fpH,
		ua:      ua,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		jsRuntime: goja.New(),
	}
}

// CompileJS 加载 dun163.js
func (d *Dun163) CompileJS(path string) error {
	code, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	globalObj := d.jsRuntime.GlobalObject()
	d.jsRuntime.Set("global", globalObj) // 给 JS 提供 global
	d.jsRuntime.Set("window", globalObj) // 让 window = global
	_, err = d.jsRuntime.RunString(string(code))

	return err
}

// 调用 JS 函数
func (d *Dun163) callJS(funcName string, args ...interface{}) (goja.Value, error) {
	fn, ok := goja.AssertFunction(d.jsRuntime.Get(funcName))
	if !ok {
		return nil, fmt.Errorf("function %s not found", funcName)
	}

	// 转换参数
	gojaArgs := make([]goja.Value, len(args))
	for i, arg := range args {
		gojaArgs[i] = d.jsRuntime.ToValue(arg)
	}

	return fn(goja.Undefined(), gojaArgs...)
}

// JSONP 解析
func getJSONP(text string, v any) error {
	re := regexp.MustCompile(`\((.*)\)`)
	match := re.FindStringSubmatch(text)
	if len(match) < 2 {
		return fmt.Errorf("no jsonp body found")
	}
	return json.Unmarshal([]byte(match[1]), v)
}

// 随机 JSONP callback
func randomJSONPStr() string {
	letters := "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 7)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return "__JSONP_" + string(b) + "_"
}

// 下载图片
func downloadImg(fileName, url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dir := "pic"
	_ = os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, fileName)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return path, err
}

// OCR 识别（TODO: 这里需要对接 ddddocr 或 gocv）
func getSlideDistance(fg, bg string) int {
	// TODO 调用 Python 服务 or gocv 模板匹配
	return 100 // mock
}

// 轨迹生成（翻译 Python get_track）
func getTrack(distance int) [][]int {
	baseTrack := [][]int{
		{4, 0, 94}, {6, 0, 102}, {9, 0, 111}, {13, 0, 118}, {18, 0, 126}, {22, 0, 134},
		{28, 0, 140}, {32, 0, 148}, {35, 0, 156}, {40, 0, 164}, {42, 1, 172}, {45, 2, 180},
		{46, 2, 189}, {47, 3, 196}, {49, 4, 204}, {50, 4, 212}, {51, 4, 220}, {52, 4, 237},
		{53, 4, 244}, {54, 4, 252}, {55, 4, 260}, {57, 4, 268}, {58, 4, 276}, {60, 4, 294},
		{62, 4, 373}, {62, 4, 380}, {63, 4, 388}, {65, 4, 396}, {66, 4, 405}, {67, 4, 412},
		{68, 4, 421}, {70, 4, 428}, {73, 5, 437}, {74, 5, 444}, {75, 6, 452}, {78, 6, 460},
		{80, 7, 468}, {82, 8, 477}, {84, 8, 485}, {86, 8, 492}, {90, 8, 501}, {94, 8, 509},
		{95, 8, 518}, {98, 8, 525}, {102, 9, 533}, {105, 10, 541}, {106, 10, 588}, {107, 10, 604},
		{109, 10, 612}, {110, 10, 620}, {110, 10, 628}, {113, 11, 636}, {115, 11, 644}, {116, 11, 653},
		{118, 11, 660}, {118, 11, 668}, {120, 11, 676}, {122, 11, 684}, {122, 11, 692}, {123, 11, 700},
		{124, 12, 764}, {125, 12, 772}, {126, 12, 788}, {128, 12, 804}, {129, 12, 812}, {130, 12, 1190},
		{130, 12, 1252}, {131, 12, 1268}, {132, 12, 1340}, {134, 12, 1710},
	}
	randomY := rand.Intn(9) - 3
	ratio := float64(distance) / float64(baseTrack[len(baseTrack)-1][0]-baseTrack[0][0])
	newTrack := [][]int{}
	for _, p := range baseTrack {
		x := int(float64(p[0]) * ratio)
		y := p[1]
		if y != 0 {
			y += randomY
		}
		t := int(float64(p[2]) * ratio)
		newTrack = append(newTrack, []int{x, y, t})
	}
	return newTrack
}

// ---------- API ----------

// request_getconf
func (d *Dun163) requestGetConf() (map[string]any, error) {
	url := "https://c.dun.163.com/api/v2/getconf"
	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("referer", d.referer)
	q.Add("zoneId", "")
	q.Add("id", d.id)
	q.Add("ipv6", "false")
	q.Add("runEnv", "10")
	q.Add("iv", "3")
	q.Add("loadVersion", "2.5.3")
	q.Add("callback", randomJSONPStr()+"0")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", d.ua)
	req.Header.Set("Referer", d.referer)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Data map[string]any `json:"data"`
	}
	if err := getJSONP(string(body), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// request_get
func (d *Dun163) requestGet(dt, bid, acToken string) (map[string]any, error) {
	url := "https://c.dun.163.com/api/v3/get"
	// JS 调用生成 fp, cb
	fpVal, _ := d.callJS("get_fp", d.fpH, d.ua)
	cbVal, _ := d.callJS("get_cb")
	d.fp = fpVal.String()

	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("referer", d.referer)
	q.Add("zoneId", "CN31")
	q.Add("dt", dt)
	q.Add("acToken", acToken)
	q.Add("id", bid)
	q.Add("fp", d.fp)
	q.Add("https", "true")
	q.Add("type", "undefined")
	q.Add("version", "2.28.5")
	q.Add("dpr", "1.25")
	q.Add("dev", "1")
	q.Add("cb", cbVal.String())
	q.Add("ipv6", "false")
	q.Add("runEnv", "10")
	q.Add("iv", "3")
	q.Add("width", "320")
	q.Add("sizeType", "10")
	q.Add("smsVersion", "v3")
	q.Add("token", "")
	q.Add("callback", randomJSONPStr()+"0")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", d.ua)
	req.Header.Set("Referer", d.referer)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Data map[string]any `json:"data"`
	}
	if err := getJSONP(string(body), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// request_check
func (d *Dun163) requestCheck(dt, bid, token string, track [][]int, distance int) (map[string]any, error) {
	url := "https://c.dun.163.com/api/v3/check"

	checkData, _ := d.callJS("get_check_data", track, distance, token)
	cbVal, _ := d.callJS("get_cb")

	req, _ := http.NewRequest("GET", url, nil)
	q := req.URL.Query()
	q.Add("referer", d.referer)
	q.Add("zoneId", "CN31")
	q.Add("dt", dt)
	q.Add("id", bid)
	q.Add("token", token)
	q.Add("acToken", "undefined")
	q.Add("data", checkData.String())
	q.Add("width", "320")
	q.Add("type", "2")
	q.Add("version", "2.28.5")
	q.Add("cb", cbVal.String())
	q.Add("bf", "0")
	q.Add("runEnv", "10")
	q.Add("iv", "3")
	q.Add("callback", randomJSONPStr()+"1")
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", d.ua)
	req.Header.Set("Referer", d.referer)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var res struct {
		Data map[string]any `json:"data"`
	}
	if err := getJSONP(string(body), &res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

// ---------- RUN ----------
func (d *Dun163) Run() bool {
	conf, err := d.requestGetConf()
	if err != nil {
		fmt.Println("getconf error:", err)
		return false
	}
	dt := conf["dt"].(string)
	ac := conf["ac"].(map[string]any)
	acToken := ac["token"].(string)
	bid := ac["bid"].(string)
	if bid != d.id {
		fmt.Println("id 不一致:", bid, d.id)
		return false
	}

	// 获取图片
	getData, _ := d.requestGet(dt, bid, acToken)
	bgUrl := getData["bg"].([]any)[0].(string)
	frontUrl := getData["front"].([]any)[0].(string)
	downloadImg("bg.png", bgUrl)
	downloadImg("front.png", frontUrl)

	// 识别距离
	distance := getSlideDistance("pic/front.png", "pic/bg.png")
	track := getTrack(distance + 7)
	time.Sleep(4 * time.Second)

	// 请求校验
	checkData, _ := d.requestCheck(dt, bid, getData["token"].(string), track, distance)
	fmt.Println("check result:", checkData)
	return checkData["result"].(bool)
}

func Test_yidun(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	d := NewDun163("17c07e34e0384612bb239568b6b37643",
		"https://id.163.com/mail/retrievepassword",
		"app.miit-eidc.org.cn",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	if err := d.CompileJS("dun163.js"); err != nil {
		panic(err)
	}

	success := d.Run()
	fmt.Println("验证成功？", success)
}
