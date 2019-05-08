// Package icy contains icy title decoder.
package icy

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
)

var (
	streamTitle  = []byte("StreamTitle='")
	titleTimeout = 30 * time.Minute
)

// Opens icy stream and searches for the song title. Song and title will be pushed to the titleChannel.
func Open(urlString string, log string, titleChannel chan string) error {
	glog.V(1).Infof("[%15.15s] Starting stream %q...", log, urlString)

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	lastTitle := ""
	lastTitleTime := time.Now()

	for err == nil {
		var b []byte
		b, err = reader.ReadBytes(';')
		glog.V(10).Infof("%v", b)
		t := findStreamTitle(b)
		if t != nil && *t != "" && *t != lastTitle {
			glog.V(1).Infof("[%15.15s] New title found: %q.", log, *t)
			titleChannel <- *t
			lastTitle = *t
			lastTitleTime = time.Now()
		}
		if lastTitleTime.Add(titleTimeout).Before(time.Now()) {
			err = fmt.Errorf("title timeout, last title found: %v", lastTitleTime.Format("2006-01-02 15:04:05"))
		}
	}
	return err
}

func findStreamTitle(b []byte) *string {
	for i := 0; i < len(b)-len(streamTitle); i++ {
		if streamTitle[0] == b[i] {
			if isStreamTitle(b, i) {
				// StreamTitle='Artist - Title';
				// Remove "StreamTitle='" and
				// last two characters (';)
				res := string(b[i+len(streamTitle) : len(b)-2])
				return &res
			}
		}
	}
	return nil
}

func isStreamTitle(b []byte, pos int) bool {
	for i := 0; i < len(streamTitle); i++ {
		if streamTitle[i] != b[pos+i] {
			return false
		}
	}
	return true
}