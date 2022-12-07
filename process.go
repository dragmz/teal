package teal

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ProcessResult struct {
	Diagnostics []Diagnostic
	Symbols     []Symbol
}

func Process(source string) ProcessResult {
	ts, diag := readTokens(source)

	lines := [][]Token{}

	p := 0
	for i := 0; i < len(ts); i++ {
		t := ts[i]

		j := i + 1
		eol := t.Type() == TokenEol

		if eol || j == len(ts) {
			k := j
			if eol {
				k--
			}

			lines = append(lines, ts[p:k])
			p = j
		}
	}

	for li, l := range lines {
		for i := 0; i < len(l); i++ {
			if l[i].Type() == TokenComment {
				lines[li] = l[:i]
			}
		}
	}

	var res Listing

	for li, l := range lines {
		args := &arguments{ts: l}

		failAt := func(l int, b int, e int, err error) {
			diag = append(diag, parseError{l: l, b: b, e: e, error: err})
		}

		failToken := func(t Token, err error) {
			failAt(t.l, t.b, t.e, err)
		}

		failCurr := func(err error) {
			failToken(args.Curr(), err)
		}

		failEol := func(err error) {
			p := args.Prev()
			failAt(p.l, p.e, p.e, err)
		}

		failLine := func(err error) {
			diag = append(diag, lineError{l: li, error: err})
		}

		var e Op
		func() {
			if !args.Scan() {
				e = Empty
				return
			}

			if strings.HasSuffix(args.Text(), ":") {
				name := args.Text()
				name = name[:len(name)-1]
				if len(name) == 0 {
					failCurr(errors.New("missing label name"))
					return
				}
				e = &LabelExpr{Name: name}
				return
			}

			v := args.Text()

			mustArg := func(name string) bool {
				result := args.Scan()
				if !result {
					failEol(errors.Errorf("missing %s argument: %s", v, name))
				}
				return result
			}

			switch v {
			case "":
				e = Empty
			case "#pragma":
				if !mustArg("name") {
					return
				}

				switch args.Text() {
				case "version":
					if !mustArg("version value") {
						return
					}

					version, err := strconv.Atoi(args.Text())
					if err != nil {
						failLine(errors.Wrap(err, "failed to parse pragma version"))
						return
					}

					e = &PragmaExpr{Version: uint8(version)}
				default:
					failLine(errors.Errorf("unexpected #pragma: %s", args.Text()))
					return
				}
			case "bnz":
				if !mustArg("label name") {
					return
				}

				e = &BnzExpr{Label: &LabelExpr{Name: args.Text()}}
			case "bz":
				if !mustArg("label name") {
					return
				}

				e = &BzExpr{Label: &LabelExpr{Name: args.Text()}}
			case "b":
				if !mustArg("label name") {
					return
				}

				e = &BExpr{Label: &LabelExpr{Name: args.Text()}}
			case "bzero":
				e = Bzero
			case "getbyte":
				e = GetByte
			case "substring3":
				e = Substring3
			case "shr":
				e = Shr
			case "err":
				e = Err
			case "sha256":
				e = Sha256
			case "keccak256":
				e = Keccak256
			case "sha512_256":
				e = Sha512256
			case "ed25519verify":
				e = ED25519Verify
			case "ecdsa_verify":
				if !mustArg("curve index") {
					return
				}

				curve, err := readEcdsaCurveIndex(args.Text())
				if err != nil {
					failLine(errors.Wrap(err, "failed to read ecdsa_verify curve index"))
					return
				}
				e = &EcdsaVerifyExpr{Index: curve}
			case "ecdsa_pk_decompress":
				if !mustArg("curve index") {
					return
				}

				curve, err := readEcdsaCurveIndex(args.Text())
				if err != nil {
					failLine(errors.Wrap(err, "failed to read ecdsa_pk_decompress curve index"))
				}
				e = &EcdsaPkDecompressExpr{Index: curve}
			case "ecdsa_pk_recover":
				if !mustArg("curve index") {
					return
				}

				curve, err := readEcdsaCurveIndex(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read ecdsa_pk_recover curve index"))
					return
				}
				e = &EcdsaPkRecoverExpr{Index: curve}
			case "+":
				e = PlusOp
			case "-":
				e = MinusOp
			case "/":
				e = Div
			case "*":
				e = Mul
			case "<":
				e = Lt
			case ">":
				e = Gt
			case "<=":
				e = LtEq
			case ">=":
				e = GtEq
			case "&&":
				e = And
			case "||":
				e = Or
			case "==":
				e = Eq
			case "!=":
				e = Neq
			case "!":
				e = Not
			case "len":
				e = Len
			case "itob":
				e = Itob
			case "btoi":
				e = Btoi
			case "%":
				e = Mod
			case "|":
				e = BinOr
			case "&":
				e = BinAnd
			case "^":
				e = BinXor
			case "~":
				e = BinNot
			case "mulw":
				e = Mulw
			case "addw":
				e = Addw
			case "divmodw":
				e = DivModw
			case "divw":
				e = Divw
			case "select":
				e = Select
			case "b>=":
				e = Bgteq
			case "b<":
				e = Blt
			case "b&":
				e = Band
			case "b^":
				e = Bxor
			case "bsqrt":
				e = Bsqrt
			case "app_opted_in":
				e = AppOptedIn
			case "intcblock":
				var values []uint64

				for args.Scan() {
					value, err := readInt(args)
					if err != nil {
						failCurr(errors.Wrap(err, "failed to read int"))
						return
					}

					values = append(values, value)
				}

				e = &IntcBlockExpr{Values: values}
			case "intc":
				if !mustArg("value") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse intc value"))
					return
				}

				e = &IntcExpr{Index: uint8(value)}
			case "intc_0":
				e = Intc0
			case "intc_1":
				e = Intc1
			case "intc_2":
				e = Intc2
			case "intc_3":
				e = Intc3
			case "bytecblock":
				var values [][]byte

				for args.Scan() {
					b, err := readBytes(args)
					if err != nil {
						failCurr(errors.Wrap(err, "failed to read bytes"))
						return
					}

					values = append(values, b)
				}

				e = &BytecBlockExpr{Values: values}
			case "bytec":
				if !mustArg("index") {
					return
				}
				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse bytec index"))
					return
				}
				e = &BytecExpr{Index: uint8(value)}
			case "bytec_0":
				e = Bytec0
			case "bytec_1":
				e = Bytec1
			case "bytec_2":
				e = Bytec2
			case "bytec_3":
				e = Bytec3
			case "arg":
				if !mustArg("index") {
					return
				}
				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse arg index"))
					return
				}
				e = &ArgExpr{Index: uint8(value)}
			case "arg_0":
				e = Arg0
			case "arg_1":
				e = Arg1
			case "arg_2":
				e = Arg2
			case "arg_3":
				e = Arg3
			case "dup2":
				e = Dup2
			case "gitxna":
				if !mustArg("t") {
					return
				}

				t, err := strconv.ParseUint(args.Text(), 10, 64)
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gitxna t: %s", args.Text()))
					return
				}

				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gitxna f: %s", args.Text()))
					return
				}

				if !mustArg("i") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gitxna i"))
					return
				}

				e = &GitxnaExpr{Group: uint8(t), Field: f, Index: uint8(value)}
			case "gtxn":
				if !mustArg("t") {
					return
				}

				t, err := strconv.ParseUint(args.Text(), 10, 64)
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gtxn t: %s", args.Text()))
					return
				}

				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gtxn f: %s", args.Text()))
					return
				}

				e = &GtxnExpr{Index: uint8(t), Field: f}
			case "txn":
				if !mustArg("field") {
					return
				}

				field, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse txn field: %s", args.Text()))
					return
				}

				// TODO: check the field value
				e = &TxnExpr{Field: field}
			case "global":
				if !mustArg("field") {
					return
				}

				field, err := readGlobalField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse global field: %s", args.Text()))
					return
				}

				e = &GlobalExpr{Index: field}
			case "load":
				if !mustArg("index") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse load i"))
					return
				}

				e = &LoadExpr{Index: uint8(value)}
			case "gload":
				if !mustArg("t") {
					return
				}

				t, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gload t"))
					return
				}

				if !mustArg("i") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gload i"))
					return
				}

				e = &GloadExpr{Group: uint8(t), Index: uint8(value)}
			case "gloads":
				if !mustArg("i") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gloads i"))
					return
				}

				e = &GloadsExpr{Index: uint8(value)}
			case "store":
				if !mustArg("i") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse store i"))
					return
				}

				e = &StoreExpr{Index: uint8(value)}
			case "txna":
				if !mustArg("f") {
					return
				}

				field, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse txna field: %s", args.Text()))
					return
				}

				e = &TxnaExpr{Field: field}
			case "gtxns":
				if !mustArg("f") {
					return
				}

				field, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gtxns field: %s", args.Text()))
					return
				}

				e = &GtxnsExpr{Field: field}
			case "gaid":
				if !mustArg("t") {
					return
				}

				t, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gitxn t"))
					return
				}

				e = &GaidExpr{Group: uint8(t)}
			case "gtxna":
				if !mustArg("t") {
					return
				}

				t, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gitxn t"))
					return
				}

				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gtxnsa f: %s", args.Text()))
					return
				}

				if !mustArg("i") {
					return
				}

				i, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gtxnsa i"))
					return
				}

				e = &GtxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)}
			case "gtxnsa":
				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse gtxnsa f: %s", args.Text()))
					return
				}

				if !mustArg("i") {
					return
				}

				i, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gtxnsa i"))
					return
				}

				e = &GtxnsaExpr{Field: f, Index: uint8(i)}
			case "txnas":
				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse txnas f: %s", args.Text()))
					return
				}

				e = &TxnasExpr{Field: f}
			case "extract":
				if !mustArg("s") {
					return
				}
				start, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read extract s"))
					return
				}

				if !mustArg("l") {
					return
				}

				length, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read extract l"))
					return
				}

				e = &ExtractExpr{Start: uint8(start), Length: uint8(length)}
			case "substring":
				if !mustArg("s") {
					return
				}
				start, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read substract s"))
					return
				}

				if !mustArg("e") {
					return
				}

				end, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read substract e"))
					return
				}

				e = &SubstringExpr{Start: uint8(start), End: uint8(end)}
			case "proto":
				if !mustArg("a") {
					return
				}
				a, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read proto a"))
					return
				}

				if !mustArg("r") {
					return
				}

				r, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read proto r"))
					return
				}

				e = &ProtoExpr{Args: uint8(a), Results: uint8(r)}
			case "byte":
				if !mustArg("value") {
					return
				}

				value, err := readBytes(args)
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse byte value"))
					return
				}
				e = &ByteExpr{Value: value}
			case "pushbytes":
				if !mustArg("value") {
					return
				}

				value, err := readBytes(args)
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse pushbytes value"))
					return
				}
				e = &PushBytesExpr{Value: value}
			case "pushint":
				if !mustArg("value") {
					return
				}

				value, err := readInt(args)
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse pushint value"))
					return
				}
				e = &PushIntExpr{Value: value}
			case "asset_params_get":
				if !mustArg("f") {
					return
				}

				field, err := readAssetField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse asset_params_get f: %s", args.Text()))
					return
				}

				e = &AssetParamsGetExpr{Field: field}
			case "int":
				if !mustArg("value") {
					return
				}
				value, err := readInt(args)
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read int value"))
					return
				}
				e = &IntExpr{Value: value}
			case "sqrt":
				e = Sqrt
			case "box_del":
				e = BoxDel
			case "box_len":
				e = BoxLen
			case "box_create":
				e = BoxCreate
			case "box_get":
				e = BoxGet
			case "box_put":
				e = BoxPut
			case "box_replace":
				e = BoxReplace
			case "box_extract":
				e = BoxExtract
			case "pop":
				e = Pop
			case "swap":
				e = Swap
			case "app_global_put":
				e = AppGlobalPut
			case "app_local_put":
				e = AppLocalPut
			case "app_local_get":
				e = AppLocalGet
			case "app_global_get":
				e = AppGlobalGet
			case "app_global_get_ex":
				e = AppGlobalGetEx
			case "app_local_get_ex":
				e = AppLocalGetEx
			case "app_local_del":
				e = AppLocalDel
			case "app_global_del":
				e = AppGlobalDel
			case "itxn_next":
				e = ItxnNext
			case "min_balance":
				e = MinBalance
			case "getbit":
				e = GetBit
			case "setbit":
				e = SetBit
			case "b-":
				e = Bminus
			case "b*":
				e = Bmul
			case "b/":
				e = Bdiv
			case "b+":
				e = Bplus
			case "dig":
				if !mustArg("n") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read dig n"))
					return
				}

				e = &DigExpr{Index: uint8(value)}
			case "gtxnsas":
				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to read gtxnsas f: %s", args.Text()))
					return
				}

				e = &GtxnsasExpr{Field: f}
			case "gitxn":
				if !mustArg("t") {
					return
				}

				t, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse gitxn t"))
					return
				}

				if !mustArg("f") {
					return
				}

				f, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read gitxn f"))
					return
				}

				e = &GitxnExpr{Index: uint8(t), Field: f}
			case "asset_holding_get":
				if !mustArg("f") {
					return
				}

				f, err := readAssetHoldingField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to read asset_holding_get f: %s", args.Text()))
					return
				}

				e = &AssetHoldingGetExpr{Field: f}
			case "acct_params_get":
				if !mustArg("f") {
					return
				}

				f, err := readAcctParams(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse acct_params_get f: %s", args.Text()))
					return
				}

				e = &AcctParamsGetExpr{Field: f}

			case "app_params_get":
				if !mustArg("f") {
					return
				}

				f, err := readAppField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse app_params_get f: %s", args.Text()))
					return
				}

				e = &AppParamsGetExpr{Field: f}
			case "balance":
				e = Balance
			case "retsub":
				e = RetSub
			case "return":
				e = Return
			case "exp":
				e = Exp
			case "log":
				e = Log
			case "extract3":
				e = Extract3
			case "sha3_256":
				e = Sha3256
			case "extract_uint64":
				e = ExtractUint64
			case "vrf_verify":
				if !mustArg("f") {
					return
				}
				f, err := readVrfVerifyField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse vrf_verify f: %s", args.Text()))
					return
				}

				e = &VrfVerifyExpr{Field: f}
			case "block":
				if !mustArg("f") {
					return
				}

				f, err := readBlockField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse block f: %s", args.Text()))
					return
				}

				e = &BlockExpr{Field: f}
			case "switch":
				var labels []*LabelExpr
				for args.Scan() {
					labels = append(labels, &LabelExpr{Name: args.Text()})
				}
				e = &SwitchExpr{Targets: labels}
			case "match":
				var labels []*LabelExpr
				for args.Scan() {
					labels = append(labels, &LabelExpr{Name: args.Text()})
				}
				e = &MatchExpr{Targets: labels}
			case "callsub":
				if !mustArg("label name") {
					return
				}
				e = &CallSubExpr{Label: &LabelExpr{Name: args.Text()}}
			case "assert":
				e = Assert
			case "dup":
				e = Dup
			case "frame_bury":
				if !mustArg("index") {
					return
				}
				value, err := readInt8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse frame_bury i"))
					return
				}
				e = &FrameBuryExpr{Index: value}
			case "frame_dig":
				if !mustArg("index") {
					return
				}
				value, err := readInt8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse frame_dig i"))
					return
				}

				e = &FrameDigExpr{Index: value}
			case "setbyte":
				e = SetByte
			case "uncover":
				if !mustArg("index") {
					return
				}
				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to parse uncover i"))
					return
				}
				e = &UncoverExpr{Index: uint8(value)}
			case "cover":
				if !mustArg("n") {
					return
				}

				value, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrap(err, "failed to read cover n"))
					return
				}

				e = &CoverExpr{Index: uint8(value)}
			case "concat":
				e = Concat
			case "itxn_begin":
				e = ItxnBegin
			case "itxn_submit":
				e = ItxnSubmit
			case "itxn":
				if !mustArg("f") {
					return
				}

				field, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse itxn f: %s", args.Text()))
					return
				}

				e = &ItxnExpr{Field: field}
			case "itxn_field":
				if !mustArg("f") {
					return
				}

				field, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to read itxn_field f: %s", args.Text()))
					return
				}

				e = &ItxnFieldExpr{Field: field}
			default:
				failCurr(errors.Errorf("unexpected opcode: %s", args.Text()))
				return
			}
		}()

		if e != nil {
			res = append(res, e)
		}
	}

	l := &Linter{l: res}
	l.Lint()

	for _, le := range l.res {
		diag = append(diag, lintError{l: le.Line(), error: le})
	}

	syms := []Symbol{}

	for i, op := range res {
		switch op := op.(type) {
		case *LabelExpr:
			syms = append(syms, labelSymbol{
				n: op.Name,
				l: i,
			})
		}
	}

	result := ProcessResult{
		Diagnostics: diag,
		Symbols:     syms,
	}

	return result
}
