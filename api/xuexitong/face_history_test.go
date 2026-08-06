package xuexitong

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetHistoryFaceImgClosesImageResponse(t *testing.T) {
	metadataClient := &http.Client{Transport: faceRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return faceResponse(`{"result":1,"data":{"http":"https://passport2-api.chaoxing.com/face.png"}}`), nil
	})}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	imageBody := &trackedReadCloser{Reader: bytes.NewReader(encoded.Bytes())}
	imageClient := &http.Client{Transport: faceRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		if len(req.Cookies()) != 0 {
			t.Fatalf("history face image request forwarded cookies: %v", req.Cookies())
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: imageBody}, nil
	})}

	_, img, err := (&XueXiTUserCache{}).getHistoryFaceImg("user-1", metadataClient, imageClient)
	if err != nil {
		t.Fatalf("getHistoryFaceImg() error = %v", err)
	}
	if img == nil {
		t.Fatal("getHistoryFaceImg() image = nil")
	}
	if !imageBody.closed {
		t.Fatal("getHistoryFaceImg() did not close image response body")
	}
}

func TestGetHistoryFaceImgWrapsDecodeErrorAndClosesBody(t *testing.T) {
	metadataClient := &http.Client{Transport: faceRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return faceResponse(`{"result":1,"data":{"http":"https://passport2-api.chaoxing.com/face.png"}}`), nil
	})}
	imageBody := &trackedReadCloser{Reader: strings.NewReader("not an image")}
	imageClient := &http.Client{Transport: faceRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: imageBody}, nil
	})}

	_, _, err := (&XueXiTUserCache{}).getHistoryFaceImg("user-1", metadataClient, imageClient)
	if err == nil || !strings.Contains(err.Error(), "图片解码失败") {
		t.Fatalf("getHistoryFaceImg() error = %v, want decode context", err)
	}
	if !imageBody.closed {
		t.Fatal("getHistoryFaceImg() did not close image response body after decode failure")
	}
}

func faceResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type faceRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn faceRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type trackedReadCloser struct {
	io.Reader
	closed bool
}

func (body *trackedReadCloser) Close() error {
	if body.closed {
		return fmt.Errorf("response body closed more than once")
	}
	body.closed = true
	return nil
}
