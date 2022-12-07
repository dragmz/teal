package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

const (
	semanticTokenKeyword = 0
	semanticTokenString  = 1
	semanticTokenComment = 2
	semanticTokenMethod  = 3
	semanticTokenMacro   = 4
)

type lsp struct {
	docs     map[string]string
	shutdown bool
	exit     bool

	tp *textproto.Reader
	w  *bufio.Writer

	debug *bufio.Writer
}

type LspOption func(l *lsp) error

func WithDebug(w io.Writer) LspOption {
	return func(l *lsp) error {
		l.debug = bufio.NewWriter(w)
		return nil
	}
}

func New(r io.Reader, w io.Writer, opts ...LspOption) (*lsp, error) {
	l := &lsp{
		tp:   textproto.NewReader(bufio.NewReader(r)),
		w:    bufio.NewWriter(w),
		docs: map[string]string{},
	}

	for _, opt := range opts {
		err := opt(l)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set lsp option")
		}
	}

	return l, nil
}

type jsonRpcHeader struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Method  string      `json:"method"`
}

type jsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error,omitempty"`
	Id      interface{} `json:"id"`
}

type lspFullDocumentDiagnosticReport struct {
	Kind  string          `json:"kind"`
	Items []lspDiagnostic `json:"items"`
}

type lspRequest[T any] struct {
	Params T `json:"params"`
}

type lspDiagnosticProvider struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

type lspCompletionItem struct {
	LabelDetailsSupport *bool `json:"labelDetailsSupport,omitempty"`
}

type lspCompletionProvider struct {
	TriggerCharacters   []string           `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string           `json:"allCommitCharacters,omitempty"`
	ResolveProvider     *bool              `json:"resolveProvider,omitempty"`
	CompletionItem      *lspCompletionItem `json:"completionItem,omitempty"`
}

type lspExecuteCommandProvider struct {
	Commands []string `json:"commands"`
}

type lspError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type lspSemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

type lspSemanticTokensProvider struct {
	Legend lspSemanticTokensLegend `json:"legend"`
	Range  *bool                   `json:"range"`
	Full   *bool                   `json:"full"`
}

type lspSemanticTokens struct {
	Data []uint32 `json:"data"`
}

type lspServerCapabilities struct {
	TextDocumentSync          *int                       `json:"textDocumentSync,omitempty"`
	DiagnosticProvider        *lspDiagnosticProvider     `json:"diagnosticProvider,omitempty"`
	CompletionProvider        *lspCompletionProvider     `json:"completionProvider,omitempty"`
	DocumentSymbolProvider    *bool                      `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider        *bool                      `json:"codeActionProvider,omitempty"`
	ExecuteCommandProvider    *lspExecuteCommandProvider `json:"executeCommandProvider,omitempty"`
	RenameProvider            *lspRenameOptions          `json:"renameProvider,omitempty"`
	ColorProvider             *bool                      `json:"colorProvider,omitempty"`
	DocumentHighlightProvider *bool                      `json:"documentHighlightProvider,omitempty"`
	SemanticTokensProvider    *lspSemanticTokensProvider `json:"semanticTokensProvider,omitempty"`
}

type lspInitializeResult struct {
	Capabilities *lspServerCapabilities `json:"capabilities"`
}

type lspSymbolKind int

const (
	lspSymbolKindMethod   = 6
	lspSymbolKindOperator = 25
)

type lspDocumentSymbol struct {
	Name           string        `json:"name"`
	Kind           lspSymbolKind `json:"kind"`
	Range          lspRange      `json:"range"`
	SelectionRange lspRange      `json:"selectionRange"`
}

type lspInitializeClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type lspInitializeRequestParams struct {
	ProcessId  int                      `json:"id"`
	ClientInfo *lspInitializeClientInfo `json:"clientInfo"`
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

type lspDidSaveTextDocument struct {
	Uri string `json:"uri"`
}

type lspDidSaveParams struct {
	TextDocument *lspDidSaveTextDocument `json:"textDocument"`
}

type lspDocumentSymbolTextDocument struct {
	Uri string `json:"uri"`
}

type lspDocumentSymbolParams struct {
	TextDocument *lspDocumentSymbolTextDocument `json:"textDocument"`
}

type tealRenameCommandArgs struct {
	Range lspRange `json:"range"`
}

type lspWorkspaceExecuteCommandHeader struct {
	Command string `json:"command"`
}

type lspWorkspaceExecuteCommandBody[T any] struct {
	Params T `json:"params"`
}

type lspWorkspaceExecuteCommand lspRequest[*lspWorkspaceExecuteCommandHeader]

// notifications
type lspDidChange lspRequest[*lspDidChangeParams]
type lspDidOpen lspRequest[*lspDidOpenParams]
type lspDidSave lspRequest[*lspDidSaveParams]

// requests
type lspDocumentSymbolRequest lspRequest[*lspDocumentSymbolParams]

type lspDiagnosticRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspDiagnosticRequestParams struct {
	TextDocument *lspDiagnosticRequestTextDocument `json:"textDocument"`
}

type lspDiagnosticRequest lspRequest[*lspDiagnosticRequestParams]

type lspCodeActionTextDocument struct {
	Uri string `json:"uri"`
}

type lspCodeActionContext struct {
	Diagnosticts []lspDiagnostic `json:"diagnostics"`
	Only         []string        `json:"only,omitempty"`
	TriggerKind  *int            `json:"triggerKind,omitempty"`
}

type lspCodeActionRequestParams struct {
	TextDocument lspCodeActionTextDocument `json:"textDocument"`
	Range        lspRange                  `json:"range"`
	Context      lspCodeActionContext      `json:"context"`
}

type lspCodeActionRequest lspRequest[*lspCodeActionRequestParams]

type lspRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspRenameRequestParams struct {
	TextDocument lspRenameRequestTextDocument `json:"textDocument"`
	Position     lspPosition
	NewName      string `json:"newName"`
}

type lspRenameRequest lspRequest[*lspRenameRequestParams]

type lspPrepareRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspPrepareRenameRequestParams struct {
	TextDocument lspPrepareRenameRequestTextDocument `json:"textDocument"`
	Position     lspPosition
}

type lspPrepareRenameRequest lspRequest[*lspPrepareRenameRequestParams]

type lspDocumentColorRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspColor struct {
	R float32 `json:"r"`
	G float32 `json:"g"`
	B float32 `json:"b"`
	A float32 `json:"a"`
}

type lspColorInformation struct {
	Range lspRange `json:"range"`
	Color lspColor `json:"color"`
}

type lspDocumentColorRequestParams struct {
	TextDocument lspDocumentColorRequestTextDocument `json:"textDocument"`
}

type lspDocumentColorRequest lspRequest[*lspDocumentColorRequestParams]

type lspPrepareRenameResponse struct {
	Range       lspRange `json:"range"`
	Placeholder string   `json:"placeholder"`
}

type lspCommand struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

type lspTextEdit struct {
	Range   lspRange `json:"range"`
	NewText string   `json:"newText"`
}

type lspWorkspaceEdit struct {
	Changes map[string][]lspTextEdit `json:"changes,omitempty"`
}

type lspCodeAction struct {
	Title       string            `json:"title"`
	Kind        *string           `json:"kind,omitempty"`
	Diagnostics []lspDiagnostic   `json:"diagnostics,omitempty"`
	IsPreferred *bool             `json:"isPreferred,omitempty"`
	Edit        *lspWorkspaceEdit `json:"edit,omitempty"`
	Command     *lspCommand       `json:"command,omitempty"`
}

type lspDidCloseTextDocument struct {
	Uri string `json:"uri"`
}

type lspDidCloseRequestParams struct {
	TextDocument *lspDidCloseTextDocument `json:"textDocument"`
}

type lspDidCloseRequest lspRequest[*lspDidCloseRequestParams]

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

type lspRenameOptions struct {
	PrepareProvider *bool `json:"prepareProvider,omitempty"`
}

type lspDocumentHighlightRequestTextDocument struct {
	Uri string `json:"uri`
}

type lspDocumentHighlightRequestParams struct {
	TextDocument lspDocumentHighlightRequestTextDocument `json:"textDocument"`
	Position     lspPosition                             `json:"position"`
}

type lspDocumentHighlightRequest lspRequest[*lspDocumentHighlightRequestParams]

type lspDocumentHighlight struct {
	Range lspRange `json:"range"`
	Kind  *int     `json:"kind"`
}

type lspTextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type lspSemanticTokensFullRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

type lspSemanticTokensFullRequest lspRequest[*lspSemanticTokensFullRequestParams]

func readInto(b []byte, v interface{}) error {
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}

func read[T any](b []byte) (T, error) {
	var v T

	err := readInto(b, &v)
	if err != nil {
		return v, err
	}

	return v, nil
}

func (l *lsp) notifyDiagnostics(uri string, lds []lspDiagnostic) error {
	return l.write(lspNotification{
		JsonRpc: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: &lspPublishDiagnostic{
			Uri:         uri,
			Diagnostics: lds,
		},
	})
}

func (l *lsp) doDiagnostic(uri string) []lspDiagnostic {
	text := l.docs[uri]

	lds := []lspDiagnostic{}

	res := teal.Process(text)
	for _, d := range res.Diagnostics {
		sev := int(d.Severity())

		lds = append(lds, lspDiagnostic{
			Range: lspRange{
				Start: lspPosition{
					Line:      d.Line(),
					Character: d.Begin(),
				},
				End: lspPosition{
					Line:      d.Line(),
					Character: d.End(),
				},
			},
			Severity: &sev,
			Message:  d.String(),
		})
	}

	return lds
}

func (l *lsp) handle(h jsonRpcHeader, b []byte) error {
	if l.shutdown {
		return errors.New("server is shut down")
	}

	switch h.Method {
	case "initialized":
	case "$/cancelRequest":

	case "exit":
		l.exit = true
	case "shutdown":
		l.shutdown = true

	case "textDocument/didSave":
		_, err := read[lspDidSave](b)
		if err != nil {
			return err
		}

		// TODO: handle save

	case "textDocument/didClose":
		req, err := read[lspDidCloseRequest](b)
		if err != nil {
			return err
		}

		delete(l.docs, req.Params.TextDocument.Uri)

	case "workspace/executeCommand":
		req, err := read[lspWorkspaceExecuteCommand](b)
		if err != nil {
			return err
		}

		switch req.Params.Command {
		default:
			err = l.write(jsonRpcResponse{
				JsonRpc: "2.0",
				Id:      h.Id,
				Error: &lspError{
					Code:    1,
					Message: fmt.Sprintf("unknown command: %s", req.Params.Command),
				},
			})
			if err != nil {
				return err
			}
		}

	case "textDocument/prepareRename":
		req, err := read[lspPrepareRenameRequest](b)
		if err != nil {
			return err
		}

		doc := l.docs[req.Params.TextDocument.Uri]

		res := teal.Process(doc)

		for _, sym := range res.Symbols {
			if sym.Line() == req.Params.Position.Line && req.Params.Position.Character >= sym.Begin() && req.Params.Position.Character <= sym.End() {
				err = l.write(jsonRpcResponse{
					JsonRpc: "2.0",
					Id:      h.Id,
					Result: lspPrepareRenameResponse{
						Range: lspRange{
							Start: lspPosition{
								Line:      sym.Line(),
								Character: sym.Begin(),
							},
							End: lspPosition{
								Line:      sym.Line(),
								Character: sym.Begin() + len(sym.Name()),
							},
						},
						Placeholder: sym.Name(),
					},
				})
				if err != nil {
					return err
				}
				return nil
			}
		}

		for _, ref := range res.SymbolRefs {
			if ref.Line() == req.Params.Position.Line && req.Params.Position.Character >= ref.Begin() && req.Params.Position.Character <= ref.End() {
				err = l.write(jsonRpcResponse{
					JsonRpc: "2.0",
					Id:      h.Id,
					Result: lspPrepareRenameResponse{
						Range: lspRange{
							Start: lspPosition{
								Line:      ref.Line(),
								Character: ref.Begin(),
							},
							End: lspPosition{
								Line:      ref.Line(),
								Character: ref.Begin() + len(ref.Name()),
							},
						},
						Placeholder: ref.Name(),
					},
				})
				if err != nil {
					return err
				}
				return nil
			}
		}

	case "textDocument/rename":
		req, err := read[lspRenameRequest](b)
		if err != nil {
			return err
		}

		doc := l.docs[req.Params.TextDocument.Uri]
		res := teal.Process(doc)

		chs := []lspTextEdit{}

		for _, edited := range res.Symbols {
			if edited.Line() == req.Params.Position.Line && req.Params.Position.Character >= edited.Begin() && req.Params.Position.Character <= edited.End() {
				for _, sym := range res.Symbols {
					if sym.Name() == edited.Name() {
						chs = append(chs, lspTextEdit{
							Range: lspRange{
								Start: lspPosition{
									Line:      sym.Line(),
									Character: sym.Begin(),
								},
								End: lspPosition{
									Line:      sym.Line(),
									Character: sym.Begin() + len(sym.Name()),
								},
							},
							NewText: req.Params.NewName,
						})
					}
				}
				for _, ref := range res.SymbolRefs {
					if ref.Name() == edited.Name() {
						chs = append(chs, lspTextEdit{
							Range: lspRange{
								Start: lspPosition{
									Line:      ref.Line(),
									Character: ref.Begin(),
								},
								End: lspPosition{
									Line:      ref.Line(),
									Character: ref.End(),
								},
							},
							NewText: req.Params.NewName,
						})
					}
				}
			}
		}

		for _, edited := range res.SymbolRefs {
			if edited.Line() == req.Params.Position.Line && req.Params.Position.Character >= edited.Begin() && req.Params.Position.Character <= edited.End() {
				for _, sym := range res.Symbols {
					if sym.Name() == edited.Name() {
						chs = append(chs, lspTextEdit{
							Range: lspRange{
								Start: lspPosition{
									Line:      sym.Line(),
									Character: sym.Begin(),
								},
								End: lspPosition{
									Line:      sym.Line(),
									Character: sym.Begin() + len(sym.Name()),
								},
							},
							NewText: req.Params.NewName,
						})
					}
				}
				for _, ref := range res.SymbolRefs {
					if ref.Name() == edited.Name() {
						chs = append(chs, lspTextEdit{
							Range: lspRange{
								Start: lspPosition{
									Line:      ref.Line(),
									Character: ref.Begin(),
								},
								End: lspPosition{
									Line:      ref.Line(),
									Character: ref.End(),
								},
							},
							NewText: req.Params.NewName,
						})
					}
				}
			}
		}

		err = l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result: lspWorkspaceEdit{
				Changes: map[string][]lspTextEdit{
					req.Params.TextDocument.Uri: chs,
				},
			},
		})
		if err != nil {
			return err
		}

	case "textDocument/codeAction":
		_, err := read[lspCodeActionRequest](b)
		if err != nil {
			return err
		}

		l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result:  []lspCodeAction{},
		})
	case "textDocument/diagnostic":
		req, err := read[lspDiagnosticRequest](b)
		if err != nil {
			return err
		}

		ds := l.doDiagnostic(req.Params.TextDocument.Uri)

		return l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result: &lspFullDocumentDiagnosticReport{
				Kind:  "full",
				Items: ds,
			},
		})

	case "textDocument/didOpen":
		req, err := read[lspDidOpen](b)
		if err != nil {
			return err
		}

		l.docs[req.Params.TextDocument.Uri] = req.Params.TextDocument.Text

	case "textDocument/didChange":
		req, err := read[lspDidChange](b)
		if err != nil {
			return err
		}

		for _, ch := range req.Params.ContentChanges {
			l.docs[req.Params.TextDocument.Uri] = ch.Text
		}

	case "textDocument/documentSymbol":
		req, err := read[lspDocumentSymbolRequest](b)
		if err != nil {
			return err
		}

		syms := []lspDocumentSymbol{}

		text := l.docs[req.Params.TextDocument.Uri]
		res := teal.Process(text)
		for _, s := range res.Symbols {
			r := lspRange{
				Start: lspPosition{
					Line:      s.Line(),
					Character: s.Begin(),
				},
				End: lspPosition{
					Line:      s.Line(),
					Character: s.End(),
				},
			}
			syms = append(syms, lspDocumentSymbol{
				Name:           s.Name(),
				Kind:           lspSymbolKindMethod,
				Range:          r,
				SelectionRange: r,
			})
		}

		l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result:  syms,
		})

	case "textDocument/semanticTokens/full":
		req, err := read[lspSemanticTokensFullRequest](b)
		if err != nil {
			return err
		}

		doc := l.docs[req.Params.TextDocument.Uri]
		res := teal.Process(doc)

		st := teal.SemanticTokens{}

		for i, op := range res.Listing {
			switch op.(type) {
			case *teal.PragmaExpr:
				ts := res.Lines[i]

				f := ts[0]
				l := ts[len(ts)-1]

				st = append(st, teal.SemanticToken{
					Line:      i,
					Index:     f.Begin(),
					Length:    l.End() - f.Begin(),
					Type:      semanticTokenMacro,
					Modifiers: 0,
				})
			case teal.Nop:
			case *teal.LabelExpr:
			default:
				ts := res.Lines[i]
				if len(ts) > 0 {
					f := ts[0]
					if f.Type() == teal.TokenValue {
						st = append(st, teal.SemanticToken{
							Line:      f.Line(),
							Index:     f.Begin(),
							Length:    f.End() - f.Begin(),
							Type:      semanticTokenKeyword,
							Modifiers: 0,
						})
					}
				}
			}
		}

		for _, t := range res.Tokens {
			switch t.Type() {
			case teal.TokenComment:
				st = append(st, teal.SemanticToken{
					Line:      t.Line(),
					Index:     t.Begin(),
					Length:    t.End() - t.Begin(),
					Type:      semanticTokenComment,
					Modifiers: 0,
				})
			}
		}

		for _, s := range res.Symbols {
			st = append(st, teal.SemanticToken{
				Line:      s.Line(),
				Index:     s.Begin(),
				Length:    s.End() - s.Begin(),
				Type:      semanticTokenMethod,
				Modifiers: 0,
			})
		}

		for _, s := range res.SymbolRefs {
			st = append(st, teal.SemanticToken{
				Line:      s.Line(),
				Index:     s.Begin(),
				Length:    s.End() - s.Begin(),
				Type:      semanticTokenString,
				Modifiers: 0,
			})
		}

		data := st.Encode()

		err = l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result: lspSemanticTokens{
				Data: data,
			},
		})

		if err != nil {
			return err
		}

	case "initialize":
		sync := new(int)
		*sync = 1

		symbol := new(bool)
		*symbol = true

		action := new(bool)
		*action = true

		rename := new(bool)
		*rename = true

		highlight := new(bool)
		*highlight = true

		fullSemantic := new(bool)
		*fullSemantic = true

		return l.write(jsonRpcResponse{
			JsonRpc: "2.0",
			Id:      h.Id,
			Result: &lspInitializeResult{
				Capabilities: &lspServerCapabilities{
					TextDocumentSync:       sync,
					DiagnosticProvider:     &lspDiagnosticProvider{},
					DocumentSymbolProvider: symbol,
					//CodeActionProvider:     action,
					ExecuteCommandProvider: &lspExecuteCommandProvider{
						Commands: []string{},
					},
					RenameProvider: &lspRenameOptions{
						PrepareProvider: rename,
					},
					SemanticTokensProvider: &lspSemanticTokensProvider{
						Full: fullSemantic,
						Legend: lspSemanticTokensLegend{
							TokenTypes:     []string{"keyword", "string", "comment", "method", "macro"},
							TokenModifiers: []string{},
						},
					},
				},
			},
		})
	default:
		return errors.New("unknown method")
	}

	return nil
}

func (l *lsp) write(v interface{}) error {
	rb, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "failed to marshal response")
	}

	l.trace(fmt.Sprintf("OUT: %s", string(rb)))

	h := http.Header{}
	h.Set("Content-Length", strconv.Itoa(len(rb)))

	err = h.Write(l.w)
	if err != nil {
		return errors.Wrap(err, "failed to write response headers")
	}

	_, err = l.w.Write([]byte("\r\n"))
	if err != nil {
		return errors.Wrap(err, "failed to write")
	}

	_, err = l.w.Write(rb)
	if err != nil {
		return errors.Wrap(err, "failed to write response body")
	}

	err = l.w.Flush()
	if err != nil {
		return errors.Wrap(err, "failed to flush")
	}

	return nil
}

func (l *lsp) trace(s string) {
	if l.debug == nil {
		return
	}

	l.debug.WriteString(s)
	l.debug.WriteString("\n")

	l.debug.Flush()
}

func (l *lsp) Run() error {
	l.trace("TEAL LSP running..")
	defer func() {
		l.trace("TEAL LSP exited.")
	}()

	for !l.exit {
		err := func() error {
			mh, err := l.tp.ReadMIMEHeader()
			if err != nil {
				return errors.Wrap(err, "failed to read request headers")
			}

			h := http.Header(mh)

			length, err := strconv.Atoi(h.Get("Content-Length"))
			if err != nil {
				return errors.Wrap(err, "failed to parse content length")
			}

			data := make([]byte, length)
			_, err = io.ReadFull(l.tp.R, data)
			if err != nil {
				return errors.Wrap(err, "failed to read content body")
			}

			l.trace(fmt.Sprintf("IN: %s", string(data)))

			var jh jsonRpcHeader
			err = json.Unmarshal(data, &jh)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal json rpc header")
			}

			err = l.handle(jh, data)
			if err != nil {
				return errors.Wrap(err, "failed to handle request")
			}

			return nil
		}()

		if err != nil {
			l.trace(fmt.Sprintf("ERR: %s", err))
		}

	}

	return nil
}
