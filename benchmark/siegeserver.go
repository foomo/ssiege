package benchmark

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

type Service struct {
	Siege    *Siege
	Analyzer *Analyzer
}

func NewService(siege *Siege) *Service {
	s := &Service{
		Siege:    siege,
		Analyzer: NewAnalyzer(siege),
	}
	return s
}

const routeAPI = "/"

func jsonEncode(data interface{}) []byte {
	jsonBytes, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		panic(err)
	}
	return jsonBytes
}

func jsonReply(w http.ResponseWriter, data interface{}) {
	w.Write(jsonEncode(data))
}

func a(name string) []byte {
	data, err := Asset("benchmark/frontend" + name)
	if err != nil {
		panic(err)
	}
	return data
}

func (s *Service) ServeHTTP(w http.ResponseWriter, incomingRequest *http.Request) {
	//log.Println(incomingRequest.URL.Path)
	cmd := strings.TrimPrefix(incomingRequest.URL.Path, "/")
	switch true {
	case strings.HasPrefix(cmd, "asset/"):
		mime := "octect/stream"
		switch true {
		case strings.HasSuffix(cmd, ".css"):
			mime = "text/css"
		case strings.HasSuffix(cmd, ".js"):
			mime = "application/javascript"
		}
		w.Header().Add("Content-Type", mime)
		w.Write(a(cmd[5:]))
	case cmd == "":
		t := template.New("index")
		_, err := t.Parse(string(a("/templates/index.html")))
		if err != nil {
			panic(err)
		}
		g := s.Analyzer.Graph()
		w.Header().Add("Content-Type", "text/html;charset=utf-8;")
		t.Execute(w, &struct{ GraphJSON []byte }{GraphJSON: jsonEncode(g)})
	case cmd == "status":
		jsonReply(w, s.Analyzer.Graph())
	default:
		jsonReply(w, map[string]interface{}{
			"routes": []string{"status", "asset"},
		})
	}
	incomingRequest.Body.Close()
}
