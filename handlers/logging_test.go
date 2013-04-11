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

func TestLogging(t *testing.T) {
	path := fmt.Sprintf("http://localhost%s/logging", svrAddr)
	log.SetFlags(0)
	now := time.Now()

	h := NewLoggingHandler(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(200)
			w.Write([]byte("body"))
		}), nil, "")
	startServer(h, "/logging")

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
			regexp.MustCompile(`^/logging\n$`),
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
			regexp.MustCompile(`^GET /logging 200  - 0\.1\d\d s\n$`),
		},
		"short": {
			Lshort,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - GET /logging HTTP/1\.1 200  - 0\.1\d\d s\n$`),
		},
		"default": {
			Ldefault,
			regexp.MustCompile(`^127\.0\.0\.1:\d+ - - \[\d{4}-\d{2}-\d{2}\] "GET /logging HTTP/1\.1" 200  "http://www\.test\.com" "Go \d+\.\d+ package http"\n$`),
		},
		// TODO : The next test fails, because only headers explicitly set through
		// ResponseWriter.Header are available (not those written with the actual response)
		/*"res[Content-Length]": {
			"%s",
			regexp.MustCompile(`^text/plain\n$`),
		},*/
	}
	for k, v := range formats {
		h.Format = v.fmt
		h.Tokens = []string{k}
		h.DateFormat = "2006-01-02"
		buf := bytes.NewBuffer(nil)
		log.SetOutput(buf)

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
