package dbg

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
	"github.com/dragmz/teal"
	"github.com/dragmz/teal/sim"
	"github.com/pkg/errors"
)

var yes = new(bool)

func init() {
	*yes = true
}

type dbg struct {
	config DbgConfig
	ac     *algod.Client

	id int

	shutdown bool

	exit     bool
	exitCode int

	tp *textproto.Reader
	w  *bufio.Writer

	debug  *bufio.Writer
	replay *models.SimulateResponse

	bid int
	bs  []dbgBreakpoint

	vm *dbgVm
	lz int
	cz int
}

type dbgVm struct {
	tvm  *sim.Vm
	name string
	path string
}

type dbgBreakpoint struct {
	id   int
	l    int
	name string
	path string
}

type DbgOption func(l *dbg) error

type DbgAppConfig struct {
	Debug    *bool            `json:"debug,omitempty"`
	Args     []string         `json:"args,omitempty"`
	Accounts []string         `json:"accounts,omitempty"`
	Apps     []uint64         `json:"apps,omitempty"`
	Assets   []uint64         `json:"assets,omitempty"`
	Schema   *DbgConfigSchema `json:"schema,omitempty"`
}

type DbgConfigSchemaValues struct {
	Bytes   uint8 `json:"bytes,omitempty"`
	Uint64s uint8 `json:"uints,omitempty"`
}

type DbgConfigSchema struct {
	Local  DbgConfigSchemaValues `json:"local,omitempty"`
	Global DbgConfigSchemaValues `json:"global,omitempty"`
}

type DbgConfig struct {
	Args     []string         `json:"args,omitempty"`
	Accounts []string         `json:"accounts,omitempty"`
	Apps     []uint64         `json:"apps,omitempty"`
	Assets   []uint64         `json:"assets,omitempty"`
	Schema   *DbgConfigSchema `json:"schema,omitempty"`

	Create DbgAppConfig `json:"create,omitempty"`
	Call   DbgAppConfig `json:"call,omitempty"`
}

func WithConfig(c DbgConfig) DbgOption {
	return func(l *dbg) error {
		l.config = c
		return nil
	}
}

func WithDebug(w io.Writer) DbgOption {
	return func(l *dbg) error {
		l.debug = bufio.NewWriter(w)
		return nil
	}
}

func WithAlgod(addr string, token string) DbgOption {
	return func(l *dbg) error {
		var err error

		l.ac, err = algod.MakeClient(addr, token)
		if err != nil {
			return errors.Wrap(err, "failed to make algod client")
		}

		return nil
	}
}

func WithReplay(reply models.SimulateResponse) DbgOption {
	return func(l *dbg) error {
		l.replay = &reply
		return nil
	}
}

func New(r io.Reader, w io.Writer, opts ...DbgOption) (*dbg, error) {
	l := &dbg{
		tp: textproto.NewReader(bufio.NewReader(r)),
		w:  bufio.NewWriter(w),
	}

	for _, opt := range opts {
		err := opt(l)
		if err != nil {
			return nil, errors.Wrap(err, "failed to set dbg option")
		}
	}

	return l, nil
}

type dapHeader struct {
	Seq  int    `json:"seq"`
	Type string `json:"type"`
}

type dapRequestHeader struct {
	Command string `json:"command"`
}

type dapRequest[T any] struct {
	Arguments T `json:"arguments,omitempty"`
}

type dapInitializeRequestParams struct {
	ClientID        string `json:"clientID,omitempty"`
	ClientName      string `json:"clientName,omitempty"`
	AdapterID       string `json:"adapterID,omitempty"`
	LinesStartAt1   *bool  `json:"linesStartAt1,omitempty"`
	ColumnsStartAt1 *bool  `json:"columnsStartAt1,omitempty"`
	PathFormat      string `json:"pathFormat,omitempty"`
}

type dapLaunchRequestParams struct {
	Program string `json:"program"`
}

type dapThreadEventParams struct {
	Reason   string `json:"reason"`
	ThreadId *int   `json:"threadId"`
}

type dapStackTraceRequestParams struct {
	ThreadId   int  `json:"threadId"`
	StartFrame *int `json:"startFrame,omitempty"`
	Levels     *int `json:"levels,omitempty"`
}

type dapStackFrame struct {
	Id        int        `json:"id"`
	Name      string     `json:"name"`
	Source    *dapSource `json:"source,omitempty"`
	Line      int        `json:"line"`
	EndLine   *int       `json:"endLine,omitempty"`
	Column    int        `json:"column"`
	EndColumn *int       `json:"endColumn,omitempty"`
}

type dapStackTraceResponse struct {
	StackFrames []dapStackFrame `json:"stackFrames"`
	TotalFrames *int            `json:"totalFrames,omitempty"`
}

type dapScopesRequestParams struct {
	FrameId int `json:"frameId"`
}

type dapVariablesRequestParams struct {
	VariablesReference int  `json:"variablesReference"`
	Start              *int `json:"start,omitempty"`
	Count              *int `json:"count,omitempty"`
}

type dapVariablesResponse struct {
	Variables []dapVariable `json:"variables"`
}

type dapVariable struct {
	Name               string `json:"name"`
	Value              string `json:"value"`
	VariablesReference int    `json:"variablesReference"`
}

type dapScopesResponse struct {
	Scopes []dapScope `json:"scopes"`
}

type dapScope struct {
	Name               string `json:"name"`
	VariablesReference int    `json:"variablesReference"`
	IndexedVariables   *int   `json:"indexedVariables,omitempty"`
	Expensive          bool   `json:"expensive"`
}

type dapNextRequestParams struct {
	ThreadId     int    `json:"threadId"`
	SingleThread *bool  `json:"singleThread,omitempty"`
	Granularity  string `json:"granularity,omitempty"`
}

type dapContinueRequestParams struct {
	ThreadId int `json:"threadId"`
}

type dapBreakpointLocationsRequestParams struct {
	Source    dapSource `json:"source"`
	Line      int       `json:"line"`
	Column    *int      `json:"column,omitempty"`
	EndLine   *int      `json:"endLine,omitempty"`
	EndColumn *int      `json:"endColumn,omitempty"`
}

type dapInitializeRequest dapRequest[*dapInitializeRequestParams]
type dapSetBreakpointsRequest dapRequest[*dapSetBreakpointsRequestParams]
type dapLaunchRequest dapRequest[*dapLaunchRequestParams]
type dapStackTraceRequest dapRequest[*dapStackTraceRequestParams]
type dapScopesRequest dapRequest[*dapScopesRequestParams]
type dapVariablesRequest dapRequest[*dapVariablesRequestParams]
type dapNextRequest dapRequest[*dapNextRequestParams]
type dapContinueRequest dapRequest[*dapContinueRequestParams]
type dapBreakpointLocationsRequest dapRequest[*dapBreakpointLocationsRequestParams]

type dapCapabilities struct {
	SupportsInstructionBreakpoints     *bool `json:"supportsInstructionBreakpoints,omitempty"`
	SupportsHitConditionalBreakpoints  *bool `json:"supportsHitConditionalBreakpoints,omitempty"`
	SupportsFunctionBreakpoints        *bool `json:"supportsFunctionBreakpoints,omitempty"`
	SupportsModulesRequest             *bool `json:"supportsModulesRequest,omitempty"`
	SupportsConfigurationDoneRequest   *bool `json:"supportsConfigurationDoneRequest,omitempty"`
	SupportsBreakpointLocationsRequest *bool `json:"supportsBreakpointLocationsRequest,omitempty"`
}

type dapResponse struct {
	Seq        int    `json:"seq"`
	Type       string `json:"type"`
	RequestSeq int    `json:"request_seq"`
	Success    bool   `json:"success"`
	Command    string `json:"command"`
	Message    string `json:"message,omitempty"`
	Body       any    `json:"body,omitempty"`
}

type dapContinueResponse struct {
	AllThreadsContinued *bool `json:"allThreadsContinued,omitempty"`
}

type dapThreadsResponse struct {
	Threads []dapThread `json:"threads"`
}

type dapSetBreakpointsResponse struct {
	Breakpoints []dapBreakpoint `json:"breakpoints"`
}

type dapBreakpointLocationsResponse struct {
	Breakpoints []dapBreakpointLocation `json:"breakpoints"`
}

type dapSourceBreakpoint struct {
	Line int `json:"line"`
}

type dapSource struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

type dapSetBreakpointsRequestParams struct {
	Source      dapSource             `json:"source"`
	Breakpoints []dapSourceBreakpoint `json:"breakpoints,omitempty"`
}

type dapBreakpoint struct {
	Id       *int       `json:"id,omitempty"`
	Verified bool       `json:"verified"`
	Message  string     `json:"message,omitempty"`
	Source   *dapSource `json:"source,omitempty"`
	Line     *int       `json:"line,omitempty"`
}

type dapBreakpointLocation struct {
	Line      int  `json:"line"`
	Column    *int `json:"column,omitempty"`
	EndLine   *int `json:"endLine,omitempty"`
	EndColumn *int `json:"endColumn,omitempty"`
}

type dapThread struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type dapEvent struct {
	Seq   int    `json:"seq"`
	Type  string `json:"type"`
	Event string `json:"event"`
	Body  any    `json:"body,omitempty"`
}

type dapProcessEventParams struct {
	Name        string `json:"name"`
	StartMethod string `json:"startMethod,omitempty"`
}

type dapStoppedEventParams struct {
	Reason            string `json:"reason"`
	Description       string `json:"description,omitempty"`
	ThreadId          *int   `json:"threadId,omitempty"`
	AllThreadsStopped *bool  `json:"allThreadsStopped,omitempty"`
	HitBreakpointIds  []int  `json:"hitBreakpointIds,omitempty"`
	PreserveFocusHint *bool  `json:"preserveFocusHint,omitempty"`
	Text              string `json:"text,omitempty"`
}

type dapExitedEventParams struct {
	ExitCode int `json:"exitCode"`
}

func readInto(b []byte, v interface{}) error {
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}

func (l *dbg) reply(id int, cmd string, msg string, result interface{}, err interface{}) error {
	l.id++
	return l.write(dapResponse{
		Seq:        l.id,
		Type:       "response",
		RequestSeq: id,
		Success:    err == nil,
		Command:    cmd,
		Message:    msg,
		Body:       result,
	})
}

func read[T any](b []byte) (T, error) {
	var v T

	err := readInto(b, &v)
	if err != nil {
		return v, err
	}

	return v, nil
}

func (l *dbg) notify(event string, params interface{}) error {
	l.id++
	return l.write(dapEvent{
		Seq:   l.id,
		Type:  "event",
		Event: event,
		Body:  params,
	})
}

func makeArgs(ca []string) ([][]byte, error) {
	args := make([][]byte, len(ca))

	for i, arg := range ca {
		var bs []byte
		var err error

		if strings.HasPrefix(arg, "b64:") {
			bs, err = base64.StdEncoding.DecodeString(arg[4:])
			if err != nil {
				return nil, err
			}
		} else {
			bs = []byte(arg)
		}

		args[i] = bs
	}

	return args, nil
}

func makeRunConfig(config DbgConfig, dbg DbgAppConfig, debugDefault bool) (sim.AppRunConfig, error) {
	var run sim.AppRunConfig

	var err error
	var args [][]byte

	if dbg.Args != nil {
		args, err = makeArgs(dbg.Args)
	} else {
		args, err = makeArgs(config.Args)
	}
	if err != nil {
		return run, err
	}

	var debug bool
	if dbg.Debug != nil {
		debug = *dbg.Debug
	} else {
		debug = debugDefault
	}

	accounts := dbg.Accounts
	if accounts == nil {
		accounts = config.Accounts
	}

	apps := dbg.Apps
	if apps == nil {
		apps = config.Apps
	}

	assets := dbg.Assets
	if assets == nil {
		assets = config.Assets
	}

	schema := dbg.Schema
	if schema == nil {
		schema = config.Schema
	}

	var appSchema sim.AppRunConfigSchema
	if schema != nil {
		appSchema = sim.AppRunConfigSchema{
			Local: sim.AppRunConfigSchemaValues{
				Bytes:   schema.Local.Bytes,
				Uint64s: schema.Local.Uint64s,
			},
			Global: sim.AppRunConfigSchemaValues{
				Bytes:   schema.Global.Bytes,
				Uint64s: schema.Global.Uint64s,
			},
		}
	}

	run = sim.AppRunConfig{
		Debug:    debug,
		Args:     args,
		Accounts: accounts,
		Apps:     apps,
		Assets:   assets,
		Schema:   appSchema,
	}

	return run, nil
}

func (l *dbg) handle(h dapHeader, b []byte) error {
	switch h.Type {
	case "request":
		req, err := read[dapRequestHeader](b)
		if err != nil {
			return err
		}

		switch req.Command {
		case "initialize":
			ireq, err := read[dapInitializeRequest](b)
			if err != nil {
				return err
			}

			if ireq.Arguments.LinesStartAt1 == nil || *ireq.Arguments.LinesStartAt1 {
				l.lz = 1
			}

			if ireq.Arguments.ColumnsStartAt1 == nil || *ireq.Arguments.ColumnsStartAt1 {
				l.cz = 1
			}

			err = l.reply(h.Seq, req.Command, "", dapCapabilities{
				SupportsConfigurationDoneRequest: yes,
			}, nil)
			if err != nil {
				return err
			}

			err = l.notify("initialized", nil)
			if err != nil {
				return err
			}
		case "launch":
			lreq, err := read[dapLaunchRequest](b)
			if err != nil {
				return err
			}

			bs, err := os.ReadFile(lreq.Arguments.Program)
			if err != nil {
				return err
			}

			createConfig, err := makeRunConfig(l.config, l.config.Create, true)
			if err != nil {
				return err
			}

			callConfig, err := makeRunConfig(l.config, l.config.Call, true)
			if err != nil {
				return err
			}

			var r sim.Result
			if l.replay != nil {
				r, err = sim.Replay(bs, *l.replay, sim.ReplayConfig{
					Ac: l.ac,
				})
				if err != nil {
					return err
				}
			} else {
				src := string(bs)

				pr := teal.Process(src)

				clear := "int 1"
				if pr.Version > 1 {
					clear = fmt.Sprintf("#pragma version %d\r\n", pr.Version) + clear
				}

				r, err = sim.Run([]byte(src), []byte(clear), sim.RunConfig{
					Ac:     l.ac,
					Sender: "F77YBQEP4EAJYCQPS4GYEW2WWJXU6DQ4OJHRYSV74UXHOTRWXYRN7HNP3U",
					Create: createConfig,
					Call:   callConfig,
				})

				if err != nil {
					return err
				}
			}

			tvm, err := sim.NewVm(r)
			if err != nil {
				return err
			}

			l.vm = &dbgVm{
				tvm:  tvm,
				name: lreq.Arguments.Program,
				path: lreq.Arguments.Program,
			}

			err = l.reply(h.Seq, req.Command, "", nil, nil)
			if err != nil {
				return err
			}

			err = l.notify("process", dapProcessEventParams{
				Name:        l.vm.name,
				StartMethod: "launch",
			})
			if err != nil {
				return err
			}

			var tid *int
			if l.vm.tvm.Branch != nil {
				tid = new(int)
				*tid = l.vm.tvm.Branch.Id
			}

			return l.notify("stopped", dapStoppedEventParams{
				Reason:            "entry",
				AllThreadsStopped: yes,
				ThreadId:          tid,
			})

		case "disconnect":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "evaluate":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "setFunctionBreakpoints":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "setBreakpoints":
			ser, err := read[dapSetBreakpointsRequest](b)
			if err != nil {
				return err
			}

			bs := []dapBreakpoint{}

			if l.vm != nil {

				l.bs = []dbgBreakpoint{}
				for _, b := range ser.Arguments.Breakpoints {
					bt := dbgBreakpoint{
						id:   b.Line - l.lz, // TODO: line number used as id - review the design
						l:    b.Line - l.lz,
						name: ser.Arguments.Source.Name,
						path: ser.Arguments.Source.Path,
					}

					l.bs = append(l.bs, bt)
				}

				lns := []int{}
				for _, b := range l.bs {
					lns = append(lns, b.l)
				}

				verified := l.vm.tvm.SetBreakpoints(lns)

				for _, b := range l.bs {
					id := b.id
					ln := b.l + l.lz

					bs = append(bs, dapBreakpoint{
						Id:       &id,
						Verified: verified[b.id], // TODO: line number used as id - review the design
						Line:     &ln,
						Source:   &ser.Arguments.Source,
					})
				}
			}

			err = l.reply(h.Seq, req.Command, "", dapSetBreakpointsResponse{
				Breakpoints: bs,
			}, nil)
			if err != nil {
				return err
			}

			return nil

		case "setInstructionBreakpoints":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "setExceptionBreakpoints":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "threads":
			ts := []dapThread{}
			if l.vm != nil {
				for _, b := range l.vm.tvm.Branches {
					ts = append(ts, dapThread{
						Id:   b.Id,
						Name: fmt.Sprintf("Call %d", b.Id),
					})
				}
			}

			return l.reply(h.Seq, req.Command, "", dapThreadsResponse{
				Threads: ts,
			}, nil)
		case "pause":
			err := l.reply(h.Seq, req.Command, "", nil, nil)
			if err != nil {
				return err
			}

			l.vm.tvm.Pause = true

			var tid *int
			if l.vm.tvm.Branch != nil {
				tid = new(int)
				*tid = l.vm.tvm.Branch.Id
			}

			return l.notify("stopped", dapStoppedEventParams{
				Reason:            "pause",
				AllThreadsStopped: yes,
				ThreadId:          tid,
			})
		case "continue":
			creq, err := read[dapContinueRequest](b)
			if err != nil {
				return err
			}
			err = l.reply(h.Seq, req.Command, "", dapContinueResponse{
				AllThreadsContinued: yes,
			}, nil)

			if err != nil {
				return err
			}

			if l.vm != nil {
				l.vm.tvm.Run()

				if l.vm.tvm.Error != nil {
					return l.notify("stopped", dapStoppedEventParams{
						Reason:            "exception",
						AllThreadsStopped: yes,
						Description:       fmt.Sprintf("Error: %s", l.vm.tvm.Error),
					})
				}

				if len(l.vm.tvm.Triggered) > 0 {
					var tid int
					ids := l.vm.tvm.Triggered
					if _, ok := ids[creq.Arguments.ThreadId]; ok {
						tid = creq.Arguments.ThreadId
					} else {
						for id := range l.vm.tvm.Triggered {
							tid = id
							break
						}
					}

					return l.notify("stopped", dapStoppedEventParams{
						Reason:            "breakpoint",
						AllThreadsStopped: yes,
						ThreadId:          &tid,
						HitBreakpointIds:  ids[tid],
					})
				} else if l.vm.tvm.Branch == nil {
					return l.notify("stopped", dapStoppedEventParams{
						Reason:            "breakpoint",
						AllThreadsStopped: yes,
					})
				}
			}

		case "configurationDone":
			return l.reply(h.Seq, req.Command, "", nil, nil)
		case "next":
			nreq, err := read[dapNextRequest](b)
			if err != nil {
				return err
			}

			err = l.reply(h.Seq, req.Command, "", nil, nil)
			if err != nil {
				return err
			}

			var tid *int
			if l.vm.tvm.Branch != nil {
				tid = new(int)
				*tid = l.vm.tvm.Branch.Id
			}

			if l.vm != nil {
				l.vm.tvm.Switch(nreq.Arguments.ThreadId)
				l.vm.tvm.Step()

				if l.vm.tvm.Error != nil {
					return l.notify("stopped", dapStoppedEventParams{
						Reason:            "exception",
						AllThreadsStopped: yes,
						Description:       fmt.Sprintf("Error: %s", l.vm.tvm.Error),
					})
				}
			}

			return l.notify("stopped", dapStoppedEventParams{
				Reason:            "step",
				AllThreadsStopped: yes,
				ThreadId:          tid,
			})
		case "variables":
			vreq, err := read[dapVariablesRequest](b)
			if err != nil {
				return err
			}

			vs := []dapVariable{}

			if vreq.Arguments.VariablesReference > 0 {
				r := vreq.Arguments.VariablesReference - 1
				f := r / 10
				i := r % 10
				// TODO: variables lifetime

				if l.vm != nil {
					for _, b := range l.vm.tvm.Branches {
						if b.Id == f {
							switch i {
							case 0:
							case 1:
								for i := len(b.Stack.Items) - 1; i >= 0; i-- {
									vs = append(vs, dapVariable{
										Name:  strconv.Itoa(i),
										Value: b.Stack.Items[i].String(),
									})
								}
							case 2:
								s := 0
								e := len(l.vm.tvm.Branch.Scratch.Items)

								if vreq.Arguments.Start != nil {
									s = *vreq.Arguments.Start
								}

								if vreq.Arguments.Count != nil {
									e = s + *vreq.Arguments.Count
								}

								for i := s; i < e; i++ {
									v := l.vm.tvm.Branch.Scratch.Items[i]
									if !v.IsNone() {
										vs = append(vs, dapVariable{
											Name:  strconv.Itoa(i),
											Value: v.String(),
										})
									}
								}
							case 3:
								for i := 0; i < len(b.Global.Items); i++ {
									v := b.Global.Items[i]

									var sv string
									switch v.Type {
									case sim.VmStateItemTypeBytes:
										sv = string(v.Bytes)
									case sim.VmStateItemTypeUint:
										sv = strconv.FormatUint(v.Uint, 10)
									}

									vs = append(vs, dapVariable{
										Name:  string(v.Key),
										Value: sv,
									})
								}
							case 4:
								for i := 0; i < len(b.Boxes.Items); i++ {
									v := b.Boxes.Items[i]

									vs = append(vs, dapVariable{
										Name:  string(v.Key),
										Value: string(v.Value),
									})
								}
							}
						}
					}
				}
			}
			return l.reply(h.Seq, "variables", "", dapVariablesResponse{
				Variables: vs,
			}, nil)

		case "scopes":
			sreq, err := read[dapScopesRequest](b)
			if err != nil {
				return err
			}

			ss := []dapScope{}

			if l.vm != nil {
				for _, b := range l.vm.tvm.Branches {
					if b.Id == sreq.Arguments.FrameId {
						stacklen := len(b.Stack.Items)
						ss = append(ss, dapScope{
							Name:               "Stack",
							VariablesReference: 2 + 10*b.Id, // TODO: make sure var ref calculation is reliable
							IndexedVariables:   &stacklen,
						})

						scratchlen := 256
						ss = append(ss, dapScope{
							Name:               "Scratch",
							VariablesReference: 3 + 10*b.Id,
							IndexedVariables:   &scratchlen,
						})

						globallen := len(b.Global.Items)
						ss = append(ss, dapScope{
							Name:               "Global",
							VariablesReference: 4 + 10*b.Id,
							IndexedVariables:   &globallen,
						})

						// TODO: local

						boxeslen := len(b.Boxes.Items)
						ss = append(ss, dapScope{
							Name:               "Boxes",
							VariablesReference: 5 + 10*b.Id,
							IndexedVariables:   &boxeslen,
						})
					}
				}
			}

			return l.reply(h.Seq, "scopes", "", dapScopesResponse{
				Scopes: ss,
			}, nil)

		case "stackTrace":
			sreq, err := read[dapStackTraceRequest](b)
			if err != nil {
				return err
			}

			sf := []dapStackFrame{}

			if l.vm != nil {
				for _, b := range l.vm.tvm.Branches {
					if b.Id == sreq.Arguments.ThreadId {
						line := b.Line
						name := b.Name
						for i := len(b.Frames) - 1; i >= 0; i-- {
							f := b.Frames[i]
							sf = append(sf, dapStackFrame{
								Id:     b.Id,
								Name:   name,
								Line:   line + l.lz,
								Column: l.cz,
								Source: &dapSource{
									Name: l.vm.name,
									Path: l.vm.path,
								},
							})

							line = f.Return
							name = f.Name
						}
						sf = append(sf, dapStackFrame{
							Id:     b.Id,
							Name:   name,
							Line:   line + l.lz,
							Column: l.cz,
							Source: &dapSource{
								Name: l.vm.name,
								Path: l.vm.path,
							},
						})
					}
				}
			}

			one := new(int)
			*one = 1

			return l.reply(h.Seq, req.Command, "", dapStackTraceResponse{
				StackFrames: sf,
				TotalFrames: one,
			}, nil)
		}
	}
	return nil
}

func (l *dbg) write(v interface{}) error {
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

func (l *dbg) trace(s string) {
	if l.debug == nil {
		return
	}

	l.debug.WriteString(s)
	l.debug.WriteString("\n")

	l.debug.Flush()
}

func (l *dbg) Run() (int, error) {
	l.trace("TEAL dbg running..")
	defer func() {
		l.trace("TEAL dbg exited.")
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

			var jh dapHeader
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
