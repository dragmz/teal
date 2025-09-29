package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type tealCompletionMode int

const (
	tealCompletionNone = tealCompletionMode(iota)
	tealCompletionOp
	tealCompletionArg
)

const (
	semanticTokenKeyword  = 0
	semanticTokenString   = 1
	semanticTokenComment  = 2
	semanticTokenMethod   = 3
	semanticTokenMacro    = 4
	semanticTokenValue    = 5
	semanticTokenNumber   = 6
	semanticTokenOperator = 7
	semanticTokenFunction = 8
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
	id int

	config tealConfig

	docs     map[string]*lspDoc
	shutdown bool

	exit     bool
	exitCode int

	tp *textproto.Reader
	w  *bufio.Writer

	debug *bufio.Writer

	prepareDiagnostics PrepareDiagnosticsHandler
	prepareCodeLens    PrepareCodeLensHandler
	prepareInlayHints  PrepareInlayHintsHandler
	prepareSymbols     PrepareSymbolsHandler
	prepareOffsets     PrepareOffsetsHandler

	opDocShort OpDocHandler
	opDocExtra OpDocHandler
}

type LspOption func(l *lsp) error

func WithDebug(w io.Writer) LspOption {
	return func(l *lsp) error {
		l.debug = bufio.NewWriter(w)
		return nil
	}
}

type OpDocHandler func(op string) string

func WithOpDocShortHandler(h OpDocHandler) LspOption {
	return func(l *lsp) error {
		l.opDocShort = h
		return nil
	}
}

func WithOpDocExtraHandler(h OpDocHandler) LspOption {
	return func(l *lsp) error {
		l.opDocExtra = h
		return nil
	}
}

type PrepareDiagnosticsHandler func(source string) []LspDiagnostic

func WithPrepareDiagnosticsHandler(h PrepareDiagnosticsHandler) LspOption {
	return func(l *lsp) error {
		l.prepareDiagnostics = h
		return nil
	}
}

type PrepareCodeLensHandler func(source string) []LspCodeLens

func WithPrepareCodeLensHandler(h PrepareCodeLensHandler) LspOption {
	return func(l *lsp) error {
		l.prepareCodeLens = h
		return nil
	}
}

type PrepareInlayHintsHandler func(source string) []LspInlayHint

func WithPrepareInlayHintsHandler(h PrepareInlayHintsHandler) LspOption {
	return func(l *lsp) error {
		l.prepareInlayHints = h
		return nil
	}
}

type PrepareSymbolsHandler func(source string) []LspDocumentSymbol

func WithPrepareSymbolsHandler(h PrepareSymbolsHandler) LspOption {
	return func(l *lsp) error {
		l.prepareSymbols = h
		return nil
	}
}

type SourceLocation struct {
	Line   int
	Column int
}

type PrepareOffsetsHandler func(source string) map[int]SourceLocation

func WithPrepareOffsetsHandler(h PrepareOffsetsHandler) LspOption {
	return func(l *lsp) error {
		l.prepareOffsets = h
		return nil
	}
}

func New(r io.Reader, w io.Writer, opts ...LspOption) (*lsp, error) {
	l := &lsp{
		tp:   textproto.NewReader(bufio.NewReader(r)),
		w:    bufio.NewWriter(w),
		docs: map[string]*lspDoc{},
		config: tealConfig{
			SemanticTokens: true,
			InlayNamed:     true,
			InlayDecoded:   true,
			LensRefs:       true,
		},
	}

	for _, opt := range opts {
		err := opt(l)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set lsp option")
		}
	}

	if l.opDocShort == nil {
		l.opDocShort = func(op string) string {
			return teal.OpDocByName(op)
		}
	}

	if l.opDocExtra == nil {
		l.opDocExtra = func(op string) string {
			return teal.OpDocExtraByName(op)
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
	JsonRpc string `json:"jsonrpc"`

	Id interface{} `json:"id"`

	Method string `json:"method"`

	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type jsonRpcResponse struct {
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Id      interface{} `json:"id"`
}

type lspFullDocumentDiagnosticReport struct {
	Kind  string          `json:"kind"`
	Items []LspDiagnostic `json:"items"`
}

type lspRequest[T any] struct {
	Params T `json:"params"`
}

type lspDiagnosticProvider struct {
	InterFileDependencies bool `json:"interFileDependencies"`
	WorkspaceDiagnostics  bool `json:"workspaceDiagnostics"`
}

type lspCompletionProviderItem struct {
	LabelDetailsSupport *bool `json:"labelDetailsSupport,omitempty"`
}

type lspCompletionItemLabelDetails struct {
	Detail      string `json:"detail,omitempty"`
	Description string `json:"description,omitempty"`
}

type lspCompletionItem struct {
	Label        string                         `json:"label"`
	LabelDetails *lspCompletionItemLabelDetails `json:"labelDetails,omitempty"`
	Kind         *int                           `json:"kind,omitempty"`
	Detail       string                         `json:"detail,omitempty"`

	// string | MarkupContent
	Documentation interface{} `json:"documentation,omitempty"`

	Deprecated          *bool         `json:"deprecated,omitempty"`
	Preselect           *bool         `json:"preselect,omitempty"`
	SortText            string        `json:"sortText,omitempty"`
	FilterText          string        `json:"filterText,omitempty"`
	InsertText          string        `json:"insertText,omitempty"`
	InsertTextFormat    *int          `json:"insertTextFormat,omitempty"`
	InsertTextMode      *int          `json:"insertTextMode,omitempty"`
	TextEdit            *lspTextEdit  `json:"textEdit,omitempty"`
	TextEditText        string        `json:"textEditText,omitempty"`
	AdditionalTextEdits []lspTextEdit `json:"additionalTextEdits,omitempty"`
	CommitCharacters    []string      `json:"commitCharacters,omitempty"`
	Command             *LspCommand   `json:"command,omitempty"`
	Data                interface{}   `json:"data,omitempty"`
}

type lspCompletionList struct {
	IsIncomplete bool                `json:"isIncomplete"`
	Items        []lspCompletionItem `json:"items"`
}

type lspCompletionProvider struct {
	TriggerCharacters   []string                   `json:"triggerCharacters,omitempty"`
	AllCommitCharacters []string                   `json:"allCommitCharacters,omitempty"`
	ResolveProvider     *bool                      `json:"resolveProvider,omitempty"`
	CompletionItem      *lspCompletionProviderItem `json:"completionItem,omitempty"`
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

type lspCodeLensProvider struct {
	ResolveProvider *bool `json:"resolveProvider,omitempty"`
}

type lspSemanticTokens struct {
	Data []uint32 `json:"data"`
}

type lspSignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters,omitempty"`
	RetriggerCharacters []string `json:"retriggerCharacters,omitempty"`
}

type lspParameterInformation struct {
	Label string `json:"label"`

	// string | [uinteger, uinteger]
	Documentation interface{} `json:"documentation,omitempty"`
}

type lspSignatureInformation struct {
	Label string `json:"label"`

	//string | Markupcontent
	Documentation interface{} `json:"documentation,omitempty"`

	Parameters      []lspParameterInformation `json:"parameters,omitempty"`
	ActiveParameter *int                      `json:"activeParameter,omitempty"`
}

type lspSignatureHelp struct {
	Signatures []lspSignatureInformation `json:"signatures,omitempty"`
}

type lspServerCapabilities struct {
	TextDocumentSync           *int                       `json:"textDocumentSync,omitempty"`
	DiagnosticProvider         *lspDiagnosticProvider     `json:"diagnosticProvider,omitempty"`
	CompletionProvider         *lspCompletionProvider     `json:"completionProvider,omitempty"`
	DocumentSymbolProvider     *bool                      `json:"documentSymbolProvider,omitempty"`
	CodeActionProvider         *bool                      `json:"codeActionProvider,omitempty"`
	ExecuteCommandProvider     *lspExecuteCommandProvider `json:"executeCommandProvider,omitempty"`
	RenameProvider             *lspRenameOptions          `json:"renameProvider,omitempty"`
	ColorProvider              *bool                      `json:"colorProvider,omitempty"`
	DocumentHighlightProvider  *bool                      `json:"documentHighlightProvider,omitempty"`
	SemanticTokensProvider     *lspSemanticTokensProvider `json:"semanticTokensProvider,omitempty"`
	DocumentFormattingProvider *bool                      `json:"documentFormattingProvider,omitempty"`
	DefinitionProvider         *bool                      `json:"definitionProvider,omitempty"`
	HoverProvider              *bool                      `json:"hoverProvider,omitempty"`
	SignatureHelpProvider      *lspSignatureHelpOptions   `json:"signatureHelpProvider,omitempty"`
	InlayHintProvider          *bool                      `json:"inlayHintProvider,omitempty"`
	InlineValueProvider        *bool                      `json:"inlineValueProvider,omitempty"`
	CodeLensProvider           *lspCodeLensProvider       `json:"codeLensProvider,omitempty"`
}

type lspInitializeResult struct {
	Capabilities *lspServerCapabilities `json:"capabilities"`
}

type LspSymbolKind int

const (
	LspSymbolKindFile          = 1
	LspSymbolKindModule        = 2
	LspSymbolKindNamespace     = 3
	LspSymbolKindPackage       = 4
	LspSymbolKindClass         = 5
	LspSymbolKindMethod        = 6
	LspSymbolKindProperty      = 7
	LspSymbolKindField         = 8
	LspSymbolKindConstructor   = 9
	LspSymbolKindEnum          = 10
	LspSymbolKindInterface     = 11
	LspSymbolKindFunction      = 12
	LspSymbolKindVariable      = 13
	LspSymbolKindConstant      = 14
	LspSymbolKindString        = 15
	LspSymbolKindNumber        = 16
	LspSymbolKindBoolean       = 17
	LspSymbolKindArray         = 18
	LspSymbolKindObject        = 19
	LspSymbolKindKey           = 20
	LspSymbolKindNull          = 21
	LspSymbolKindEnumMember    = 22
	LspSymbolKindStruct        = 23
	LspSymbolKindEvent         = 24
	LspSymbolKindOperator      = 25
	LspSymbolKindTypeParameter = 26
)

type LspDocumentSymbol struct {
	Name           string        `json:"name"`
	Kind           LspSymbolKind `json:"kind"`
	Range          LspRange      `json:"range"`
	SelectionRange LspRange      `json:"selectionRange"`
}

type lspInitializeClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type tealInitializationOptions struct {
	SemanticTokens *bool `json:"semanticTokens,omitempty"`
	InlayNamed     *bool `json:"inlayNamed,omitempty"`
	InlayDecoded   *bool `json:"inlayDecoded,omitempty"`
	LensRefs       *bool `json:"lensRefs,omitempty"`
}

type tealConfig struct {
	SemanticTokens bool
	InlayNamed     bool
	InlayDecoded   bool
	LensRefs       bool
}

type lspCompletionCompletionItemClientCapabilities struct {
	SnippetSupport          *bool `json:"snippetSupport,omitempty"`
	CommitCharactersSupport *bool `json:"commitCharactersSupport,omitempty"`
}

type lspCompletionTagSupportClientCapabilities struct {
	ValueSet []int `json:"valueSet,omitempty"`
}

type lspCompletionInsertTextModeSupportClientCapabilities struct {
	ValueSet []int `json:"valueSet,omitempty"`
}

type lspCompletionClientCapabilities struct {
	DynamicRegistration   *bool                                                 `json:"dynamicRegistration,omitempty"`
	CompletionItem        *lspCompletionCompletionItemClientCapabilities        `json:"completionItem,omitempty"`
	DocumentationFormat   []string                                              `json:"documentationFormat,omitempty"`
	DeprecatedSupport     *bool                                                 `json:"deprecatedSupport,omitempty"`
	PreselectSupport      *bool                                                 `json:"preselectSupport,omitempty"`
	TagSupport            *lspCompletionTagSupportClientCapabilities            `json:"tagSupport,omitempty"`
	InsertReplaceSupport  *bool                                                 `json:"insertReplaceSupport,omitempty"`
	ResolveSupport        []string                                              `json:"resolveSupport,omitempty"`
	InsertTextModeSupport *lspCompletionInsertTextModeSupportClientCapabilities `json:"insertTextModeSupport,omitempty"`
	LabelDetailsSupport   *bool                                                 `json:"labelDetailsSupport,omitempty"`
}

type lspTextDocumentCompletionListClientCapabilities struct {
	ItemDefaults []string `json:"itemDefaults,omitempty"`
}

type lspTextDocumentClientCapabilities struct {
	Completion         *lspCompletionClientCapabilities                 `json:"completion"`
	CompletionItemKind []int                                            `json:"completionItemKind,omitempty"`
	ContextSupport     *bool                                            `json:"contextSupport,omitempty"`
	InsertTextMode     *int                                             `json:"insertTextMode,omitempty"`
	CompletionList     *lspTextDocumentCompletionListClientCapabilities `json:"completionList,omitempty"`
}

type lspClientCapabilities struct {
	TextDocument *lspTextDocumentClientCapabilities `json:"textDocument,omitempty"`
}

type lspInitializeRequestParams struct {
	ProcessId             int                        `json:"id"`
	ClientInfo            *lspInitializeClientInfo   `json:"clientInfo"`
	InitializationOptions *tealInitializationOptions `json:"initializationOptions,omitempty"`
	Capabilities          lspClientCapabilities      `json:"capabilities"`
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

type tealGotoPcCommandArgs struct {
	Uri string `json:"uri"`
	Pc  int    `json:"pc"`
}

type tealUpdateVersion struct {
	Uri     string `json:"uri"`
	Version uint64 `json:"version"`
}

type tealReplaceValueCommandArgs struct {
	Uri   string   `json:"uri"`
	Range LspRange `json:"range"`
	Value string   `json:"name"`
}

type tealCreateLabelCommandArgs struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type tealRemoveLabelCommandArgs struct {
	Uri  string `json:"uri"`
	Name string `json:"name"`
}

type tealRemoveLineCommandArgs struct {
	Uri  string `json:"uri"`
	Line int    `json:"line"`
}

type tealRemoveCallCommandArgs struct {
	Uri     string `json:"uri"`
	Line    int    `json:"line"`
	Subline int    `json:"subline"`
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
	Diagnosticts []LspDiagnostic `json:"diagnostics"`
	Only         []string        `json:"only,omitempty"`
	TriggerKind  *int            `json:"triggerKind,omitempty"`
}

type lspCodeActionRequestParams struct {
	TextDocument lspCodeActionTextDocument `json:"textDocument"`
	Range        LspRange                  `json:"range"`
	Context      lspCodeActionContext      `json:"context"`
}

type lspRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspRenameRequestParams struct {
	TextDocument  lspRenameRequestTextDocument `json:"textDocument"`
	Position      LspPosition
	NewName       string      `json:"newName"`
	WorkDoneToken interface{} `json:"workDoneToken,omitempty"`
}

type lspPrepareRenameRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspPrepareRenameRequestParams struct {
	TextDocument  lspPrepareRenameRequestTextDocument `json:"textDocument"`
	Position      LspPosition
	WorkDoneToken interface{} `json:"workDoneToken,omitempty"`
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
	Range LspRange `json:"range"`
	Color lspColor `json:"color"`
}

type lspDocumentColorRequestParams struct {
	TextDocument lspDocumentColorRequestTextDocument `json:"textDocument"`
}

type lspPrepareRenameResponse struct {
	Range       LspRange `json:"range"`
	Placeholder string   `json:"placeholder"`
}

type LspCommand struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

type lspTextEdit struct {
	Range   LspRange `json:"range"`
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

type lspShowMessageParams struct {
	Type    int    `json:"type"`
	Message string `json:"message"`
}

type lspCodeAction struct {
	Title       string            `json:"title"`
	Kind        *string           `json:"kind,omitempty"`
	Diagnostics []LspDiagnostic   `json:"diagnostics,omitempty"`
	IsPreferred *bool             `json:"isPreferred,omitempty"`
	Edit        *lspWorkspaceEdit `json:"edit,omitempty"`
	Command     *LspCommand       `json:"command,omitempty"`
}

type lspDidCloseTextDocument struct {
	Uri string `json:"uri"`
}

type lspDidCloseRequestParams struct {
	TextDocument *lspDidCloseTextDocument `json:"textDocument"`
}

type LspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

func (p LspPosition) StartLine() int {
	return p.Line
}

func (p LspPosition) StartCharacter() int {
	return p.Character
}

func (p LspPosition) EndLine() int {
	return p.Line
}

func (p LspPosition) EndCharacter() int {
	return p.Character
}

type LspRange struct {
	Start LspPosition `json:"start"`
	End   LspPosition `json:"end"`
}

func (r LspRange) StartLine() int {
	return r.Start.Line
}

func (r LspRange) StartCharacter() int {
	return r.Start.Character
}

func (r LspRange) EndLine() int {
	return r.End.Line
}

func (r LspRange) EndCharacter() int {
	return r.End.Character
}

type LspDiagnostic struct {
	Range    LspRange `json:"range"`
	Severity *int     `json:"severity,omitempty"`
	Message  string   `json:"message"`
}

type lspPublishDiagnostic struct {
	Uri         string          `json:"uri"`
	Diagnostics []LspDiagnostic `json:"diagnostics"`
}

type lspNotification struct {
	JsonRpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type lspProgressParams struct {
	Token interface{} `json:"token"`
	Value interface{} `json:"value"`
}

type lspWorkDoneProgressBegin struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Cancellable *bool  `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  *int   `json:"percentage,omitempty"`
}

type lspWorkDoneProgressReport struct {
	Kind        string `json:"kind"`
	Cancellable *bool  `json:"cancellable,omitempty"`
	Message     string `json:"message,omitempty"`
	Percentage  *int   `json:"percentage,omitempty"`
}

type lspWorkDoneProgressEnd struct {
	Kind    string `json:"kind"`
	Message string `json:"message,omitempty"`
}

type lspRenameOptions struct {
	PrepareProvider  *bool `json:"prepareProvider,omitempty"`
	WorkDoneProgress *bool `json:"workDoneProgress,omitempty"`
}

type lspDocumentHighlightRequestTextDocument struct {
	Uri string `json:"uri"`
}

type lspDocumentHighlightRequestParams struct {
	TextDocument lspDocumentHighlightRequestTextDocument `json:"textDocument"`
	Position     LspPosition                             `json:"position"`
}

type lspDocumentHighlight struct {
	Range LspRange `json:"range"`
	Kind  *int     `json:"kind"`
}

type lspTextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type lspSemanticTokensFullRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

type lspCompletionRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Position     LspPosition               `json:"position"`
}

type lspDocumentFormattingRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

type lspDefinitionRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Position     LspPosition               `json:"position"`
}

type lspLocation struct {
	Uri   string   `json:"uri"`
	Range LspRange `json:"range"`
}

type lspMarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type lspHover struct {
	Contents lspMarkupContent `json:"contents"`
	Range    LspRange         `json:"range,omitempty"`
}

type lspHoverRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Position     LspPosition               `json:"position"`
}

type lspSignatureHelpRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Position     LspPosition               `json:"position"`
}

type lspInlayHintRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Range        LspRange                  `json:"range,omitempty"`
}

type LspInlayHint struct {
	Position    LspPosition `json:"position"`
	Label       string      `json:"label"`
	Kind        *int        `json:"kind,omitempty"`
	Tooltip     string      `json:"tooltip,omitempty"`
	PaddingLeft *bool       `json:"paddingLeft,omitempty"`
}

type lspInlineValueRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
	Range        LspRange                  `json:"range,omitempty"`
}

type lspInlineValueText struct {
	Range LspRange `json:"range"`
	Text  string   `json:"text"`
}

type lspCodeLensRequestParams struct {
	TextDocument lspTextDocumentIdentifier `json:"textDocument"`
}

type LspCodeLens struct {
	Range   LspRange    `json:"range"`
	Command *LspCommand `json:"command,omitempty"`
	Data    any         `json:"data,omitempty"`
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
type lspCompletionRequest lspRequest[*lspCompletionRequestParams]
type lspDocumentFormattingRequest lspRequest[*lspDocumentFormattingRequestParams]
type lspDefinitionRequest lspRequest[*lspDefinitionRequestParams]
type lspHoverRequest lspRequest[*lspHoverRequestParams]
type lspSignatureHelpRequest lspRequest[*lspSignatureHelpRequestParams]
type lspInlayHintRequest lspRequest[*lspInlayHintRequestParams]
type lspInlineValueRequest lspRequest[*lspInlineValueRequestParams]
type lspInitializeRequest lspRequest[*lspInitializeRequestParams]
type lspCodeLensRequest lspRequest[*lspCodeLensRequestParams]

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

func (l *lsp) request(method string, params interface{}) error {
	l.id++
	return l.write(jsonRpcRequest{
		JsonRpc: "2.0",
		Id:      strconv.Itoa(l.id),
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

func (l *lsp) notifyDiagnostics(uri string, lds []LspDiagnostic) error {
	return l.notify("textDocument/publishDiagnostics", lspPublishDiagnostic{
		Uri:         uri,
		Diagnostics: lds,
	})
}

func (l *lsp) notifyProgress(token interface{}, value interface{}) error {
	if token == nil {
		return nil
	}
	return l.notify("$/progress", lspProgressParams{
		Token: token,
		Value: value,
	})
}

func (l *lsp) reportProgressBegin(token interface{}, title string, message string) error {
	return l.notifyProgress(token, lspWorkDoneProgressBegin{
		Kind:    "begin",
		Title:   title,
		Message: message,
	})
}

func (l *lsp) reportProgressEnd(token interface{}, message string) error {
	return l.notifyProgress(token, lspWorkDoneProgressEnd{
		Kind:    "end",
		Message: message,
	})
}

func (l *lsp) prepare(uri string) (*lspDoc, *teal.ProcessResult, error) {
	doc := l.docs[uri]
	if doc == nil {
		return nil, nil, errors.New("doc not found")
	}

	return doc, doc.Results(), nil
}

func (l *lsp) handle(h jsonRpcHeader, b []byte) error {

	if h.Result != nil {
		// TODO: handle success
		return nil
	}

	if h.Error != nil {
		// TODO: handle failure
		return nil
	}

	if h.Method == "" {
		// TODO: handle response
		return nil
	}

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
			return l.success(h.Id, []struct{}{})

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
			case "teal.pc.resolve":
				var body lspWorkspaceExecuteCommandBody[tealGotoPcCommandArgs]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				if l.prepareOffsets == nil {
					return l.fail(h.Id, lspError{
						Code:    2,
						Message: "prepare offsets handler is not set",
					})
				}

				doc := l.docs[body.Params.Arguments.Uri]
				if doc == nil {
					return errors.New("doc not found")
				}

				offsets := l.prepareOffsets(doc.s)

				loc, ok := offsets[body.Params.Arguments.Pc]
				if !ok {
					return l.fail(h.Id, lspError{
						Code:    3,
						Message: "pc not found",
					})
				}

				return l.success(h.Id, LspPosition{
					Line: loc.Line, Character: loc.Column,
				})

			case "teal.version.update":
				var body lspWorkspaceExecuteCommandBody[[]tealUpdateVersion]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				args := body.Params.Arguments
				if len(args) != 1 {
					return errors.New("unexpected number of args")
				}

				arg := args[0]

				doc := l.docs[arg.Uri]
				if doc == nil {
					return errors.New("doc not found")
				}

				res := doc.Results()

				edits := []lspTextEdit{
					prepareVersionEdit(res.VersionToken, arg.Version),
				}

				return l.request("workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: "Update version",
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: arg.Uri,
								},
								Edits: edits,
							},
						},
					},
				})

			case "teal.value.replace":
				var body lspWorkspaceExecuteCommandBody[[]tealReplaceValueCommandArgs]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				args := body.Params.Arguments
				if len(args) != 1 {
					return errors.New("unexpected number of args")
				}

				arg := args[0]

				doc := l.docs[arg.Uri]
				if doc == nil {
					return errors.New("doc not found")
				}

				edits := []lspTextEdit{
					{
						Range:   arg.Range,
						NewText: arg.Value,
					},
				}

				return l.request("workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: "Replace with named value",
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: arg.Uri,
								},
								Edits: edits,
							},
						},
					},
				})
			case "teal.call.remove":
				var body lspWorkspaceExecuteCommandBody[[]tealRemoveCallCommandArgs]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				args := body.Params.Arguments
				if len(args) != 1 {
					return errors.New("unexpected number of args")
				}

				arg := args[0]

				doc := l.docs[arg.Uri]
				if doc == nil {
					return errors.New("doc not found")
				}

				res := doc.Results()

				line := arg.Line
				subline := arg.Subline

				edits := []lspTextEdit{prepareRemoveSublineEdit(res, line, subline)}

				return l.request("workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: "Remove call",
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: arg.Uri,
								},
								Edits: edits,
							},
						},
					},
				})
			case "teal.label.remove":
				var body lspWorkspaceExecuteCommandBody[[]tealRemoveLabelCommandArgs]
				err := readInto(b, &body)
				if err != nil {
					return err
				}

				args := body.Params.Arguments
				if len(args) != 1 {
					return errors.New("unexpected number of args")
				}

				arg := args[0]

				_, res, err := l.prepare(arg.Uri)
				if err != nil {
					return err
				}

				name := arg.Name

				edits := []lspTextEdit{}

				for _, sym := range res.SymByName(body.Params.Arguments[0].Name) {
					edits = append(edits, prepareRemoveSymbolEdit(sym))
				}

				return l.request("workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: fmt.Sprintf("Remove label: %s", name),
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: arg.Uri,
								},
								Edits: edits,
							},
						},
					},
				})

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

				arg := args[0]

				_, res, err := l.prepare(arg.Uri)
				if err != nil {
					return err
				}

				name := arg.Name

				return l.request("workspace/applyEdit", lspWorkspaceApplyEditRequestParams{
					Label: fmt.Sprintf("Create label: %s", name),
					Edit: lspWorkspaceEdit{
						DocumentChanges: []lspTextDocumentEdit{
							{
								TextDocument: lspOptionalVersionedTextDocumentIdentifier{
									Uri: arg.Uri,
								},
								Edits: []lspTextEdit{
									prepareCreateSymbolEdit(len(res.Lines), name),
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

			err = l.reportProgressBegin(req.Params.WorkDoneToken, "Preparing Rename", "Checking symbol for rename")
			if err != nil {
				l.trace(fmt.Sprintf("Failed to report progress begin: %s", err))
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				l.reportProgressEnd(req.Params.WorkDoneToken, "Prepare rename failed")
				return err
			}

			symbols := res.SymbolsWithin(req.Params.Position)
			if len(symbols) > 0 {
				sym := symbols[0]

				err = l.reportProgressEnd(req.Params.WorkDoneToken, "Symbol ready for rename")
				if err != nil {
					l.trace(fmt.Sprintf("Failed to report progress end: %s", err))
				}

				return l.success(h.Id, lspPrepareRenameResponse{
					Range: LspRange{
						Start: LspPosition{
							Line:      sym.Line(),
							Character: sym.Begin(),
						},
						End: LspPosition{
							Line:      sym.Line(),
							Character: sym.Begin() + len(sym.Name()),
						},
					},
					Placeholder: sym.Name(),
				})
			}

			references := res.SymbolRefsWithin(req.Params.Position)
			if len(references) > 0 {
				ref := references[0]
				return l.success(h.Id, lspPrepareRenameResponse{
					Range: LspRange{
						Start: LspPosition{
							Line:      ref.Line(),
							Character: ref.Begin(),
						},
						End: LspPosition{
							Line:      ref.Line(),
							Character: ref.Begin() + len(ref.String()),
						},
					},
					Placeholder: ref.String(),
				})
			}

			err = l.reportProgressEnd(req.Params.WorkDoneToken, "No symbol found for rename")
			if err != nil {
				l.trace(fmt.Sprintf("Failed to report progress end: %s", err))
			}

			return l.success(h.Id, struct{}{})

		case "textDocument/rename":
			req, err := read[lspRenameRequest](b)
			if err != nil {
				return err
			}

			err = l.reportProgressBegin(req.Params.WorkDoneToken, "Renaming Symbol", fmt.Sprintf("Renaming to '%s'", req.Params.NewName))
			if err != nil {
				l.trace(fmt.Sprintf("Failed to report progress begin: %s", err))
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				l.reportProgressEnd(req.Params.WorkDoneToken, "Rename failed")
				return err
			}

			chs := []lspTextEdit{}

			for _, edited := range res.SymbolsWithin(req.Params.Position) {
				for _, sym := range res.SymByName(edited.Name()) {
					chs = append(chs, prepareRenameSymbolEdit(sym, req.Params.NewName))
				}

				for _, ref := range res.SymRefByName(edited.Name()) {
					chs = append(chs, prepareRenameSymbolRefEdit(ref, req.Params.NewName))
				}
			}

			for _, edited := range res.SymbolRefsWithin(req.Params.Position) {
				for _, sym := range res.SymByName(edited.String()) {
					chs = append(chs, prepareRenameSymbolEdit(sym, req.Params.NewName))
				}

				for _, ref := range res.SymRefByName(edited.String()) {
					chs = append(chs, prepareRenameSymbolRefEdit(ref, req.Params.NewName))
				}
			}

			message := fmt.Sprintf("Renamed %d locations", len(chs))
			err = l.reportProgressEnd(req.Params.WorkDoneToken, message)
			if err != nil {
				l.trace(fmt.Sprintf("Failed to report progress end: %s", err))
			}

			return l.success(h.Id, lspWorkspaceEdit{
				Changes: map[string][]lspTextEdit{
					req.Params.TextDocument.Uri: chs,
				},
			})

		case "textDocument/inlineValue":
			req, err := read[lspInlineValueRequest](b)
			if err != nil {
				return err
			}

			_, _, err = l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			ls := []lspInlineValueText{}

			return l.success(h.Id, ls)

		case "textDocument/codeLens":
			req, err := read[lspCodeLensRequest](b)
			if err != nil {
				return err
			}

			doc, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			var cls []LspCodeLens

			if l.config.LensRefs {
				for _, sym := range res.Symbols {
					count := res.RefCounts[sym.Name()]
					if count > 0 {
						cls = append(cls, LspCodeLens{
							Range: LspRange{
								Start: LspPosition{
									Line: sym.StartLine(),
								},
								End: LspPosition{
									Line: sym.EndLine(),
								},
							},
							Command: &LspCommand{
								Title: fmt.Sprintf("refs: %d", count),
							},
						})
					}
				}
			}

			if l.prepareCodeLens != nil {
				more := l.prepareCodeLens(doc.s)
				cls = append(cls, more...)
			}

			return l.success(h.Id, cls)

		case "textDocument/inlayHint":
			req, err := read[lspInlayHintRequest](b)
			if err != nil {
				return err
			}

			doc, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			ihs := []LspInlayHint{}
			parameter := new(int)
			*parameter = 2

			padding := new(bool)
			*padding = true

			hs := res.InlayHints(req.Params.Range)

			if l.config.InlayNamed {
				for _, named := range hs.Named {
					ihs = append(ihs, LspInlayHint{
						Position: LspPosition{
							Line:      named.T.Line(),
							Character: named.T.End(),
						},
						Label:       named.Name,
						Kind:        parameter,
						PaddingLeft: padding,
					})
				}
			}

			if l.config.InlayDecoded {
				for _, decoded := range hs.Decoded {
					ihs = append(ihs, LspInlayHint{
						Position: LspPosition{
							Line:      decoded.T.Line(),
							Character: decoded.T.End(),
						},
						Label:       decoded.Value,
						Kind:        parameter,
						PaddingLeft: padding,
					})

				}
			}

			if l.prepareInlayHints != nil {
				more := l.prepareInlayHints(doc.s)
				ihs = append(ihs, more...)
			}

			return l.success(h.Id, ihs)

		case "textDocument/completion":
			req, err := read[lspCompletionRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			var ln teal.Line
			if len(res.Lines) > req.Params.Position.Line {
				ln = res.Lines[req.Params.Position.Line]
			}

			sln := ln.SublineByIndex(req.Params.Position.Character)

			ccs := []lspCompletionItem{}

			var prefix string

			mode := tealCompletionArg

			if len(sln.Tokens) == 0 {
				mode = tealCompletionOp
			} else {
				if len(sln.Tokens) > 0 {
					if req.Params.Position.Character <= sln.Tokens[0].End() {
						mode = tealCompletionOp
						prefix = sln.Tokens[0].String()
					}
				}
			}

			switch mode {
			case tealCompletionArg:
				for _, v := range res.ArgValsAt(req.Params.Position.Line, req.Params.Position.Character) {
					var d *lspCompletionItemLabelDetails
					if !v.NoValue {
						d = &lspCompletionItemLabelDetails{
							Detail: fmt.Sprintf(" = %d", v.Value),
						}
					} else if v.Signature != "" {
						d = &lspCompletionItemLabelDetails{
							Detail: fmt.Sprintf(" %s", v.Signature),
						}
					}
					ccs = append(ccs, lspCompletionItem{
						LabelDetails: d,
						Label:        v.Name,
						Documentation: lspMarkupContent{
							Kind:  "markdown",
							Value: v.Docs,
						},
					})
				}

			case tealCompletionOp:
				operator := new(int)
				*operator = 25

				snippet := 15

				snippetFormat := new(int)
				*snippetFormat = 2

				var at string
				var bt string
				for i, name := range teal.OnCompletionNames {
					if i > 0 {
						at += " "
					}

					field := fmt.Sprintf("${%d:%s}", i+1, strings.ToLower(name))

					at += field
					bt += fmt.Sprintf("%s:\n", field)

					if i < len(teal.OnCompletionNames)-1 {
						bt += fmt.Sprintf("b ${%d:then}\n", len(teal.OnCompletionNames)+2)
					}

					bt += "\n"
				}

				bt += fmt.Sprintf("${%d:then}:\n$%d", len(teal.OnCompletionNames)+2, len(teal.OnCompletionNames)+3)

				ccs = append(ccs, lspCompletionItem{
					Label:            "soc",
					Kind:             &snippet,
					Detail:           "switch on OnCompletion",
					InsertText:       fmt.Sprintf("txn OnCompletion\nswitch %s\n%s", at, bt),
					InsertTextFormat: snippetFormat,
				})

				ccs = append(ccs, lspCompletionItem{
					Label:            "func",
					Kind:             &snippet,
					Detail:           "create subroutine",
					InsertText:       "${1:sub}:\r\n\r\n\tproto ${2:0} ${3:0}\r\n\t${4}\r\n\tretsub\r\n",
					InsertTextFormat: snippetFormat,
				})

				for name := range res.Defines {
					ccs = append(ccs, lspCompletionItem{
						Label:      name,
						Kind:       operator,
						InsertText: name,
					})
				}

				for _, info := range res.AvailableOps() {
					if !strings.HasPrefix(info.Name, prefix) {
						continue
					}

					var insert string
					var format *int
					if len(info.Args) > 0 {
						var placeholders string
						for i, arg := range info.Args {
							if i > 0 {
								placeholders += " "
							}

							placeholders += fmt.Sprintf("${%d:%s}", i+1, arg.Name)
						}

						insert = fmt.Sprintf("%s %s", info.Name, placeholders)
						format = snippetFormat
					} else {
						insert = ""
						format = nil
					}

					doc := l.opDocShort(info.Name)

					ld := fmt.Sprintf("v%d", info.AppVersion)
					ccs = append(ccs, lspCompletionItem{
						Label: info.Name,
						Documentation: lspMarkupContent{
							Kind:  "markdown",
							Value: doc,
						},
						Kind:             operator,
						InsertText:       insert,
						InsertTextFormat: format,
						LabelDetails: &lspCompletionItemLabelDetails{
							Description: ld,
							Detail:      " " + info.ArgsSig,
						},
					})
				}
			}

			if len(ccs) == 0 {
				ccs = append(ccs, lspCompletionItem{
					Label: "",
				})
			}

			return l.success(h.Id, ccs)

		case "textDocument/hover":
			req, err := read[lspHoverRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			var c interface{} = struct{}{}

			s := res.DocAt(req.Params.Position.Line, req.Params.Position.Character, l.opDocShort, l.opDocExtra)
			if s != "" {
				c = lspHover{
					Contents: lspMarkupContent{
						Kind:  "plaintext",
						Value: s,
					},
				}
			}

			return l.success(h.Id, c)

		case "textDocument/definition":
			req, err := read[lspDefinitionRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			ls := []lspLocation{}

			for _, sym := range res.SymbolsForRefWithin(req.Params.Position) {
				ls = append(ls, lspLocation{
					Uri: req.Params.TextDocument.Uri,
					Range: LspRange{
						Start: LspPosition{
							Line:      sym.Line(),
							Character: sym.Begin(),
						},
						End: LspPosition{
							Line:      sym.Line(),
							Character: sym.Begin() + len(sym.Name()),
						},
					},
				})
			}

			return l.success(h.Id, ls)

		case "textDocument/formatting":
			req, err := read[lspDocumentFormattingRequest](b)
			if err != nil {
				return err
			}

			doc, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			// TODO: implement formatting
			formatted := doc.s

			return l.success(h.Id, []lspTextEdit{prepareReplaceAllTextEdit(len(res.Lines), formatted)})

		case "textDocument/signatureHelp":
			req, err := read[lspSignatureHelpRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			var sh interface{} = struct{}{}
			for _, op := range res.Ops {
				if op.Line() == req.Params.Position.Line {
					info, ok := teal.Ops.Get(teal.OpContext{
						Name:    op.String(),
						Version: res.Version,
					})
					if ok {
						_, idx, _ := res.ArgAt(req.Params.Position.Line, req.Params.Position.Character)

						active := new(int)
						*active = idx

						var doc interface{}

						fullDoc := teal.MakeFullDoc(l.opDocShort(info.Name), l.opDocExtra(info.Name))

						if fullDoc != "" {
							doc = lspMarkupContent{
								Kind:  "markdown",
								Value: fullDoc,
							}
						}

						ps := []lspParameterInformation{}

						for _, arg := range info.Args {
							ps = append(ps, lspParameterInformation{
								Label: arg.Name,
							})
						}

						sh = &lspSignatureHelp{
							Signatures: []lspSignatureInformation{
								{
									Label:           info.FullSig,
									Documentation:   doc,
									Parameters:      ps,
									ActiveParameter: active,
								},
							},
						}
					}
					break
				}
			}

			return l.success(h.Id, sh)

		case "textDocument/codeAction":
			req, err := read[lspCodeActionRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			cas := []lspCodeAction{}

			for _, red := range res.Redundants {
				if req.Params.Range.Start.Line <= red.Line() && req.Params.Range.End.Line >= red.Line() {
					kind := "quickfix"
					title := red.String()

					cas = append(cas, lspCodeAction{
						Title: title,
						Kind:  &kind,
						Command: &LspCommand{
							Title:   title,
							Command: "teal.call.remove",
							Arguments: []interface{}{
								tealRemoveCallCommandArgs{
									Uri:     req.Params.TextDocument.Uri,
									Line:    red.Line(),
									Subline: red.Subline(),
								},
							},
						},
					})
				}
			}

			for _, ref := range res.MissRefs {
				if !teal.Overlaps(req.Params.Range, ref) {
					continue
				}

				kind := "quickfix"
				cas = append(cas, lspCodeAction{
					Title: fmt.Sprintf("Create label '%s'", ref.String()),
					Kind:  &kind,
					Command: &LspCommand{
						Title:   "Create label",
						Command: "teal.label.create",
						Arguments: []interface{}{
							tealCreateLabelCommandArgs{
								Uri:  req.Params.TextDocument.Uri,
								Name: ref.String(),
							},
						},
					},
				})
			}

			hs := res.InlayHints(req.Params.Range)

			for _, named := range hs.Named {
				kind := "quickfix"
				cas = append(cas, lspCodeAction{
					Title: fmt.Sprintf("Replace with '%s'", named.Name),
					Kind:  &kind,
					Command: &LspCommand{
						Title:   "Replace with named const",
						Command: "teal.value.replace",
						Arguments: []interface{}{
							tealReplaceValueCommandArgs{
								Uri: req.Params.TextDocument.Uri,
								Range: LspRange{
									Start: LspPosition{
										Line:      named.T.Line(),
										Character: named.T.Begin(),
									},
									End: LspPosition{
										Line:      named.T.Line(),
										Character: named.T.End(),
									},
								},
								Value: named.Name,
							},
						},
					},
				})
			}

			for _, named := range hs.Decoded {
				kind := "quickfix"
				cas = append(cas, lspCodeAction{
					Title: fmt.Sprintf("Replace with literal '%s'", named.Value),
					Kind:  &kind,
					Command: &LspCommand{
						Title:   "Replace with literal",
						Command: "teal.value.replace",
						Arguments: []interface{}{
							tealReplaceValueCommandArgs{
								Uri: req.Params.TextDocument.Uri,
								Range: LspRange{
									Start: LspPosition{
										Line:      named.T.Line(),
										Character: named.T.Begin(),
									},
									End: LspPosition{
										Line:      named.T.Line(),
										Character: named.T.End(),
									},
								},
								Value: fmt.Sprintf("\"%s\"", strings.ReplaceAll(named.Value, "\"", "\\\"")),
							},
						},
					},
				})
			}

			{
				kind := "quickfix"
				for _, v := range res.Versions {
					if teal.Overlaps(req.Params.Range, v) {
						cas = append(cas, lspCodeAction{
							Title: fmt.Sprintf("Update version to %d", v.Version),
							Kind:  &kind,
							Command: &LspCommand{
								Title:   "Update version",
								Command: "teal.version.update",
								Arguments: []interface{}{
									tealUpdateVersion{
										Uri:     req.Params.TextDocument.Uri,
										Version: v.Version,
									},
								},
							},
						})
					}
				}
			}
			return l.success(h.Id, cas)
		case "textDocument/diagnostic":
			req, err := read[lspDiagnosticRequest](b)
			if err != nil {
				return err
			}

			doc := l.docs[req.Params.TextDocument.Uri]
			ds := []LspDiagnostic{}
			if doc != nil {
				if l.prepareDiagnostics != nil {
					ds = append(ds, l.prepareDiagnostics(doc.s)...)
				} else {
					ds = prepareDiagnostics(doc.Results())
				}
			}

			return l.success(h.Id, lspFullDocumentDiagnosticReport{
				Kind:  "full",
				Items: ds,
			})

		case "textDocument/documentHighlight":
			req, err := read[lspDocumentHighlightRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			hs := []lspDocumentHighlight{}

			name := res.SymOrRefAt(req.Params.Position)

			for _, sym := range res.SymByName(name) {
				hs = append(hs, prepareSymbolHighlight(sym))
			}

			for _, ref := range res.SymRefByName(name) {
				hs = append(hs, prepareSymbolRefHighlight(ref))
			}

			return l.success(h.Id, hs)
		case "textDocument/documentSymbol":
			req, err := read[lspDocumentSymbolRequest](b)
			if err != nil {
				return err
			}

			doc, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			syms := []LspDocumentSymbol{}
			for _, s := range res.Symbols {
				syms = append(syms, prepareSymbol(s))
			}

			if l.prepareSymbols != nil {
				more := l.prepareSymbols(doc.s)
				syms = append(syms, more...)
			}

			return l.success(h.Id, syms)

		case "textDocument/semanticTokens/full":
			req, err := read[lspSemanticTokensFullRequest](b)
			if err != nil {
				return err
			}

			_, res, err := l.prepare(req.Params.TextDocument.Uri)
			if err != nil {
				return err
			}

			st := teal.SemanticTokens{}

			for _, v := range res.Bools {
				st = append(st, prepareValueSemToken(v))
			}

			for _, m := range res.Macros {
				st = append(st, prepareMacroSemToken(m))
			}

			for _, op := range res.Ops {
				if op.Type() == teal.TokenValue {
					st = append(st, prepareOpSemToken(op))
				}
			}

			for _, v := range res.Numbers {
				st = append(st, prepareNumberSemToken(v))
			}

			for _, v := range res.Strings {
				st = append(st, prepareStringSemToken(v))
			}

			for _, v := range res.Keywords {
				st = append(st, prepareKeywordSemToken(v))
			}

			for _, t := range res.Tokens {
				switch t.Type() {
				case teal.TokenComment:
					st = append(st, prepareCommentSemToken(t))
				}
			}

			for _, s := range res.Symbols {
				st = append(st, prepareSymbolSemToken(s))
			}

			for _, s := range res.SymbolRefs {
				st = append(st, prepareSymbolRefSemToken(s))
			}

			data := st.Encode()

			return l.success(h.Id, lspSemanticTokens{
				Data: data,
			})

		case "initialize":
			req, err := read[lspInitializeRequest](b)
			if err != nil {
				return err
			}

			if req.Params != nil {
				if req.Params.InitializationOptions != nil {
					if req.Params.InitializationOptions.SemanticTokens != nil {
						l.config.SemanticTokens = *req.Params.InitializationOptions.SemanticTokens
					}
					if req.Params.InitializationOptions.InlayNamed != nil {
						l.config.InlayNamed = *req.Params.InitializationOptions.InlayNamed
					}
					if req.Params.InitializationOptions.InlayDecoded != nil {
						l.config.InlayDecoded = *req.Params.InitializationOptions.InlayDecoded
					}
					if req.Params.InitializationOptions.LensRefs != nil {
						l.config.LensRefs = *req.Params.InitializationOptions.LensRefs
					}
				}
			}

			sync := new(int)
			*sync = 1

			definition := new(bool)
			*definition = true

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

			formatting := new(bool)
			*formatting = true

			hover := new(bool)
			*hover = true

			inlayHint := new(bool)
			if l.config.InlayNamed || l.config.InlayDecoded {
				*inlayHint = true
			}

			inlineValue := new(bool)
			*inlineValue = true

			var semanticTokensProvider *lspSemanticTokensProvider

			if l.config.SemanticTokens {
				semanticTokensProvider = &lspSemanticTokensProvider{
					Full: fullSemantic,
					Legend: lspSemanticTokensLegend{
						TokenTypes:     []string{"keyword", "string", "comment", "method", "macro", "value", "number", "operator", "function"},
						TokenModifiers: []string{},
					},
				}
			}

			return l.success(h.Id, lspInitializeResult{
				Capabilities: &lspServerCapabilities{
					TextDocumentSync:          sync,
					DocumentHighlightProvider: highlight,
					DiagnosticProvider:        &lspDiagnosticProvider{},
					DocumentSymbolProvider:    symbol,
					CodeActionProvider:        action,
					ExecuteCommandProvider: &lspExecuteCommandProvider{
						Commands: []string{
							"teal.label.create",
							"teal.label.remove",
							"teal.value.replace",
							"teal.call.remove",
							"teal.version.update",
							"teal.pc.resolve",
						},
					},
					RenameProvider: &lspRenameOptions{
						PrepareProvider:  rename,
						WorkDoneProgress: rename,
					},
					SemanticTokensProvider: semanticTokensProvider,
					CompletionProvider: &lspCompletionProvider{
						TriggerCharacters: []string{" "},
					},
					DocumentFormattingProvider: formatting,
					DefinitionProvider:         definition,
					HoverProvider:              hover,
					SignatureHelpProvider:      &lspSignatureHelpOptions{},
					InlayHintProvider:          inlayHint,
					InlineValueProvider:        inlineValue,
					CodeLensProvider:           &lspCodeLensProvider{},
				},
			})
		default:
			return errors.New("unknown method")
		}
	}

	return nil
}

func prepareReplaceAllTextEdit(lines int, formatted string) lspTextEdit {
	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      0,
				Character: 0,
			},
			End: LspPosition{
				Line:      lines,
				Character: 0,
			},
		},
		NewText: formatted,
	}
}

func prepareRenameSymbolRefEdit(ref teal.Token, newName string) lspTextEdit {
	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      ref.Line(),
				Character: ref.Begin(),
			},
			End: LspPosition{
				Line:      ref.Line(),
				Character: ref.End(),
			},
		},
		NewText: newName,
	}
}

func prepareRenameSymbolEdit(sym teal.Symbol, newName string) lspTextEdit {
	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      sym.Line(),
				Character: sym.Begin(),
			},
			End: LspPosition{
				Line:      sym.Line(),
				Character: sym.Begin() + len(sym.Name()),
			},
		},
		NewText: newName,
	}
}

func prepareCreateSymbolEdit(lines int, name string) lspTextEdit {
	s := fmt.Sprintf("\r\n%s:\r\n", name)

	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      lines,
				Character: 0,
			},
			End: LspPosition{
				Line:      lines,
				Character: len(s),
			},
		},
		NewText: s,
	}
}

func prepareRemoveSymbolEdit(sym teal.Symbol) lspTextEdit {
	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      sym.Line(),
				Character: sym.Begin(),
			},
			End: LspPosition{
				Line:      sym.Line(),
				Character: sym.End(),
			},
		},
		NewText: "",
	}
}

func prepareSymbolRefSemToken(s teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      s.Line(),
		Index:     s.Begin(),
		Length:    s.End() - s.Begin(),
		Type:      semanticTokenString,
		Modifiers: 0,
	}
}

func prepareSymbolSemToken(s teal.Symbol) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      s.Line(),
		Index:     s.Begin(),
		Length:    s.End() - s.Begin(),
		Type:      semanticTokenMethod,
		Modifiers: 0,
	}
}

func prepareCommentSemToken(t teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      t.Line(),
		Index:     t.Begin(),
		Length:    t.End() - t.Begin(),
		Type:      semanticTokenComment,
		Modifiers: 0,
	}
}

func prepareKeywordSemToken(v teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      v.Line(),
		Index:     v.Begin(),
		Length:    v.End() - v.Begin(),
		Type:      semanticTokenKeyword,
		Modifiers: 0,
	}
}

func prepareStringSemToken(v teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      v.Line(),
		Index:     v.Begin(),
		Length:    v.End() - v.Begin(),
		Type:      semanticTokenString,
		Modifiers: 0,
	}
}

func prepareNumberSemToken(v teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      v.Line(),
		Index:     v.Begin(),
		Length:    v.End() - v.Begin(),
		Type:      semanticTokenNumber,
		Modifiers: 0,
	}
}

func prepareOpSemToken(op teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      op.Line(),
		Index:     op.Begin(),
		Length:    op.End() - op.Begin(),
		Type:      semanticTokenKeyword,
		Modifiers: 0,
	}
}

func prepareMacroSemToken(m teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      m.Line(),
		Index:     m.Begin(),
		Length:    m.End() - m.Begin(),
		Type:      semanticTokenMacro,
		Modifiers: 0,
	}
}

func prepareValueSemToken(v teal.Token) teal.SemanticToken {
	return teal.SemanticToken{
		Line:      v.Line(),
		Index:     v.Begin(),
		Length:    v.End() - v.Begin(),
		Type:      semanticTokenValue,
		Modifiers: 0,
	}
}

func prepareSymbol(s teal.Symbol) LspDocumentSymbol {
	r := LspRange{
		Start: LspPosition{
			Line:      s.Line(),
			Character: s.Begin(),
		},
		End: LspPosition{
			Line:      s.Line(),
			Character: s.End(),
		},
	}

	ds := LspDocumentSymbol{
		Name:           s.Name(),
		Kind:           LspSymbolKindMethod,
		Range:          r,
		SelectionRange: r,
	}

	return ds
}

func prepareSymbolRefHighlight(ref teal.Token) lspDocumentHighlight {
	return lspDocumentHighlight{
		Range: LspRange{
			Start: LspPosition{
				Line:      ref.Line(),
				Character: ref.Begin(),
			},
			End: LspPosition{
				Line:      ref.Line(),
				Character: ref.End(),
			},
		},
		Kind: &symbolHighlightKind,
	}
}

var symbolHighlightKind = 1

func prepareSymbolHighlight(sym teal.Symbol) lspDocumentHighlight {
	return lspDocumentHighlight{
		Range: LspRange{
			Start: LspPosition{
				Line:      sym.Line(),
				Character: sym.Begin(),
			},
			End: LspPosition{
				Line:      sym.Line(),
				Character: sym.Begin() + len(sym.Name()),
			},
		},
		Kind: &symbolHighlightKind,
	}
}

func prepareDiagnostics(res *teal.ProcessResult) []LspDiagnostic {
	lds := []LspDiagnostic{}

	for _, d := range res.Diagnostics {
		sev := int(d.Severity())

		lds = append(lds, LspDiagnostic{
			Range: LspRange{
				Start: LspPosition{
					Line:      d.Line(),
					Character: d.Begin(),
				},
				End: LspPosition{
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

func prepareRemoveLineEdit(line int) lspTextEdit {
	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      line,
				Character: 0,
			},
			End: LspPosition{
				Line:      line + 1,
				Character: 0,
			},
		},
		NewText: "",
	}
}

func prepareRemoveSublineEdit(res *teal.ProcessResult, line int, subline int) lspTextEdit {
	b := res.Lines[line].SublineBegin(subline)
	e := res.Lines[line].SublineEnd(subline)

	return lspTextEdit{
		Range: LspRange{
			Start: LspPosition{
				Line:      line,
				Character: b,
			},
			End: LspPosition{
				Line:      line,
				Character: e,
			},
		},
		NewText: "",
	}
}

func prepareVersionEdit(rg teal.Range, version uint64) lspTextEdit {
	if rg != nil {
		return lspTextEdit{
			Range: LspRange{
				Start: LspPosition{
					Line:      rg.StartLine(),
					Character: rg.StartCharacter(),
				},
				End: LspPosition{
					Line:      rg.EndLine(),
					Character: rg.EndCharacter(),
				},
			},
			NewText: fmt.Sprintf("%d", version),
		}
	} else {
		return lspTextEdit{
			Range: LspRange{
				Start: LspPosition{
					Line:      0,
					Character: 0,
				},
				End: LspPosition{
					Line:      0,
					Character: 0,
				},
			},
			NewText: fmt.Sprintf("#pragma version %d\r\n", version),
		}
	}
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

			if errors.Is(err, io.EOF) {
				break
			}
		}
	}

	return l.exitCode, nil
}
