package server

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/render"
	"github.com/devopsext/utils"
	"github.com/go-playground/form/v4"
)

type HttpServerOptions struct {
	ServerName string
	Listen     string
	Tls        bool
	Insecure   bool
	Cert       string
	Key        string
	Chain      string
	Timeout    int
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
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
		return err
	}
	return nil
}

// HttpServerCallProcessor

func (h *HttpServerCallProcessor) Path() string {
	return HttpServerCallProcessorPath
}

func (h *HttpServerCallProcessor) handleTemplate(request *HttpServerCallRequest) ([]interface{}, error) {

	options := render.TemplateOptions{
		Content:     "{{ $d := 0 }}",
		FilterFuncs: false,
	}
	tpl, err := render.NewTextTemplate(options, h.server.logger)
	if err != nil {
		return nil, nil
	}
	name := strings.ToUpper(request.Name[:1]) + request.Name[1:]
	return common.Invoke(tpl, name, request.Params...)
}

func (h *HttpServerCallProcessor) HandleRequest(w http.ResponseWriter, r *http.Request) error {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("could not parse form: %v", err), http.StatusInternalServerError)
		return err
	}

	decoder := form.NewDecoder()

	var request HttpServerCallRequest

	err = decoder.Decode(&request, r.Form)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not decode form: %v", err), http.StatusInternalServerError)
		return err
	}

	h.server.logger.Debug("%v", request)

	if utils.IsEmpty(request.Name) {
		err := fmt.Errorf("name is empty")
		http.Error(w, fmt.Sprintf("could not decode form: %v", err), http.StatusInternalServerError)
		return err
	}

	var arr []interface{}

	switch request.Package {
	case "template":
		arr, err = h.handleTemplate(&request)
	default:
		arr, err = h.handleTemplate(&request)
	}

	var rerr string
	if err != nil {
		rerr = err.Error()
	}

	var rarr []interface{}
	if len(arr) > 0 {
		rarr = arr
	}

	res := &HttpServerCallRespone{
		Request: &request,
		Result:  rarr,
		Error:   rerr,
	}

	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return err
	}

	if _, err := w.Write(data); err != nil {
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
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
		h.logger.Info("Start http server...")

		var caPool *x509.CertPool
		var certificates []tls.Certificate

		if h.options.Tls {

			// load certififcate
			var cert []byte
			if _, err := os.Stat(h.options.Cert); err == nil {

				cert, err = os.ReadFile(h.options.Cert)
				if err != nil {
					h.logger.Panic(err)
				}
			} else {
				cert = []byte(h.options.Cert)
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

			// load CA chain
			var chain []byte
			if _, err := os.Stat(h.options.Chain); err == nil {
				chain, err = os.ReadFile(h.options.Chain)
				if err != nil {
					h.logger.Panic(err)
				}
			} else {
				chain = []byte(h.options.Chain)
			}

			// make pool of chains
			caPool = x509.NewCertPool()
			if !caPool.AppendCertsFromPEM(chain) {
				h.logger.Debug("CA chain is invalid")
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

		h.logger.Info("Http server is up. Listening...")

		srv := &http.Server{
			Handler:  mux,
			ErrorLog: nil,
		}

		if h.options.Tls {

			srv.TLSConfig = &tls.Config{
				Certificates:       certificates,
				RootCAs:            caPool,
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
