package benchmark

import (
	"net/http"
	"time"
)

type Call struct {
	Method   string `json:"method"`
	URI      string `json:"uri"`
	Data     string `json:"data"`
	Mimetype string `json:"mimetype"`
	Comment  string `json:"comment"`
	Debug    bool   `json:"debug"`
	Wait     int64  `json:"wait"`
}

type CallStats struct {
	Start    time.Time
	Duration time.Duration
}

type Session struct {
	Server string  `json:"server"`
	Calls  []*Call `json:"calls"`
}

type SessionStats struct {
	Calls []*CallStats
}

type SessionRunner struct {
	Session   *Session
	Transport *http.Transport
}

func NewSessionRunner(session *Session, transport *http.Transport) *SessionRunner {
	sr := &SessionRunner{
		Session:   session,
		Transport: transport,
	}
	return sr
}

func (sr *SessionRunner) Run() *SessionStats {
	client := NewClient(sr.Session.Server, sr.Transport)
	client.ResetClient()
	sessionStats := &SessionStats{
		Calls: []*CallStats{},
	}
	for i, call := range sr.Session.Calls {
		sessionStats.Calls = append(sessionStats.Calls, client.Execute(call, i < len(sr.Session.Calls)-1))
	}
	return sessionStats
}
