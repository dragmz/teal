package teal

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type parseError struct {
	error

	l int
	b int
	e int
}

func (e parseError) Line() int {
	return e.l
}

func (e parseError) Begin() int {
	return e.b
}

func (e parseError) End() int {
	return e.e
}

func (e parseError) String() string {
	return e.error.Error()
}

func (e parseError) Severity() DiagnosticSeverity {
	return DiagErr
}

type lintError struct {
	error
	l int
}

func (e lintError) Line() int {
	return e.l
}

func (e lintError) Begin() int {
	return 0
}

func (e lintError) End() int {
	return 0
}

func (e lintError) String() string {
	return e.error.Error()
}

func (e lintError) Severity() DiagnosticSeverity {
	return DiagWarn
}

type lineError struct {
	error
	l int
}

func (e lineError) Line() int {
	return e.l
}

func (e lineError) Begin() int {
	return 0
}

func (e lineError) End() int {
	return 0
}

func (e lineError) String() string {
	return e.error.Error()
}

func (e lineError) Severity() DiagnosticSeverity {
	return DiagErr
}

func readInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse int8: %s", s)
	}

	return int8(v), nil
}

func readUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse uint8: %s", s)
	}

	return uint8(v), nil
}

func readAssetHoldingField(s string) (AssetHoldingField, error) {
	switch s {
	case "AssetBalance":
		return AssetBalance, nil
	case "AssetFrozen":
		return AssetFrozen, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse asset holding field")
		}
		return AssetHoldingField(value), nil
	}
}

func readVrfVerifyField(s string) (VrfVerifyField, error) {
	switch s {
	case "VrfAlgorand":
		return VrfAlgorand, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse vrf_verify f")
		}

		return VrfVerifyField(value), nil

	}
}

func readBlockField(s string) (BlockField, error) {
	switch s {
	case "BlkSeed":
		return BlkSeed, nil
	case "BlkTimestamp":
		return BlkTimestamp, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse block f")
		}

		return BlockField(value), nil
	}
}

func readAcctParams(s string) (AcctParamsField, error) {
	switch s {
	case "AcctBalance":
		return AcctBalance, nil
	case "AcctMinBalance":
		return AcctMinBalance, nil
	case "AcctAuthAddr":
		return AcctAuthAddr, nil
	case "AcctTotalNumUint":
		return AcctTotalNumUint, nil
	case "AcctTotalNumByteSlice":
		return AcctTotalNumByteSlice, nil
	case "AcctTotalExtraAppPages":
		return AcctTotalExtraAppPages, nil
	case "AcctTotalAppsCreated":
		return AcctTotalAppsCreated, nil
	case "AcctTotalAppsOptedIn":
		return AcctTotalAppsOptedIn, nil
	case "AcctTotalAssetsCreated":
		return AcctTotalAssetsCreated, nil
	case "AcctTotalAssets":
		return AcctTotalAssets, nil
	case "AcctTotalBoxes":
		return AcctTotalBoxes, nil
	case "AcctTotalBoxBytes":
		return AcctTotalBoxBytes, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse app params field")
		}

		return AcctParamsField(value), nil
	}
}

func readAppField(s string) (AppParamsField, error) {
	switch s {
	case "AppApprovalProgram":
		return AppApprovalProgram, nil
	case "AppClearStateProgram":
		return AppClearStateProgram, nil
	case "AppGlobalNumUint":
		return AppGlobalNumUint, nil
	case "AppGlobalNumByteSlice":
		return AppGlobalNumByteSlice, nil
	case "AppLocalNumUint":
		return AppLocalNumUint, nil
	case "AppLocalNumByteSlice":
		return AppLocalNumByteSlice, nil
	case "AppExtraProgramPages":
		return AppExtraProgramPages, nil
	case "AppCreator":
		return AppCreator, nil
	case "AppAddress":
		return AppAddress, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse app params field")
		}

		return AppParamsField(value), nil
	}
}

func readAssetField(s string) (AssetParamsField, error) {
	switch s {
	case "AssetTotal":
		return AssetTotal, nil
	case "AssetDecimals":
		return AssetDecimals, nil
	case "AssetDefaultFrozen":
		return AssetDefaultFrozen, nil
	case "AssetUnitName":
		return AssetUnitName, nil
	case "AssetName":
		return AssetName, nil
	case "AssetURL":
		return AssetURL, nil
	case "AssetMetadataHash":
		return AssetMetadataHash, nil
	case "AssetManager":
		return AssetManager, nil
	case "AssetReserve":
		return AssetReserve, nil
	case "AssetFreeze":
		return AssetFreeze, nil
	case "AssetClawback":
		return AssetClawback, nil
	case "AssetCreator":
		return AssetCreator, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse asset params field")
		}

		return AssetParamsField(value), nil
	}
}

func readGlobalField(s string) (GlobalField, error) {
	switch s {
	case "MinTxnFee":
		return GlobalMinTxnFee, nil
	case "MinBalance":
		return GlobalMinBalance, nil
	case "lMaxTxnLife":
		return GlobalMaxTxnLife, nil
	case "ZeroAddress":
		return GlobalZeroAddress, nil
	case "GroupSize":
		return GlobalGroupSize, nil
	case "LogicSigVersion":
		return GlobalLogicSigVersion, nil
	case "Round":
		return GlobalRound, nil
	case "LatestTimestamp":
		return GlobalLatestTimestamp, nil
	case "CurrentApplicationID":
		return GlobalCurrentApplicationID, nil
	case "CreatorAddress":
		return GlobalCreatorAddress, nil
	case "CurrentApplicationAddress":
		return GlobalCurrentApplicationAddress, nil
	case "GroupID":
		return GlobalGroupID, nil
	case "OpcodeBudget":
		return GlobalOpcodeBudget, nil
	case "CallerApplicationID":
		return GlobalCallerApplicationID, nil
	case "CallerApplicationAddress":
		return GlobalCallerApplicationAddress, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse global field")
		}

		return GlobalField(value), nil
	}
}

func readTxnField(s string) (TxnField, error) {
	switch s {
	case "Sender":
		return TxnSender, nil
	case "Fee":
		return TxnFee, nil
	case "FirstValid":
		return TxnFirstValid, nil
	case "FirstValidTime":
		return TxnFirstValidTime, nil
	case "LastValid":
		return TxnLastValid, nil
	case "Note":
		return TxnNote, nil
	case "Lease":
		return TxnLease, nil
	case "Receiver":
		return TxnReceiver, nil
	case "Amount":
		return TxnAmount, nil
	case "CloseRemainderTo":
		return TxnCloseRemainderTo, nil
	case "VotePK":
		return TxnVotePK, nil
	case "SelectionPK":
		return TxnSelectionPK, nil
	case "VoteFirst":
		return TxnVoteFirst, nil
	case "VoteLast":
		return TxnVoteLast, nil
	case "VoteKeyDilution":
		return TxnVoteKeyDilution, nil
	case "Type":
		return TxnType, nil
	case "TypeEnum":
		return TxnTypeEnum, nil
	case "XferAsset":
		return TxnXferAsset, nil
	case "AssetAmount":
		return TxnAssetAmount, nil
	case "AssetSender":
		return TxnAssetSender, nil
	case "AssetReceiver":
		return TxnAssetReceiver, nil
	case "AssetCloseTo":
		return TxnAssetCloseTo, nil
	case "GroupIndex":
		return TxnGroupIndex, nil
	case "TxID":
		return TxnTxID, nil
	case "ApplicationID":
		return TxnApplicationID, nil
	case "OnCompletion":
		return TxnOnCompletion, nil
	case "ApplicationArgs":
		return TxnApplicationArgs, nil
	case "NumAppArgs":
		return TxnNumAppArgs, nil
	case "Accounts":
		return TxnAccounts, nil
	case "NumAccounts":
		return TxnNumAccounts, nil
	case "ApprovalProgram":
		return TxnApprovalProgram, nil
	case "ClearStateProgram":
		return TxnClearStateProgram, nil
	case "RekeyTo":
		return TxnRekeyTo, nil
	case "ConfigAsset":
		return TxnConfigAsset, nil
	case "ConfigAssetTotal":
		return TxnConfigAssetTotal, nil
	case "ConfigAssetDecimals":
		return TxnConfigAssetDecimals, nil
	case "ConfigAssetDefaultFrozen":
		return TxnConfigAssetDefaultFrozen, nil
	case "ConfigAssetUnitName":
		return TxnConfigAssetUnitName, nil
	case "ConfigAssetName":
		return TxnConfigAssetName, nil
	case "ConfigAssetURL":
		return TxnConfigAssetURL, nil
	case "ConfigAssetMetadataHash":
		return TxnConfigAssetMetadataHash, nil
	case "ConfigAssetManager":
		return TxnConfigAssetManager, nil
	case "ConfigAssetReserve":
		return TxnConfigAssetReserve, nil
	case "ConfigAssetFreeze":
		return TxnConfigAssetFreeze, nil
	case "ConfigAssetClawback":
		return TxnConfigAssetClawback, nil
	case "FreezeAsset":
		return TxnFreezeAsset, nil
	case "FreezeAssetAccounts":
		return TxnFreezeAssetAccounts, nil
	case "FreezeAssetFrozen":
		return TxnFreezeAssetFrozen, nil
	case "Assets":
		return TxnAssets, nil
	case "NumAssets":
		return TxnNumAssets, nil
	case "Applications":
		return TxnApplications, nil
	case "NumApplications":
		return TxnNumApplications, nil
	case "GlobalNumUint":
		return TxnGlobalNumUint, nil
	case "GlobalNumByteSlice":
		return TxnGlobalNumByteSlice, nil
	case "LocalNumUint":
		return TxnLocalNumUint, nil
	case "LocalNumByteSlice":
		return TxnLocalNumByteSlice, nil
	case "ExtraProgramPages":
		return TxnExtraProgramPages, nil
	case "Nonparticipation":
		return TxnNonparticipation, nil
	case "Logs":
		return TxnLogs, nil
	case "NumLogs":
		return TxnNumLogs, nil
	case "CreatedAssetID":
		return TxnCreatedAssetID, nil
	case "CreatedApplicationID":
		return TxnCreatedApplicationID, nil
	case "LastLog":
		return TxnLastLog, nil
	case "StateProofPK":
		return TxnStateProofPK, nil
	case "ApprovalProgramPages":
		return TxnApprovalProgramPages, nil
	case "NumApprovalProgramPages":
		return TxnNumApprovalProgramPages, nil
	case "ClearStateProgramPages":
		return TxnClearStateProgramPages, nil
	case "NumClearStateProgramPages":
		return TxnNumClearStateProgramPages, nil
	default:
		value, err := readUint8(s)
		if err != nil {
			return 0, errors.Wrap(err, "failed to parse txn field")
		}

		return TxnField(value), nil
	}
}

func readHexInt(s string) (uint64, error) {
	if !strings.HasPrefix(s, "0x") {
		return 0, errors.Errorf("unexpected hex int: %s", s)
	}

	const maxLen = 16
	const minLen = 2

	l := len(s[2:])
	if l < minLen {
		return 0, errors.Errorf("hex int too short - got: %d, min: %d", l, minLen)
	}
	if l > maxLen {
		return 0, errors.Errorf("hex int too long - got: %d, max: %d", l, maxLen)
	}

	var dst [8]byte
	_, err := hex.Decode(dst[:], []byte(s[2:]))
	if err != nil {
		return 0, errors.Wrap(err, "failed to decode hex int")
	}

	v := binary.LittleEndian.Uint64(dst[:])

	return v, nil
}

func readInt(a *arguments) (uint64, error) {
	switch a.Text() {
	case "pay":
		return 1, nil
	case "keyreg":
		return 2, nil
	case "acfg":
		return 3, nil
	case "axfer":
		return 4, nil
	case "afrz":
		return 5, nil
	case "appl":
		return 6, nil
	}

	val, err := strconv.ParseUint(a.Text(), 0, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse uint64")
	}

	return val, nil
}

func readBytes(a *arguments) ([]byte, error) {
	args := []string{a.Text()}
	for a.Scan() {
		args = append(args, a.Text())
	}

	val, _, err := parseBinaryArgs(args)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse binary args")
	}

	return val, nil
}

func readEcdsaCurveIndex(s string) (EcdsaCurve, error) {
	var curve EcdsaCurve

	value, err := strconv.Atoi(s)
	if err != nil {
		return curve, errors.Wrap(err, "failed to read ecdsa_verify curve index")
	}

	curve = EcdsaCurve(value)
	switch curve {
	case EcdsaVerifySecp256k1:
	case EcdsaVerifySecp256r1:
	default:
		return curve, errors.Errorf("unexpected ecdsa_verify curve index: %d", value)
	}

	return curve, nil
}

type arguments struct {
	ts []Token
	i  int
}

func (a *arguments) Scan() bool {
	if a.i <= len(a.ts) {
		a.i++
		return a.i <= len(a.ts)
	}

	return false
}

func (a *arguments) Prev() Token {
	if a.i > 1 {
		return a.ts[a.i-2]
	}

	return Token{}
}

func (a *arguments) Curr() Token {
	if a.i > 0 && a.i <= len(a.ts) {
		return a.ts[a.i-1]
	}

	return Token{}
}

func (a *arguments) Text() string {
	return a.Curr().String()
}

func readTokens(source string) ([]Token, []Diagnostic) {
	s := &Lexer{Source: []byte(source)}

	ts := []Token{}

	for s.Scan() {
		ts = append(ts, s.Curr())
	}

	diags := make([]Diagnostic, len(s.diag))

	for i, diag := range s.diag {
		diags[i] = diag
	}

	return ts, diags
}

func Lint(source string) []Diagnostic {
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

		if !args.Scan() {
			// handle empty line
			continue
		}

		var e Op
		func() {
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

			if strings.HasPrefix(args.Text(), "#") {
				if !args.Scan() {
					failEol(errors.New("missing pragma name"))
					return
				}

				switch args.Text() {
				case "version":
					if !args.Scan() {
						failEol(errors.New("missing pragma version value"))
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

				return
			}

			v := args.Text()

			mustArg := func(name string) bool {
				result := args.Scan()
				if !result {
					failEol(errors.Errorf("missing %s argument %s", v, name))
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

	return diag
}
