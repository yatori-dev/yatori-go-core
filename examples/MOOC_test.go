package examples

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tjfoc/gmsm/sm4"
	"github.com/tjfoc/gmsm/x509"
	"github.com/yatori-dev/yatori-go-core/api/mooc"
	"github.com/yatori-dev/yatori-go-core/global"
	"github.com/yatori-dev/yatori-go-core/utils"
)

// 公钥
var publicKey = "BC60B8B9E4FFEFFA219E5AD77F11F9E2"

// RSA加密公钥
var rsaPublic = "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC5gsH+AA4XWONB5TDcUd+xCz7e\njOFHZKlcZDx+pF1i7Gsvi1vjyJoQhRtRSn950x498VUkx7rUxg1/ScBVfrRxQOZ8\nxFBye3pjAzfb22+RCuYApSVpJ3OO3KsEuKExftz9oFBv3ejxPlYc5yq7YiBO8XlT\nnQN0Sa4R4qhPO3I2MQIDAQAB\n-----END PUBLIC KEY-----"

// 测试MOOC国密加密
func TestMOOCEnc(t *testing.T) {
	content := "{\"pd\":\"imooc\",\"pkid\":\"cjJVGQM\",\"un\":\"18973485974\",\"pvSid\":\"5722fb36-7665-4510-8281-c202f414978c\",\"channel\":1,\"topURL\":\"https://www.icourse163.org/member/login.htm?returnUrl=aHR0cHM6Ly93d3cuaWNvdXJzZTE2My5vcmcvaW5kZXguaHRt#/webLoginIndex\",\"rtid\":\"fKJsUMSbYK6n16YDC8kdl9AlFYm7MPwm\"}"
	key, err := hex.DecodeString(publicKey)
	out, err := sm4.Sm4Ecb(key, []byte(content), true)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("cipher (hex): %x\n", out)
}

// 测试MOOC密码加密
func TestPasswordEnc(t *testing.T) {
	pub, err := parseRSAPublicKey(rsaPublic)
	if err != nil {
		log.Fatalf("parse pub key error: %v", err)
	}

	plain := []byte("hello world from Go RSA")

	// 用 PKCS1 v1.5 加密（和 JS 对应）
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plain)
	if err != nil {
		log.Fatalf("encrypt error: %v", err)
	}

	// Base64 输出
	cipherText := base64.StdEncoding.EncodeToString(cipherBytes)
	fmt.Println(cipherText)
}

// PowGetP接口测试
func TestPowGetP(t *testing.T) {
	utils.YatoriCoreInit()
	//测试账号
	setup()
	cache := mooc.MOOCUserCache{
		Account:   global.Config.Users[22].Account,
		Password:  global.Config.Users[22].Password,
		IpProxySW: false,
		ProxyIP:   "",
	}
	cache.InitCookies()
	cache.Gt()
	cache.PowGetP()
	cache.Login()
	//encStr := MOOCEncMS4(mooc.BuildInitParams("imooc", "cjJVGQM", "www.icourse163.org", 1, "https://www.icourse163.org/member/login.htm?returnUrl=aHR0cHM6Ly93d3cuaWNvdXJzZTE2My5vcmcvaW5kZXguaHRt#/webLoginIndex", "XFuXB7Y6xxJm8zsbasAejdCtmc0FvdqF"))
	//fmt.Println(encStr)
}

func TestSign(t *testing.T) {
	runTimes, spendTime, T, x := mooc.PowGetPTurnLData("woVmIfMmB3qI6a7ywfvS+/7oyCpQ0cGCf+o2wYqut+iq5vqHIxB+7GJa4ajOM08G2q0gIL8ePHLU\\r\\nLJXtOD3oWZS2Iu+jxYAt2MSm9Np7UDm6StJCqWB+n6O19tvyhbLaBzgTZjaa+Wt9P46BBWplwJBz\\r\\nw1bFrbQdGeO5n1WpuTZZu0OykSyQ1t7H/x7PwBwXIwP68tz06wYJvbFXsn/z2CmEkiRuPC12rfDh\\r\\nmWjJmyDKOK/R6I9EImUTXhbCtWNhapPNEnD25FFw7vfWZdt5kA==", "a5bac8df1b6e043e841e9467a7c20d6919", "9099200a75", 200000, 1000, 1050)
	fmt.Println(runTimes, spendTime, T, x)
}

// 解析 RSA 公钥
func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	pub, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}
	return pub, nil
}

// MOOC请求用的MS4国密算法加密
func MOOCEncMS4(content string) string {
	key, err := hex.DecodeString(publicKey)
	out, err := sm4.Sm4Ecb(key, []byte(content), true)
	if err != nil {
		error.Error(err)
	}
	fmt.Printf("cipher (hex): %x\n", out)
	return fmt.Sprintf("%x", out)
}

// MurmurHash3 32-bit 实现 (对应 powSign)
func powSign(key string, seed int) uint32 {
	data := []byte(key)
	var h1 = uint32(seed)
	var c1 uint32 = 0xcc9e2d51
	var c2 uint32 = 0x1b873593

	// body
	nblocks := len(data) / 4
	for i := 0; i < nblocks; i++ {
		k1 := uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*5 + 0xe6546b64
	}

	// tail
	var k1 uint32
	tail := data[nblocks*4:]
	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2
		h1 ^= k1
	}

	// finalization
	h1 ^= uint32(len(data))
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}

type Args struct {
	Mod    string
	T      int
	Puzzle string
	X      string
}

type Data struct {
	NeedCheck bool
	Sid       string
	HashFunc  string
	MaxTime   int64
	MinTime   int64
	Args      Args
}

func vdfAsync(data Data) {
	startTime := time.Now()

	// 解析大整数
	bigx, ok := new(big.Int).SetString(data.Args.X, 16)
	if !ok {
		panic("解析 x 失败")
	}
	bigmod, ok := new(big.Int).SetString(data.Args.Mod, 16)
	if !ok {
		panic("解析 mod 失败")
	}

	count := 0
	tmp := new(big.Int)
	sq := new(big.Int)

	for i := 0; i < data.Args.T || time.Since(startTime).Milliseconds() < data.MinTime; i++ {
		// bigx = bigx * bigx % bigmod
		sq.Mul(bigx, bigx)
		tmp.Mod(sq, bigmod)
		bigx.Set(tmp)

		count++

		if time.Since(startTime).Milliseconds() > data.MaxTime {
			break
		}
	}

	timeSpent := time.Since(startTime).Milliseconds()

	signObj := map[string]any{
		"runTimes":  count,
		"spendTime": timeSpent,
		"t":         count,
		"x":         strings.ToLower(bigx.Text(16)),
	}

	// 排序参数
	sortedParams := []string{"runTimes", "spendTime", "t", "x"}
	var encodedParams []string
	for _, key := range sortedParams {
		value := fmt.Sprintf("%v", signObj[key])
		encodedParams = append(encodedParams, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
	joined := strings.Join(encodedParams, "&")

	sign := powSign(joined, count)

	fmt.Println("Result x:", bigx.Text(16))
	fmt.Println("SignObj:", signObj)
	fmt.Println("EncodedParams:", joined)
	fmt.Println("Sign:", sign)
}

func TestCC(t *testing.T) {
	vdfAsync(Data{
		NeedCheck: true,
		Sid:       "000c9adb-26d5-42cc-8b03-cd8bba2999d1",
		HashFunc:  "VDF_FUNCTION",
		MaxTime:   1050,
		MinTime:   1000,
		Args: Args{
			Mod:    "a5bac8df1b6e043e841e9467a7c20d6919",
			T:      200000,
			Puzzle: "woVmIfMmB3qI6a7ywfvS+/7oyCpQ0cGCf+o2wYqut+iq5vqHIxB+7GJa4ajOM08G2q0gIL8ePHLU...",
			X:      "9099200a75",
		},
	})
}
