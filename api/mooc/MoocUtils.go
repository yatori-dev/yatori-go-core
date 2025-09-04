package mooc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tjfoc/gmsm/sm4"
	"github.com/tjfoc/gmsm/x509"
)

// 公钥
var publicKey = "BC60B8B9E4FFEFFA219E5AD77F11F9E2"

// RSA加密公钥
var rsaPublic = "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC5gsH+AA4XWONB5TDcUd+xCz7e\njOFHZKlcZDx+pF1i7Gsvi1vjyJoQhRtRSn950x498VUkx7rUxg1/ScBVfrRxQOZ8\nxFBye3pjAzfb22+RCuYApSVpJ3OO3KsEuKExftz9oFBv3ejxPlYc5yq7YiBO8XlT\nnQN0Sa4R4qhPO3I2MQIDAQAB\n-----END PUBLIC KEY-----"

// 生成rtid
func BuildRtId() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 32)
	for i := 0; i < 32; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
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

// 构建PowGetP的Params参数
func BuildPowGetPParams(pd, pkid, un, pvSid string, channel int, topURL, rtid string) string {
	return `{"pd":"` + pd + `","pkid":"` + pkid + `","un":"` + un + `","pvSid":"` + pvSid + `","channel":"` + strconv.Itoa(channel) + `","topURL": "` + topURL + `","rtid":"` + rtid + `"}`
}

// 构建ZCInit的Params参数
func BuildZCInitParams(pd, pkid, pkht string, channel int, topURL, rtid string) string {
	return `{"channel":` + strconv.Itoa(channel) + `"pd":"` + pd + `","pkht":"` + pkht + `","pkid":"` + pkid + `","rtid":"` + rtid + `","topURL":"` + topURL + `"}`
}

// 构建DLInit的Params参数
func BuildDLInitParams(pd, pkid, pkht string, channel int, topURL, rtid string) string {
	return `{"pd":"` + pd + `","pkid":"` + pkid + `","pkht":"` + pkht + `","channel":` + strconv.Itoa(channel) + `,"topURL":"` + topURL + `","rtid":"` + rtid + `"}`
}

// 构建GT的Params参数
func BuildGTParams(un string, channel int, pd, pkid string, topURL, rtid string) string {
	return `{"un":"` + un + `","pd":"` + pd + `","pkid":"` + pkid + `","channel":` + strconv.Itoa(channel) + `,"topURL":"` + topURL + `","rtid":"` + rtid + `"}`
}

func BuildLParams(l, d int, un, pw, pd, pkid, tk, domains, puzzle string, spendTime, runTimes int, sid, x string, t, sign, channel int, topURL, rtid string) string {
	return `{"l":` + strconv.Itoa(l) + `,"d":` + strconv.Itoa(d) + `,"un":"` + un + `","pw":"` + pw + `","pd":"` + pd + `","pkid":"` + pkid + `","tk":"` + tk + `","domains":"` + domains + `","pVParam":{"puzzle":"` + puzzle + `","spendTime":` + strconv.Itoa(spendTime) + `,"runTimes":` + strconv.Itoa(runTimes) + `,"sid":"` + sid + `","args": "{\"x\":\"` + x + `\",\"t\":` + strconv.Itoa(t) + `,\"sign\":` + strconv.Itoa(sign) + `}"},"channel":` + strconv.Itoa(channel) + `,"topURL":"` + topURL + `","rtid":"` + rtid + `"}`
}

// powgetp转登录数据
func PowGetPTurnLData(puzzle, mod, x string, t int, minTime, maxTime int64) (int, int64, int, string) {
	startTime := time.Now()
	count := 0

	// 解析大整数
	bigx, ok := new(big.Int).SetString(x, 16)
	if !ok {
		panic("解析X失败")
	}
	bigmod, ok := new(big.Int).SetString(mod, 16)
	if !ok {
		panic("解析Mod失败")
	}

	for i := 0; i < t || time.Since(startTime).Milliseconds() < minTime; i++ {
		// bigx = bigx * bigx % bigmod
		bigx.Mul(bigx, bigx)
		bigx.Mod(bigx, bigmod)

		count++

		if time.Since(startTime).Milliseconds() > maxTime {
			break
		}
	}

	timeSpent := time.Since(startTime).Milliseconds()

	//signObj := map[string]any{
	//	"runTimes":  count,
	//	"spendTime": timeSpent,
	//	"t":         count,
	//	"x":         bigx.Text(16), // 转16进制字符串
	//}
	return count, timeSpent, count, bigx.Text(16)
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

// MOOC的RSA加密
func MOOCRSA(content string) string {
	// 解析公钥
	block, _ := pem.Decode([]byte(rsaPublic))
	if block == nil {
		return ""
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return ""
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return ""
	}

	// 使用 PKCS1 v1.5 填充加密
	cipherBytes, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(content))
	if err != nil {
		return ""
	}

	// Base64 输出
	return base64.StdEncoding.EncodeToString(cipherBytes)
}

// MurmurHash3 32-bit 实现 (对应 powSign)
func PowSign(key string, seed int) uint32 {
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

func VdfAsync(data Data) (int, int64, int, string, uint32) {
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

	sign := PowSign(joined, count)

	//fmt.Println("Result x:", bigx.Text(16))
	fmt.Println("SignObj:", signObj)
	fmt.Println("EncodedParams:", joined)
	fmt.Println("Sign:", sign)
	return count, timeSpent, count, bigx.Text(16), sign
}
