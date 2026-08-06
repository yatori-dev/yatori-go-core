package xuexitong

import (
	"net"
	"strings"
	"testing"

	xuexitongapi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
)

func TestWorkNewSubmitAnswerActionPropagatesRequestError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxyAddress := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	cache := &xuexitongapi.XueXiTUserCache{
		IpProxySW: true,
		ProxyIP:   proxyAddress,
	}
	_, err = WorkNewSubmitAnswerAction(cache, xuexitongapi.Question{}, true)
	if err == nil {
		t.Fatal("WorkNewSubmitAnswerAction() error = nil, want request error")
	}
	if !strings.Contains(err.Error(), "提交章节测试任务点请求失败") {
		t.Fatalf("WorkNewSubmitAnswerAction() error = %v, want submission context", err)
	}
}
