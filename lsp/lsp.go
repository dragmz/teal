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

type lspDoc struct {
	s   string
	res *teal.ProcessResult
}

func (d *lspDoc) Update(s string) {
	d.s = s
	d.res = nil
}

func (d *lspDoc) Results() *teal.ProcessResult {
	if d.res == nil {
		d.res = teal.Process(d.s)
	}

	return d.res
}

type lsp struct {
	docs     map[string]*lspDoc
	shutdown bool

	exit     bool
	exitCode int

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
		docs: map[string]*lspDoc{},
	}

	for _, opt := range opts {
		err := opt(l)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set lsp option")
		}
	}

	return l, nil
}

type jsonRpcRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
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

type tealCreateLabelCommandArgs struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type lspWorkspaceExecuteCommandHeader struct {
	Command string `json:"command"`
}

type lspWorkspaceExecuteCommandBodyArguments[T any] struct {
	Arguments T `json:"arguments"`
}

type lspWorkspaceExecuteCommandBody[T any] struct {
	Params lspWorkspaceExecuteCommandBodyArguments[T] `json:"params"`
}

type lspDiagnosticRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspDiagnosticRequestParams struct {
	TextDocument *lspDiagnosticRequestTextDocument `json:"textDocument"`
}

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

type lspRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspRenameRequestParams struct {
	TextDocument lspRenameRequestTextDocument `json:"textDocument"`
	Position     lspPosition
	NewName      string `json:"newName"`
}

type lspPrepareRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspPrepareRenameRequestParams struct {
	TextDocument lspPrepareRenameRequestTextDocument `json:"textDocument"`
	Position     lspPosition
}

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

type lspOptionalVersionedTextDocumentIdentifier struct {
	Uri     string `json:"uri"`
	Version *int   `json:"version"`
}

type lspTextDocumentEdit struct {
	TextDocument lspOptionalVersionedTextDocumentIdentifier `json:"textDocument"`
	Edits        []lspTextEdit                              `json:"edits"`
}

type lspWorkspaceEdit struct {
	Changes         map[string][]lspTextEdit `json:"changes,omitempty"`
	DocumentChanges []lspTextDocumentEdit    `json:"documentChanges,omitempty"`
}

type lspWorkspaceApplyEditRequestParams struct {
	Label string           `json:"label,omitempty"`
	Edit  lspWorkspaceEdit `json:"edit"`
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
	Uri string `json:"uri"`
}

type lspDocumentHighlightRequestParams struct {
	TextDocument lspDocumentHighlightRequestTextDocument `json:"textDocument"`
	Position     lspPosition                             `json:"position"`
}

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

// notifications
type lspDidChange lspRequest[*lspDidChangeParams]
type lspDidOpen lspRequest[*lspDidOpenParams]
type lspDidSave lspRequest[*lspDidSaveParams]

// requests
type lspDocumentSymbolRequest lspRequest[*lspDocumentSymbolParams]
type lspWorkspaceExecuteCommand lspRequest[*lspWorkspaceExecuteCommandHeader]
type lspDiagnosticRequest lspRequest[*lspDiagnosticRequestParams]
type lspCodeActionRequest lspRequest[*lspCodeActionRequestParams]
type lspRenameRequest lspRequest[*lspRenameRequestParams]
type lspPrepareRenameRequest lspRequest[*lspPrepareRenameRequestParams]
type lspDocumentColorRequest lspRequest[*lspDocumentColorRequestParams]
type lspDidCloseRequest lspRequest[*lspDidCloseRequestParams]
type lspDocumentHighlightRequest lspRequest[*lspDocumentHighlightRequestParams]
type lspSemanticTokensFullRequest lspRequest[*lspSemanticTokensFullRequestParams]

func readInto(b []byte, v interface{}) error {
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}

func (l *lsp) reply(id interface{}, result interface{}, err interface{}) error {
	return l.write(jsonRpcResponse{
		JsonRpc: "2.0",
		Id:      id,
		Result:  result,
		Error:   err,
	})
}

func (l *lsp) fail(id interface{}, err interface{}) error {
	return l.reply(id, nil, err)
}

func (l *lsp) success(id interface{}, result interface{}) error {
	return l.reply(id, result, nil)
}

func read[T any](b []byte) (T, error) {
	var v T

	err := readInto(b, &v)
	if err != nil {
		return v, err
	}

	return v, nil
}

func (l *lsp) request(id string, method string, params interface{}) error {
	return l.write(jsonRpcRequest{
		JsonRpc: "2.0",
		Id:      id,
		Method:  method,
		Params:  params,
	})
}

func (l *lsp) notify(method string, params interface{}) error {
	return l.write(lspNotification{
		JsonRpc: "2.0",
		Method:  method,
		Params:  params,
	})
}

func (l *lsp) notifyDiagnostics(uri string, lds []lspDiagnostic) error {
	return l.notify("textDocument/publishDiagnostics", lspPublishDiagnostic{
		Uri:         uri,
		Diagnostics: lds,
	})
}

func (l *lsp) doDiagnostic(doc *lspDoc) []lspDiagnostic {
	res := doc.Results()

	lds := []lspDiagnostic{}
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
	switch h.Method { // notifications
	case "initialized":

	case "exit":
		l.exit = true
		if !l.shutdown {
			l.exitCode = 1
		}

	case "textDocument/didOpen":
		req, err := read[lspDidOpen](b)
		if err != nil {
			return err
		}

		doc := l.docs[req.Params.TextDocument.Uri]
		if doc == nil {
			doc = &lspDoc{}
			l.docs[req.Params.TextDocument.Uri] = doc
		}

		doc.Update(req.Params.TextDocument.Text)

	case "textDocument/didChange":
		req, err := read[lspDidChange](b)
		if err != nil {
			return err
		}

		for _, ch := range req.Params.ContentChanges {
			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			doc.Update(ch.Text)
		}

	case "textDocument/didSave":
		_, err := read[lspDidSave](b)
		if err != nil {
			return err
		}

		// TODO: handle save

	default: // requests

		if l.shutdown {
			return errors.New("cannot process requests - server is shut down")
		}

		switch h.Method {
		case "shutdown":
			l.shutdown = true
			return l.success(h.Id, nil)

		case "$/cancelRequest":

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
			case "teal.label.create":
				var body lspWorkspaceExecuteCommandBody[[]tealCreateLabelCommandArgs]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				args := body.Params.Arguments
				if len(args) != 1 {
					return errors.New("unexpected number of args")
				}

				doc := l.docs[args[0].Uri]
				if doc == nil {
					return errors.New("doc not found")
				}

				res := doc.Results()

				name := args[0].Name
				s := fmt.Sprintf("\r\n%s:\r\n", name)

				return l.request("1", "workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: fmt.Sprintf("Create label: %s", name),
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: args[0].Uri,
								},
								Edits: []lspTextEdit{
									{
										Range: lspRange{
											Start: lspPosition{
												Line:      len(res.Lines),
												Character: 0,
											},
											End: lspPosition{
												Line:      len(res.Lines),
												Character: len(s),
											},
										},
										NewText: s,
									},
								},
							},
						},
					},
				})

			default:
				return l.fail(h.Id, lspError{
					Code:    1,
					Message: fmt.Sprintf("unknown command: %s", req.Params.Command),
				})
			}

		case "textDocument/prepareRename":
			req, err := read[lspPrepareRenameRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			res := doc.Results()

			for _, sym := range res.Symbols {
				if sym.Line() == req.Params.Position.Line && req.Params.Position.Character >= sym.Begin() && req.Params.Position.Character <= sym.End() {
					return l.success(h.Id, lspPrepareRenameResponse{
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
					})
				}
			}

			for _, ref := range res.SymbolRefs {
				if ref.Line() == req.Params.Position.Line && req.Params.Position.Character >= ref.Begin() && req.Params.Position.Character <= ref.End() {
					return l.success(h.Id, lspPrepareRenameResponse{
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
					})
				}
			}

		case "textDocument/rename":
			req, err := read[lspRenameRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			res := doc.Results()

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

			return l.success(h.Id, lspWorkspaceEdit{
				Changes: map[string][]lspTextEdit{
					req.Params.TextDocument.Uri: chs,
				},
			})

		case "textDocument/codeAction":
			req, err := read[lspCodeActionRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			res := doc.Results()

			cas := []lspCodeAction{}
			for _, ref := range res.SymbolRefs {
				if ref.Line() < req.Params.Range.Start.Line || ref.Line() > req.Params.Range.End.Line {
					continue
				}

				if ref.Line() == req.Params.Range.Start.Line {
					if ref.Begin() < req.Params.Range.Start.Character {
						continue
					}
				} else if ref.Line() == req.Params.Range.End.Line {
					if ref.End() > req.Params.Range.End.Character {
						continue
					}
				}

				found := func() bool {
					for _, sym := range res.Symbols {
						if sym.Name() == ref.Name() {
							return true
						}
					}

					return false
				}()

				if found {
					continue
				}

				kind := "quickfix"
				cas = append(cas, lspCodeAction{
					Title: fmt.Sprintf("Create label '%s'", ref.Name()),
					Kind:  &kind,
					Command: &lspCommand{
						Title:   "Create label",
						Command: "teal.label.create",
						Arguments: []interface{}{
							tealCreateLabelCommandArgs{
								Uri:  req.Params.TextDocument.Uri,
								Name: ref.Name(),
							},
						},
					},
				})
			}

			return l.success(h.Id, cas)
		case "textDocument/diagnostic":
			req, err := read[lspDiagnosticRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			ds := l.doDiagnostic(doc)

			return l.success(h.Id, lspFullDocumentDiagnosticReport{
				Kind:  "full",
				Items: ds,
			})

		case "textDocument/documentSymbol":
			req, err := read[lspDocumentSymbolRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			res := doc.Results()

			syms := []lspDocumentSymbol{}
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

			return l.success(h.Id, syms)

		case "textDocument/semanticTokens/full":
			req, err := read[lspSemanticTokensFullRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			if doc == nil {
				return errors.New("doc not found")
			}

			res := doc.Results()

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

			return l.success(h.Id, lspSemanticTokens{
				Data: data,
			})

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

			return l.success(h.Id, lspInitializeResult{
				Capabilities: &lspServerCapabilities{
					TextDocumentSync:       sync,
					DiagnosticProvider:     &lspDiagnosticProvider{},
					DocumentSymbolProvider: symbol,
					CodeActionProvider:     action,
					ExecuteCommandProvider: &lspExecuteCommandProvider{
						Commands: []string{
							"teal.label.create",
						},
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
			})
		default:
			return errors.New("unknown method")
		}
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

func (l *lsp) Run() (int, error) {
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

	return l.exitCode, nil
}