package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/render"
	"github.com/devopsext/utils"
	"github.com/go-playground/form/v4"
)

type HttpServerOptions struct {
	ServerName      string
	Listen          string
	Tls             bool
	Insecure        bool
	CA              string
	Crt             string
	Key             string
	Timeout         int
	Methods         []string
	SensitiveFields []string
}

type HttpServer struct {
	options HttpServerOptions
	logger  common.Logger
}

type HttpServerProcessor interface {
	Path() string
	HandleRequest(w http.ResponseWriter, r *http.Request) error
}

type HttpServerHealthProcessor struct {
	server *HttpServer
}

type HttpServerCallRequest struct {
	ID      string        `form:"id"`
	Name    string        `form:"name"`
	Package string        `form:"package,omitempty"`
	Params  []interface{} `form:"params,omitempty"`
	Timeout int           `form:"timeout,omitempty"`
}

type HttpServerCallRespone struct {
	Request *HttpServerCallRequest `json:"request"`
	Result  []interface{}          `json:"result,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type HttpServerCallProcessor struct {
	server *HttpServer
}

const (
	HttpServerHealthProcessorPath = "/health"
	HttpServerCallProcessorPath   = "/call"
)

// HttpServerHealthProcessor

func (h *HttpServerHealthProcessor) Path() string {
	return HttpServerHealthProcessorPath
}

func (h *HttpServerHealthProcessor) HandleRequest(w http.ResponseWriter, r *http.Request) error {

	_, err := w.Write([]byte("OK"))

	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP Server could not write response: %v", err), http.StatusInternalServerError)
		return err
	}
	return nil
}

// HttpServerCallProcessor

func (h *HttpServerCallProcessor) Path() string {
	return HttpServerCallProcessorPath
}

func (h *HttpServerCallProcessor) request2String(request *HttpServerCallRequest) string {

	pkg := ""
	if !utils.IsEmpty(request.Package) {
		pkg = fmt.Sprintf(" package: %s", request.Package)
	}
	return fmt.Sprintf("name: %s%s params: %d timeout: %d", request.Name, pkg, len(request.Params), request.Timeout)
}

func (h *HttpServerCallProcessor) replaceByRegex(s, key string) string {

	var re = regexp.MustCompile(fmt.Sprintf(`(%s:.+?)( |})`, key))
	rep := fmt.Sprintf("%s:%s ", key, strings.Repeat("*", 8))
	return re.ReplaceAllString(s, rep)
}

func (h *HttpServerCallProcessor) params2String(params []interface{}) string {

	s := fmt.Sprintf("%v", params)

	for _, v := range h.server.options.SensitiveFields {
		s = h.replaceByRegex(s, v)
	}

	return fmt.Sprintf("params: %s", s)
}

func (h *HttpServerCallProcessor) handleTemplate(name string, params []interface{}) ([]interface{}, error) {

	options := render.TemplateOptions{
		Content:     "{{ $d := 0 }}",
		FilterFuncs: false,
	}
	tpl, err := render.NewTextTemplate(options, h.server.logger)
	if err != nil {
		return nil, nil
	}
	return common.Invoke(tpl, name, params...)
}

func (h *HttpServerCallProcessor) HandleRequest(w http.ResponseWriter, r *http.Request) error {

	var err error
	if !utils.Contains(h.server.options.Methods, r.Method) {
		err := fmt.Errorf("HTTP Server has invalid method: %v", r.Method)
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return err
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP Server could not parse form: %v", err), http.StatusInternalServerError)
		return err
	}

	decoder := form.NewDecoder()
	var request HttpServerCallRequest

	err = decoder.Decode(&request, r.Form)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP Server could not decode form: %v", err), http.StatusInternalServerError)
		return err
	}

	h.server.logger.Debug("HTTP Server reguest id: %s => %s", request.ID, h.request2String(&request))

	if utils.IsEmpty(request.Name) {
		err := fmt.Errorf("name is empty")
		http.Error(w, fmt.Sprintf("HTTP Server could not decode form: %v", err), http.StatusInternalServerError)
		return err
	}

	var arr []interface{}
	var params []interface{}

	if len(request.Params) > 0 {
		for _, v := range request.Params {

			if utils.IsEmpty(v) {
				continue
			}

			s, ok := v.(string)
			if ok {
				// try as map[string]interface{}
				var m map[string]interface{}
				err := json.Unmarshal([]byte(s), &m)
				if err == nil {
					params = append(params, m)
					continue
				}

				// try as []string
				var sa []string
				err = json.Unmarshal([]byte(s), &sa)
				if err == nil {
					params = append(params, sa)
					continue
				}

				// try as []interface{}
				var ia []string
				err = json.Unmarshal([]byte(s), &ia)
				if err == nil {
					params = append(params, ia)
					continue
				}
			}

			params = append(params, v)
		}
	}

	h.server.logger.Debug("HTTP Server request id: %s => %s", request.ID, h.params2String(params))

	name := strings.ToUpper(request.Name[:1]) + request.Name[1:]

	switch request.Package {
	case "template":
		arr, err = h.handleTemplate(name, params)
	default:
		arr, err = h.handleTemplate(name, params)
	}

	var rerr string
	if err != nil {
		rerr = err.Error()
	}

	var rarr []interface{}
	if len(arr) > 0 {

		for _, v := range arr {
			switch v.(type) {
			case []byte:

				var i interface{}
				err := json.Unmarshal(v.([]byte), &i)
				if err != nil {
					rarr = append(rarr, v)
					continue
				}
				rarr = append(rarr, i)
			default:
				rarr = append(rarr, v)
			}
		}
	}

	res := &HttpServerCallRespone{
		Request: &request,
		Result:  rarr,
		Error:   rerr,
	}

	serr := ""
	if !utils.IsEmpty(rerr) {
		serr = fmt.Sprintf(" error: %s", rerr)
	}

	sarr := "no result"
	if !utils.IsEmpty(rarr) {
		sarr = fmt.Sprintf("result: %v", rarr)
	}

	h.server.logger.Debug("HTTP Server request id: %s => %s%s", request.ID, sarr, serr)

	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP Server could not marshal response: %v", err), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write(data); err != nil {
		http.Error(w, fmt.Sprintf("HTTP Server could not write response: %v", err), http.StatusInternalServerError)
		return err
	}

	return nil
}

// HttpServer

func (h *HttpServer) processURL(url string, mux *http.ServeMux, p HttpServerProcessor) {

	urls := strings.Split(url, ",")
	for _, url := range urls {

		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {

			err := p.HandleRequest(w, r)
			if err != nil {
				h.logger.Error(err)
			}
		})
	}
}

func (h *HttpServer) Start(wg *sync.WaitGroup) {

	wg.Add(1)
	go func(wg *sync.WaitGroup) {

		defer wg.Done()
		h.logger.Info("Start HTTP Server...")

		var caPool *x509.CertPool
		var certificates []tls.Certificate

		if h.options.Tls {

			// load certififcate
			var cert []byte
			if _, err := os.Stat(h.options.Crt); err == nil {

				cert, err = os.ReadFile(h.options.Crt)
				if err != nil {
					h.logger.Panic(err)
				}
			} else {
				cert = []byte(h.options.Crt)
			}

			// load key
			var key []byte
			if _, err := os.Stat(h.options.Key); err == nil {
				key, err = os.ReadFile(h.options.Key)
				if err != nil {
					h.logger.Panic(err)
				}
			} else {
				key = []byte(h.options.Key)
			}

			// make pair from certificate and pair
			pair, err := tls.X509KeyPair(cert, key)
			if err != nil {
				h.logger.Panic(err)
			}

			certificates = append(certificates, pair)

			// load CA
			var ca []byte
			if _, err := os.Stat(h.options.CA); err == nil {
				ca, err = os.ReadFile(h.options.CA)
				if err != nil {
					h.logger.Panic(err)
				}
			} else {
				ca = []byte(h.options.CA)
			}

			// make pool of CA
			caPool = x509.NewCertPool()
			if !caPool.AppendCertsFromPEM(ca) {
				h.logger.Debug("HTTP Server CA is invalid")
			}
		}

		mux := http.NewServeMux()

		processors := h.getProcessors()
		for u, p := range processors {
			h.processURL(u, mux, p)
		}

		listener, err := net.Listen("tcp", h.options.Listen)
		if err != nil {
			h.logger.Panic(err)
		}

		h.logger.Info("HTTP Server is up. Listening...")

		srv := &http.Server{
			Handler:  mux,
			ErrorLog: nil,
		}

		if h.options.Tls {

			srv.TLSConfig = &tls.Config{
				ClientAuth:         tls.RequireAndVerifyClientCert,
				ClientCAs:          caPool,
				Certificates:       certificates,
				InsecureSkipVerify: h.options.Insecure,
				ServerName:         h.options.ServerName,
			}

			err = srv.ServeTLS(listener, "", "")
			if err != nil {
				h.logger.Panic(err)
			}
		} else {
			err = srv.Serve(listener)
			if err != nil {
				h.logger.Panic(err)
			}
		}
	}(wg)
}

func (h *HttpServer) getProcessors() map[string]HttpServerProcessor {

	m := make(map[string]HttpServerProcessor)
	m[HttpServerHealthProcessorPath] = &HttpServerHealthProcessor{server: h}
	m[HttpServerCallProcessorPath] = &HttpServerCallProcessor{server: h}
	return m
}

func NewHttpServer(options HttpServerOptions, logger common.Logger) *HttpServer {

	server := &HttpServer{
		options: options,
		logger:  logger,
	}
	return server
}
