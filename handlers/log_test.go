package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	log.SetFlags(0)
	now := time.Now()

	formats := map[string]struct {
		fmt string
		rx  *regexp.Regexp
	}{
		"remote-addr": {
			"%s",
			regexp.MustCompile(`^127\.0\.0\.1:\d+\n$`),
		},
		"date": {
			"%s",
			regexp.MustCompile(`^` + fmt.Sprintf("%04d-%02d-%02d", now.Year(), now.Month(), now.Day()) + `\n$`),
		},
		"method": {
			"%s",
			regexp.MustCompile(`^GET\n$`),
		},
		"url": {
			"%s",
			regexp.MustCompile(`^/log\d+\n$`),
		},
		"http-version": {
			"%s",
			regexp.MustCompile(`^1\.1\n$`),
		},
		"status": {
			"%d",
			regexp.MustCompile(`^200\n$`),
		},
		"referer": {
			"%s",
			regexp.MustCompile(`^http://www\.test\.com\n$`),
		},
		"referrer": {
			"%s",
			regexp.MustCompile(`^http://www\.test\.com\n$`),
		},
		"user-agent": {
			"%s",
			regexp.MustCompile(`^Go \d+\.\d+ package http\n$`),
		},
		"bidon": {
			"%s",
			regexp.MustCompile(`^\?\n$`),
		},
		"response-time": {
			"%.3f",
			regexp.MustCompile(`^0\.1\d\d\n$`),
		},
		"req[Accept-Encoding]": {
			"%s",
			regexp.MustCompile(`^gzip\n$`),
		},
		"res[blah]": {
			"%s",
			regexp.MustCompile(`^$`),
		},
		"tiny": {
			Ltiny,
			regexp.MustCompile(`^GET /log\d+ 200  - 0\.1\d\d s\n$`),
		},
		"short": {
			Lshort,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - GET /log\d+ HTTP/1\.1 200  - 0\.1\d\d s\n$`),
		},
		"default": {
			Ldefault,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - - \[\d{4}-\d{2}-\d{2}\] "GET /log\d+ HTTP/1\.1" 200  "http://www\.test\.com" "Go \d+\.\d+ package http"\n$`),
		},
		// TODO : The next test fails, because only headers explicitly set through
		// ResponseWriter.Header are available (not those written with the actual response)
		/*"res[Content-Length]": {
			"%s",
			regexp.MustCompile(`^text/plain\n$`),
		},*/
	}
	cnt := 0
	for k, v := range formats {
		cnt++
		buf := bytes.NewBuffer(nil)
		log.SetOutput(buf)
		opts := NewLogOptions(nil, v.fmt, k)
		opts.DateFormat = "2006-01-02"
		h := LogHandler(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(200)
				w.Write([]byte("body"))
			}), opts)
		path := fmt.Sprintf("http://localhost%s/log%d", svrAddr, cnt)
		startServer(h, fmt.Sprintf("/log%d", cnt))

		t.Logf("running %s...", k)
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Referer", "http://www.test.com")
		req.Header.Set("Accept-Encoding", "gzip")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		assertStatus(http.StatusOK, res.StatusCode, t)
		ac := buf.String()
		assertTrue(v.rx.MatchString(ac), fmt.Sprintf("expected log to match '%s', got '%s'", v.rx.String(), ac), t)
	}
}

// TODO : Tests for Immediate = true and CustomTokens
