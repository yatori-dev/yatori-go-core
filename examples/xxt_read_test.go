package examples

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestXXTRead(t *testing.T) {
	for {
		time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
		rateStop := rand.Intn(100)
		if rateStop < 30 {
			time.Sleep(time.Duration(rand.Intn(10)+10) * time.Second)
		}
		submitReadLog()
		log.Println("action1")

	}
}

func submitReadLog() {

	t := time.Now().Format("20060102150405") + "000" // 伪造毫秒值
	wc := rand.Intn(1000) + 4000                     //观看的字数
	h := rand.Intn(100) + 4000                       //高度
	//e := `H4sIAAAAAAAAA42aTc4lqQ5EV9PTVtpAAtPu/e+pK4+50hs8ncwafCopQxf8gyNsaH/lv/fKP3/3ms//Z/59//m35hUxY62/2j/tMybA3Iq5wPQPmPYBk4aZG0woZoG5FPNYvedQzA1G7ToY3/MA4/t5Vtn3UkwDo/E6GN8zcb81FpO432rXTSyG5gYW7aFr3fhwqH9ufNjVP1i9u/oHq3dX/2D17r4f/NN8P+Rzc/8URvNw4OemPhz4uem5GOR8qn+I5k7fT2F8P8QifT9gQn1I5uxQHw7iFRrTg9F8HsQ0NO5k4L4c8+x2bfUhGbi8rvYGRv3cE4z6uTBT/dwDjO/5AqM+bBuM+odTs261/WA0pgfjaz3ZvobazgldQ20/GLf9+YXV9weMnsFGbniNohosr1GNuHuNOhhfi9xobju54TUhyY3UPScxDfXPwfhaYC7NMSrhutQuMHOrD/MG42sNMHpO84nCdD12MB9+Z374nek+BHNrPh+M5k8+Hp5+vg7G7QLjGuBgPuzHz05hnLsP5sNa7cNa+WGt9FwF43x6MB/27FxZmOvDWteHtZxPy+rt+VyY97Vu59yDec/V+8MZvD+cwdv7i4Nx/4BxPi1F5zxYGehavdSjc9zBuF3s5OUMssrLGSzMOzfdfgbrazpXXv8TEePu2zkFzNjvOmF4PuOZsd/1z3BOQSEM15kHozmP9hsvOvMC45oWjPMOenXcrlcTjMaiMH4uyOTh/HUwrsOJqfMXX8dLr4d/PFfZyXjp47D9pUd7/o6X/gvbXY8djK9FPl/v/elwTiGawzkFTHfNRmfRl+Yhyqd7zh+MzxyenXSf7fC1u9Zi3tI9nzlZ3fMZxumez3SCvasPD+Z9HtWdC6gY3bkAD3fnAhi5ew/CjLF7PjNd7K5/qJbdc5VpZ7/Uz1Tm5rl6ML5WB6O2o47ay0y4MOrn9Xim+blgJ821TWFc29BVNa/z+/nahp7ljQ/9XODh5rPKwnidZzrUuq81wWhuFMZ7fJRG0z6lFHhTjTRRNU31z7zwoXJKdQRNOeWH8bXwoc7rqkNpqscmv5DaX0yYK/V8/TC6H6pKLotXdWepPcgPYzVqxgCjsSAKqef0h9F4MbFJ1XXVmaae9xlPrqbeg0yqd2pN+GE0XvRxqVw5yfYc/jv4R2eDP4zaTt+UOhucqL7UujHpibJrLDjpqZw7mful1pYJ26by8g+jPuRrKndPeqLUO46a2KRq40lPlDpvmVidXusORv2Mh1P1xqQHSZ3b/DCaG6isDM2Ng1E/k12pWr2mcKlafTIlSNU/NfFLr8/0IKEz2JouhtdneofYGi+qQXgNR0WE13B6h1AdNblDDK/zqJpQrTW50wyv81TCmKZ/Jt10OBfQg4TeE02q9wsGFgjni4PRWHBPHTormPQFoffdE3YL1ZA1UQ+9b5p4OJxT6B1iaG4QzXDeIStC53UThRDax036gnBuIktDNW3dXIT2g5OTFc47nNBo6h+6oXBOQdGFzlIWujeULxbdWShf1C1S6LxlUQlDOeWHsXjVrVYopyx0byin/DDm54U2DuWUBVOEavUF44TyTt3oher5xSpvGHJVe/OFpg3lr4X34nLbC6O2E6lQjltE/A2DD3WutcguHS0vElmnCYvzoGRa18F6n7k4eUqldfGsTLrq5t0hXOSqRfVWQENZzxvU6HpJoQWlHnao0X2+QxYF1SDUa71gWpR97dsWDKOjyYVsUqZecJk+DFmIJuXyhR5SKl9F4uqXuhNQo2sbalG5RC0qx+oJIII6dFqQuOqFRfKrFKjHRsrymxql5Lyv+xVCrddxbj3E6ua6TYV+gTDItnzZ1ENtT+shmyqJTSesD/Dq6Z1qjXrlp+3r5npMu9ddHZNul2P/Aqm5hkJq1GCQuzpXhdBsa4w4ASof6sGvqod6f6ziYe94g/xhTh4yKaTyWyEY/H+N/g9A0+u6ei0AAA==`
	e := ""
	d := `{"a":null,"r":"218393659,436809534","t":"special","l":1,"f":0,"wc":` + fmt.Sprintf("%d", wc) + `,"ic":0,"v":2,"s":2,"h":` + fmt.Sprintf("%d", h) + `,"e":"` + e + `","ext":"{\"_from_\":\"256268467_132232726_339543304_c8f68a62e7ef7fa3a704d6b031d19697\",\"rtag\":\"1054242600_477554005_read-218393659\"}"}`
	settings := map[string]any{
		"f": "readPoint",
		"u": "339543304", //userId
		"s": "",
		"d": url.QueryEscape(d),
		"t": t,
	}

	enc := GenerateEnc(settings)
	//url := "https://data-xxt.aichaoxing.com/analysis/ac_mark?&f=readPoint&u=339543304&d=%257B%2522a%2522%253Anull%252C%2522r%2522%253A%2522218403608%252C437039890%2522%252C%2522t%2522%253A%2522special%2522%252C%2522l%2522%253A1%252C%2522f%2522%253A0%252C%2522wc%2522%253A1457%252C%2522ic%2522%253A1%252C%2522v%2522%253A2%252C%2522s%2522%253A2%252C%2522h%2522%253A168.6666717529297%252C%2522e%2522%253A%2522H4sIAAAAAAAAA43QMQ6AMAgF0NO4GqDAh1XvfycZ6qIMdKDJz4ME7JAbLFVXRFUWyOn1xBJu9R3rsi%252FiCaIBym0CVEO5sr%252BJgcHA%252BMDYwOg25prKSt6YBW1SGXS%252BpwXDJCXRbds10jBjogckSOsm9QEAAA%253D%253D%2522%252C%2522ext%2522%253A%2522%257B%255C%2522_from_%255C%2522%253A%255C%2522256268467_132232726_339543304_c8f68a62e7ef7fa3a704d6b031d19697%255C%2522%252C%255C%2522rtag%255C%2522%253A%255C%25221054242600_477554005_read-218403608%255C%2522%257D%2522%257D&t=20251212142327644&enc=11091c1703d2ba723afdf2338f24b8ae"
	urlStr := "https://data-xxt.aichaoxing.com/analysis/ac_mark?&f=readPoint&u=339543304&d=" + url.QueryEscape(d) + "&t=" + t + "&enc=" + enc
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, urlStr, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Origin", "https://mooc1-1.chaoxing.com")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("sec-ch-ua", "\"Microsoft Edge\";v=\"143\", \"Chromium\";v=\"143\", \"Not A(Brand\";v=\"24\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("Host", "data-xxt.aichaoxing.com")

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
