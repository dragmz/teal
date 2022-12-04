package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strconv"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type args struct {
}

type jsonRpcHeader struct {
	JsonRpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Method  string `json:"method"`
}

type jsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error,omitempty"`
	Id      int         `json:"id"`
}

type lspDiagnosticProvider struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

type lspTextDocumentSync struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
}

type lspServerCapabilities struct {
	TextDocumentSync   int                    `json:"textDocumentSync,omitempty"`
	DiagnosticProvider *lspDiagnosticProvider `json:"diagnosticProvider,omitempty"`
}

type lspInitializeResult struct {
	Capabilities *lspServerCapabilities `json:"capabilities"`
}

type lsp struct {
	docs map[string]string
}

type lspDidOpenTextDocument struct {
	Uri     string `json:"uri"`
	Version int    `json:"version"`
	Text    string `json:"text"`
}

type lspDidOpenParams struct {
	TextDocument *lspDidOpenTextDocument `json:"textDocument"`
}

type lspDidChangeTextDocument struct {
	Uri     string `json:"uri"`
	Version int    `json:"version"`
}

type lspContentChange struct {
	Text string `json:"text"`
}

type lspDidChangeParams struct {
	TextDocument   *lspDidChangeTextDocument `json:"textDocument"`
	ContentChanges []*lspContentChange       `json:"contentChanges"`
}

type lspDidChange struct {
	Params *lspDidChangeParams `json:"params"`
}

type lspDidOpen struct {
	Params *lspDidOpenParams `json:"params"`
}

type lspDiagnosticRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspDiagnosticRequestParams struct {
	TextDocument *lspDiagnosticRequestTextDocument `json:"textDocument"`
}

type lspDiagnosticRequest struct {
	Params *lspDiagnosticRequestParams `json:"params"`
}

type lspDidCloseTextDocument struct {
	Uri string `json:"uri"`
}

type lspDidCloseRequestParams struct {
	TextDocument *lspDidCloseTextDocument `json:"textDocument"`
}

type lspDidCloseRequest struct {
	Params *lspDidCloseRequestParams `json:"params"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspDiagnostic struct {
	Range    lspRange `json:"range"`
	Severity *int     `json:"severity,omitempty"`
	Message  string   `json:"message"`
}

type lspPublishDiagnostic struct {
	Uri         string          `json:"uri"`
	Diagnostics []lspDiagnostic `json:"diagnostics"`
}

type lspNotification struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

func (l *lsp) handle(h jsonRpcHeader, b []byte) (interface{}, error) {
	doDiagnostic := func(uri string) (interface{}, error) {
		text := l.docs[uri]
		p, err := teal.Parse(text)
		if err != nil {
			return nil, err
		}

		diags := []lspDiagnostic{}

		ls := teal.Compile(p)
		for _, le := range ls.Lint() {
			sev := new(int)
			*sev = 2

			diags = append(diags, lspDiagnostic{
				Range: lspRange{
					Start: lspPosition{
						Line: le.Line() - 1,
					},
					End: lspPosition{
						Line: le.Line() - 1,
					},
				},
				Severity: sev,
				Message:  fmt.Sprintf("%s", le),
			})
		}

		return &lspNotification{
			JsonRpc: "2.0",
			Method:  "textDocument/publishDiagnostics",
			Params: &lspPublishDiagnostic{
				Uri:         uri,
				Diagnostics: diags,
			},
		}, nil
	}

	switch h.Method {
	case "initialized":
	case "textDocument/didSave":
	case "$/cancelRequest":
	case "shutdown":

	case "textDocument/didClose":
		var req lspDidCloseRequest
		err := json.Unmarshal(b, &req)
		if err != nil {
			return nil, err
		}

		delete(l.docs, req.Params.TextDocument.Uri)

	case "textDocument/diagnostic":
		var req lspDiagnosticRequest
		err := json.Unmarshal(b, &req)
		if err != nil {
			return nil, err
		}

		return doDiagnostic(req.Params.TextDocument.Uri)

	case "textDocument/didOpen":
		var req lspDidOpen
		err := json.Unmarshal(b, &req)
		if err != nil {
			return nil, err
		}
		l.docs[req.Params.TextDocument.Uri] = req.Params.TextDocument.Text

		return doDiagnostic(req.Params.TextDocument.Uri)

	case "textDocument/didChange":
		var req lspDidChange
		err := json.Unmarshal(b, &req)
		if err != nil {
			return nil, err
		}
		for _, ch := range req.Params.ContentChanges {
			l.docs[req.Params.TextDocument.Uri] = ch.Text
		}

		return doDiagnostic(req.Params.TextDocument.Uri)

	case "initialize":
		return &jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result: &lspInitializeResult{
				Capabilities: &lspServerCapabilities{
					TextDocumentSync: 1,
					DiagnosticProvider: &lspDiagnosticProvider{
						InterFileDependencies: false,
						WorkspaceDiagnostics:  false,
					},
				},
			},
		}, nil
	default:
		return nil, errors.New("unknown method")
	}

	return nil, nil
}

func (l *lsp) run(rd io.Reader, wr io.Writer) error {
	r := bufio.NewReader(rd)
	tp := textproto.NewReader(r)

	for {
		mh, err := tp.ReadMIMEHeader()
		if err != nil {
			return err
		}

		h := http.Header(mh)

		length, err := strconv.Atoi(h.Get("Content-Length"))
		if err != nil {
			return err
		}

		data := make([]byte, length)
		_, err = io.ReadFull(tp.R, data)
		if err != nil {
			return err
		}

		var jh jsonRpcHeader
		err = json.Unmarshal(data, &jh)
		if err != nil {
			return err
		}

		go func() error {
			resp, err := l.handle(jh, data)
			if err != nil {
				return err
			}

			if resp != nil {
				rb, err := json.Marshal(resp)
				if err != nil {
					return err
				}

				h := http.Header{}
				h.Set("Content-Length", strconv.Itoa(len(rb)))

				err = h.Write(wr)
				if err != nil {
					return err
				}

				_, err = wr.Write([]byte("\r\n"))
				if err != nil {
					return err
				}

				_, err = wr.Write(rb)
				if err != nil {
					return err
				}
			}

			return nil
		}()
	}
}

func run(a args) error {
	l := &lsp{
		docs: map[string]string{},
	}

	err := l.run(os.Stdin, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	var a args
	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
