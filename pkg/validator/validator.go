package validator

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/bocchi-the-cache/inspector/pkg/client"
	"github.com/bocchi-the-cache/inspector/pkg/common/logger"
	"github.com/bocchi-the-cache/inspector/pkg/common/result_logger"
	"github.com/bocchi-the-cache/inspector/pkg/monitor"
	"github.com/bocchi-the-cache/inspector/pkg/storage"
	"net/http"
	"strings"
	"sync"
	"time"
)

var DefaultValidator *Validator

func init() {
	DefaultValidator = &Validator{}
}

type Validator struct{}

func PushRequest(r *http.Request) {
	DefaultValidator.PushRequest(r)
}

func (v *Validator) PushRequest(r *http.Request) {
	if ok := v.CheckRequest(r); ok {
		go v.Validate(r)
	}
}

func (v *Validator) CheckRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	return true
}

// isIncompleteRangeRequest false when "" or "Range: <unit>=<range-start>-<range-end>"
func isIncompleteRangeRequest(r *http.Request) bool {
	if r.Header.Get("Range") == "" {
		return false
	}
	rg := strings.Trim(r.Header.Get("Range"), " ")
	if strings.Contains(rg, ",") || // eg: Range: bytes=0-1,2-3
		strings.Contains(rg, "=-") || // eg: Range: bytes=-1
		strings.HasSuffix(rg, "-") { // eg: Range: bytes=0-
		logger.Info("incompleted range request, range: %s", r.Header.Get("Range"))
		return true
	}
	return false
}

func (v *Validator) Validate(r *http.Request) {
	defer handlePanic()
	monitor.RequestReceiveTotalCounterIncr(r.Method, r.Host)

	//host := r.Host // eg: localhost:4399
	//url := r.URL   // eg: /blabla/123/abc.txt

	wg := sync.WaitGroup{}
	var BaselineContent *client.Content
	var errBaseline error
	var TestContent *client.Content
	var errTest error

	// Get Baseline Content
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.Now()
		BaselineContent, errBaseline = GetBaselineContent(r)
		elapsed := time.Since(t)
		monitor.ElapsedMonitorIncr("BaselineFetch", float64(elapsed/10e6))
	}()

	// Get Test Content
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.Now()
		TestContent, errTest = GetTestContent(r)
		elapsed := time.Since(t)
		monitor.ElapsedMonitorIncr("TestFetch", float64(elapsed/10e6))
	}()

	wg.Wait()
	if errBaseline != nil {
		logger.Errorf("get baseline content error, err: %s", errBaseline)
		monitor.ErrorTotalCounterIncr("GetContent", "baseline", "errBaseline")
	}
	if errTest != nil {
		logger.Errorf("get test content error, err: %s", errTest)
		monitor.ErrorTotalCounterIncr("GetContent", "test", "errTest")
	}

	// Generate Report
	t := time.Now()
	CheckContentAndReport(r, BaselineContent, errBaseline, TestContent, errTest)
	elapsed := time.Since(t)
	monitor.ElapsedMonitorIncr("ContentCompare", float64(elapsed/10e6))
}

func GetBaselineContent(r *http.Request) (*client.Content, error) {
	// Don't Find in cache
	status, header, content, err := client.BaselineFetcher.Do(r)
	if err != nil {
		return nil, err
	}
	return &client.Content{
		Status:  status,
		Header:  header,
		Content: content,
	}, nil
}

func GetTestContent(r *http.Request) (*client.Content, error) {
	status, header, content, err := client.TestFetcher.Do(r)
	if err != nil {
		return nil, err
	}
	return &client.Content{
		Status:  status,
		Header:  header,
		Content: content,
	}, nil
}

func CheckContentAndReport(r *http.Request, BaselineContent *client.Content, errBaseline error, TestContent *client.Content, errTest error) {
	state := "PASS"

	if errBaseline != nil || errTest != nil {
		// 1. Primary Error Check
		state = "FETCH_ERROR"
	} else if BaselineContent.Status != TestContent.Status {
		// 2. Status Check
		state = "STATUS_NOT_MATCH"
	} else {
		if BaselineContent.Content == nil || TestContent.Content == nil {
			// 3.1 Empty Content Check
			state = "EMPTY_CONTENT"
		} else if (BaselineContent.Status != http.StatusOK) && (BaselineContent.Status != http.StatusPartialContent) {
			// 3.2 Skip if status is not 200
			state = "STATUS_NOT_200/206_SKIP"
		} else {
			// 3.3 Content Check
			ok := compareContent(BaselineContent, TestContent)
			if !ok {
				state = "CONTENT_NOT_MATCH"
				go func() {
					path := r.URL.Path + ".test"
					err := storage.Write(path, TestContent.Content)
					if err != nil {
						logger.Errorf("write test content error, path: %s, err: %s", path, err)
					}
				}()
				go func() {
					path := r.URL.Path + ".baseline"
					err := storage.Write(path, BaselineContent.Content)
					if err != nil {
						logger.Errorf("write baseline content error, path: %s, err: %s", path, err)
					}
				}()
			}
		}
	}
	var baselineHash string
	var testHash string
	if BaselineContent != nil && BaselineContent.Content != nil {
		h := md5.New()
		h.Write(BaselineContent.Content)
		baselineHash = hex.EncodeToString(h.Sum(nil))
	}
	if TestContent != nil && TestContent.Content != nil {
		h := md5.New()
		h.Write(TestContent.Content)
		testHash = hex.EncodeToString(h.Sum(nil))
	}

	// process result
	var baselineHeader, testHeader http.Header
	var baselineStatus, testStatus int

	if BaselineContent == nil {
		baselineHeader = http.Header{}
		baselineStatus = 0
	} else {
		baselineHeader = BaselineContent.Header
		baselineStatus = BaselineContent.Status
	}
	if TestContent == nil {
		testHeader = http.Header{}
		testStatus = 0
	} else {
		testHeader = TestContent.Header
		testStatus = TestContent.Status
	}

	monitor.ResultTotalCounterIncr("ContentCompare", state)
	result_logger.Infof("[REQ]\t%v\t%v\t%v\t%v\t%v\t%v\t%v \t%v\t%v\t%v\t%v ", state, r.Host, r.URL.Path,
		errBaseline, baselineStatus, baselineHash, baselineHeader,
		errTest, testStatus, testHash, testHeader)
}

func compareContent(b *client.Content, t *client.Content) bool {
	if len(b.Content) != len(t.Content) {
		return false
	}
	for i := 0; i < len(b.Content); i++ {
		if b.Content[i] != t.Content[i] {
			return false
		}
	}
	return true
}
