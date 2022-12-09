package teal

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var PseudoOps = pseudoOps
var OpDocByName = opDocByName
var OpDocExtras = opDocExtras

type parserContext struct{}

type parserFunc func(c *parserContext) error
type protoFunc func(p string)

type OpSpecProto struct{}
type OpSpecDetails struct {
	NamesMap map[string]bool
	Names    []string
}

type OpSpec struct {
	Code      byte
	Name      string
	Parse     parserFunc
	Proto     OpSpecProto
	Version   uint8
	OpDetails OpSpecDetails
}

// TODO
func assembler(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func costly(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

// TODO
func detDefault() OpSpecDetails {
	return OpSpecDetails{}
}

func detSwitch() OpSpecDetails {
	return OpSpecDetails{}
}

// TODO

func costByField(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func typed(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

// TODO
const (
	modeSig = 1
)

// TODO
func (d OpSpecDetails) only(vs ...interface{}) OpSpecDetails {
	return d
}

func (d OpSpecDetails) assembler(vs ...interface{}) OpSpecDetails {
	return d
}

func only(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

var asmInt = []interface{}{}
var asmByte = []interface{}{}
var asmIntC = []interface{}{}
var asmArg = []interface{}{}
var asmAddr = []interface{}{}
var asmMethod = []interface{}{}

var typeLoads = []interface{}{}
var typeStores = []interface{}{}

var asmByteCBlock = []interface{}{}
var checkByteImmArgs = []interface{}{}
var immBytess = []interface{}{}
var typeTxField = []interface{}{}
var TxnFields = []interface{}{}

func immediates(names ...string) OpSpecDetails {
	m := map[string]bool{}
	r := []string{}

	for _, name := range names {
		if !m[name] {
			m[name] = true
			r = append(r, name)
		}
	}
	return OpSpecDetails{
		NamesMap: m,
		Names:    r,
	}
}

func field(name string, vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{
		NamesMap: map[string]bool{name: true},
		Names:    []string{name},
	}
}

func detBranch(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

var typeBury = []interface{}{}

func (d OpSpecDetails) costs(vs ...interface{}) OpSpecDetails {
	return d
}

func (d OpSpecDetails) field(name string, vs ...interface{}) OpSpecDetails {
	if !d.NamesMap[name] {
		d.NamesMap[name] = true
		d.Names = append(d.Names, name)
	}

	return d
}

func (d OpSpecDetails) typed(vs ...interface{}) OpSpecDetails {
	return d
}

func (d OpSpecDetails) trust(vs ...interface{}) OpSpecDetails {
	return d
}

func (d OpSpecDetails) costByLength(vs ...interface{}) OpSpecDetails {
	return d
}

var typeLoad = []interface{}{}
var typeStore = []interface{}{}

var asmByteC = []interface{}{}
var asmItxn = []interface{}{}
var asmGitxn = []interface{}{}
var TxnScalarFields = []interface{}{}
var GlobalFields = []interface{}{}
var TxnArrayFields = []interface{}{}
var modeApp = []interface{}{}
var AssetHoldingFields = []interface{}{}
var AssetParamsFields = []interface{}{}
var AppParamsFields = []interface{}{}
var AcctParamsFields = []interface{}{}

var asmPushBytes = []interface{}{}
var immBytes = []interface{}{}
var asmPushInt = []interface{}{}
var immInt = []interface{}{}
var asmPushBytess = []interface{}{}
var asmPushInts = []interface{}{}
var VrfStandards = []interface{}{}
var BlockFields = []interface{}{}

// TODO
func proto(vs ...interface{}) OpSpecProto {
	return OpSpecProto{}
}

// TODO
func opErr(c *parserContext) error {
	return nil
}

func opSHA256(c *parserContext) error {
	return nil
}

func opKeccak256(c *parserContext) error {
	return nil
}
func opSHA512_256(c *parserContext) error {
	return nil
}
func opEd25519Verify(c *parserContext) error {
	return nil
}
func opEcdsaVerify(c *parserContext) error {
	return nil
}
func opEcdsaPkDecompress(c *parserContext) error {
	return nil
}
func opEcdsaPkRecover(c *parserContext) error {
	return nil
}
func opPlus(c *parserContext) error {
	return nil
}
func opMinus(c *parserContext) error {
	return nil
}
func opDiv(c *parserContext) error {
	return nil
}
func opMul(c *parserContext) error {
	return nil
}
func opLt(c *parserContext) error {
	return nil
}
func opGt(c *parserContext) error {
	return nil
}
func opLe(c *parserContext) error {
	return nil
}
func opGe(c *parserContext) error {
	return nil
}
func opAnd(c *parserContext) error {
	return nil
}
func opOr(c *parserContext) error {
	return nil
}
func opEq(c *parserContext) error {
	return nil
}
func opNeq(c *parserContext) error {
	return nil
}
func opNot(c *parserContext) error {
	return nil
}
func opLen(c *parserContext) error {
	return nil
}
func opItob(c *parserContext) error {
	return nil
}
func opBtoi(c *parserContext) error {
	return nil
}
func opModulo(c *parserContext) error {
	return nil
}
func opBitOr(c *parserContext) error {
	return nil
}
func opBitAnd(c *parserContext) error {
	return nil
}
func opBitXor(c *parserContext) error {
	return nil
}
func opBitNot(c *parserContext) error {
	return nil
}
func opMulw(c *parserContext) error {
	return nil
}
func opAddw(c *parserContext) error {
	return nil
}
func opDivModw(c *parserContext) error {
	return nil
}
func opIntConstBlock(c *parserContext) error {
	return nil
}
func opIntConstLoad(c *parserContext) error {
	return nil
}
func opIntConst0(c *parserContext) error {
	return nil
}
func opIntConst1(c *parserContext) error {
	return nil
}
func opIntConst2(c *parserContext) error {
	return nil
}
func opIntConst3(c *parserContext) error {
	return nil
}
func opByteConstBlock(c *parserContext) error {
	return nil
}
func opByteConstLoad(c *parserContext) error {
	return nil
}
func opByteConst0(c *parserContext) error {
	return nil
}
func opByteConst1(c *parserContext) error {
	return nil
}
func opByteConst2(c *parserContext) error {
	return nil
}
func opByteConst3(c *parserContext) error {
	return nil
}
func opArg(c *parserContext) error {
	return nil
}
func opArg0(c *parserContext) error {
	return nil
}
func opArg1(c *parserContext) error {
	return nil
}
func opArg2(c *parserContext) error {
	return nil
}
func opArg3(c *parserContext) error {
	return nil
}
func opTxn(c *parserContext) error {
	return nil
}
func opGlobal(c *parserContext) error {
	return nil
}
func opGtxn(c *parserContext) error {
	return nil
}
func opLoad(c *parserContext) error {
	return nil
}
func opStore(c *parserContext) error {
	return nil
}
func opTxna(c *parserContext) error {
	return nil
}
func opGtxna(c *parserContext) error {
	return nil
}
func opGtxns(c *parserContext) error {
	return nil
}
func opGtxnsa(c *parserContext) error {
	return nil
}
func opGload(c *parserContext) error {
	return nil
}
func opGloads(c *parserContext) error {
	return nil
}
func opGaid(c *parserContext) error {
	return nil
}
func opGaids(c *parserContext) error {
	return nil
}
func opLoads(c *parserContext) error {
	return nil
}
func opStores(c *parserContext) error {
	return nil
}
func opBnz(c *parserContext) error {
	return nil
}
func opBz(c *parserContext) error {
	return nil
}
func opB(c *parserContext) error {
	return nil
}
func opReturn(c *parserContext) error {
	return nil
}
func opAssert(c *parserContext) error {
	return nil
}
func opBury(c *parserContext) error {
	return nil
}
func opPopN(c *parserContext) error {
	return nil
}
func opDupN(c *parserContext) error {
	return nil
}

func opPop(c *parserContext) error {
	return nil
}
func opDup(c *parserContext) error {
	return nil
}
func opDup2(c *parserContext) error {
	return nil
}
func opDig(c *parserContext) error {
	return nil
}
func opSwap(c *parserContext) error {
	return nil
}
func opSelect(c *parserContext) error {
	return nil
}
func opCover(c *parserContext) error {
	return nil
}
func opUncover(c *parserContext) error {
	return nil
}
func opConcat(c *parserContext) error {
	return nil
}
func opSubstring(c *parserContext) error {
	return nil
}
func opSubstring3(c *parserContext) error {
	return nil
}
func opGetBit(c *parserContext) error {
	return nil
}
func opSetBit(c *parserContext) error {
	return nil
}
func opGetByte(c *parserContext) error {
	return nil
}
func opSetByte(c *parserContext) error {
	return nil
}
func opExtract(c *parserContext) error {
	return nil
}
func opExtract3(c *parserContext) error {
	return nil
}
func opExtract16Bits(c *parserContext) error {
	return nil
}
func opExtract32Bits(c *parserContext) error {
	return nil
}
func opExtract64Bits(c *parserContext) error {
	return nil
}
func opReplace2(c *parserContext) error {
	return nil
}
func opReplace3(c *parserContext) error {
	return nil
}
func opBase64Decode(c *parserContext) error {
	return nil
}
func opJSONRef(c *parserContext) error {
	return nil
}
func opBalance(c *parserContext) error {
	return nil
}
func opAppOptedIn(c *parserContext) error {
	return nil
}
func opAppLocalGet(c *parserContext) error {
	return nil
}
func opAppLocalGetEx(c *parserContext) error {
	return nil
}
func opAppGlobalGet(c *parserContext) error {
	return nil
}
func opAppGlobalGetEx(c *parserContext) error {
	return nil
}
func opAppLocalPut(c *parserContext) error {
	return nil
}
func opAppGlobalPut(c *parserContext) error {
	return nil
}
func opAppLocalDel(c *parserContext) error {
	return nil
}
func opAppGlobalDel(c *parserContext) error {
	return nil
}
func opAssetHoldingGet(c *parserContext) error {
	return nil
}
func opAssetParamsGet(c *parserContext) error {
	return nil
}
func opAppParamsGet(c *parserContext) error {
	return nil
}
func opAcctParamsGet(c *parserContext) error {
	return nil
}
func opMinBalance(c *parserContext) error {
	return nil
}
func opPushBytes(c *parserContext) error {
	return nil
}
func opPushInt(c *parserContext) error {
	return nil
}
func opPushBytess(c *parserContext) error {
	return nil
}
func opEd25519VerifyBare(c *parserContext) error {
	return nil
}
func opPushInts(c *parserContext) error {
	return nil
}
func opCallSub(c *parserContext) error {
	return nil
}
func opRetSub(c *parserContext) error {
	return nil
}
func opProto(c *parserContext) error {
	return nil
}
func opFrameDig(c *parserContext) error {
	return nil
}
func opFrameBury(c *parserContext) error {
	return nil
}
func opSwitch(c *parserContext) error {
	return nil
}
func opMatch(c *parserContext) error {
	return nil
}
func opShiftLeft(c *parserContext) error {
	return nil
}
func opShiftRight(c *parserContext) error {
	return nil
}
func opSqrt(c *parserContext) error {
	return nil
}
func opBitLen(c *parserContext) error {
	return nil
}
func opExp(c *parserContext) error {
	return nil
}
func opExpw(c *parserContext) error {
	return nil
}
func opBytesSqrt(c *parserContext) error {
	return nil
}
func opDivw(c *parserContext) error {
	return nil
}
func opSHA3_256(c *parserContext) error {
	return nil
}
func opBn256Add(c *parserContext) error {
	return nil
}
func opBn256ScalarMul(c *parserContext) error {
	return nil
}
func opBn256Pairing(c *parserContext) error {
	return nil
}
func opBytesPlus(c *parserContext) error {
	return nil
}
func opBytesMinus(c *parserContext) error {
	return nil
}
func opBytesDiv(c *parserContext) error {
	return nil
}
func opBytesMul(c *parserContext) error {
	return nil
}
func opBytesLt(c *parserContext) error {
	return nil
}
func opBytesGt(c *parserContext) error {
	return nil
}
func opBytesLe(c *parserContext) error {
	return nil
}
func opBytesGe(c *parserContext) error {
	return nil
}
func opBytesEq(c *parserContext) error {
	return nil
}
func opBytesNeq(c *parserContext) error {
	return nil
}
func opBytesModulo(c *parserContext) error {
	return nil
}
func opBytesBitOr(c *parserContext) error {
	return nil
}
func opBytesBitAnd(c *parserContext) error {
	return nil
}
func opBytesBitXor(c *parserContext) error {
	return nil
}
func opBytesBitNot(c *parserContext) error {
	return nil
}
func opBytesZero(c *parserContext) error {
	return nil
}
func opLog(c *parserContext) error {
	return nil
}
func opTxBegin(c *parserContext) error {
	return nil
}
func opItxnField(c *parserContext) error {
	return nil
}
func opItxnSubmit(c *parserContext) error {
	return nil
}
func opItxn(c *parserContext) error {
	return nil
}
func opItxna(c *parserContext) error {
	return nil
}
func opItxnNext(c *parserContext) error {
	return nil
}
func opGitxn(c *parserContext) error {
	return nil
}
func opGitxna(c *parserContext) error {
	return nil
}
func opBoxCreate(c *parserContext) error {
	return nil
}
func opBoxExtract(c *parserContext) error {
	return nil
}
func opBoxReplace(c *parserContext) error {
	return nil
}
func opBoxDel(c *parserContext) error {
	return nil
}
func opBoxLen(c *parserContext) error {
	return nil
}
func opBoxGet(c *parserContext) error {
	return nil
}
func opBoxPut(c *parserContext) error {
	return nil
}
func opTxnas(c *parserContext) error {
	return nil
}
func opGtxnas(c *parserContext) error {
	return nil
}
func opGtxnsas(c *parserContext) error {
	return nil
}
func opArgs(c *parserContext) error {
	return nil
}
func opGloadss(c *parserContext) error {
	return nil
}
func opItxnas(c *parserContext) error {
	return nil
}
func opGitxnas(c *parserContext) error {
	return nil
}
func opVrfVerify(c *parserContext) error {
	return nil
}
func opBlock(c *parserContext) error {
	return nil
}

var EcdsaCurves = []interface{}{}
var ecdsaVerifyCosts = []interface{}{}
var ecdsaDecompressCosts = []interface{}{}
var typeEquals = []interface{}{}
var typePopN = []interface{}{}
var typeDupN = []interface{}{}
var typeDup = []interface{}{}
var typeDupTwo = []interface{}{}
var typeDig = []interface{}{}
var typeSwap = []interface{}{}
var typeSelect = []interface{}{}
var typeCover = []interface{}{}
var typeUncover = []interface{}{}
var asmSubstring = []interface{}{}
var typeSetBit = []interface{}{}
var Base64Encodings = []interface{}{}
var JSONRefTypes = []interface{}{}
var typeProto = []interface{}{}
var immInt8 = []interface{}{}
var typePushBytess = []interface{}{}
var typePushInts = []interface{}{}
var typeFrameDig = []interface{}{}
var typeFrameBury = []interface{}{}
var asmItxnField = []interface{}{}

func constants(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func immKinded(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

var asmIntCBlock = []interface{}{}
var checkIntImmArgs = []interface{}{}
var immInts = []interface{}{}

type ProcessResult struct {
	Version     uint8
	Diagnostics []Diagnostic
	Symbols     []Symbol
	SymbolRefs  []Symbol
	Tokens      []Token
	Listing     Listing
	Lines       [][]Token
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

func Process(source string) *ProcessResult {
	type recoverable struct{}

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
			t := l[i]
			if t.Type() == TokenComment {
				lines[li] = l[:i]
			}
		}
	}

	var lts [][]Token
	var res Listing
	version := uint8(1)

	for _, l := range lines {
		args := &arguments{ts: l}

		failAt := func(l int, b int, e int, err error) {
			diag = append(diag, parseError{l: l, b: b, e: e, error: err})
			panic(recoverable{})
		}

		failToken := func(t Token, err error) {
			failAt(t.l, t.b, t.e, err)
		}

		failCurr := func(err error) {
			failToken(args.Curr(), err)
		}

		failPrev := func(err error) {
			failToken(args.Prev(), err)
		}

		var e Op
		func() {
			defer func() {
				switch v := recover().(type) {
				case recoverable:
					e = Empty // consider replacing with Raw string expr
				case nil:
				default:
					fmt.Printf("unrecoverable: %v", v)
					panic(v)
				}
			}()

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

			mustReadArg := func(name string) {
				if !args.Scan() {
					failPrev(errors.Errorf("missing arg: %s", name))
				}
			}

			parseUint64 := func(name string) uint64 {
				v, err := readInt(args)
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse uint64: %s", name))
				}

				return v
			}

			parseUint8 := func(name string) uint8 {
				v, err := readUint8(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse uint8: %s", name))
				}
				return v
			}

			parseInt8 := func(name string) int8 {
				v, err := readInt8(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse int8: %s", name))
				}
				return v
			}

			parseTxnField := func(name string) TxnField {
				v, err := readTxnField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse txn field: %s", name))
				}
				return v
			}

			parseBytes := func(name string) []byte {
				v, err := readBytes(args)
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse bytes: %s", name))
				}
				return v
			}

			parseEcdsaCurveIndex := func(name string) EcdsaCurve {
				v, err := readEcdsaCurveIndex(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse ESCDS curve index: %s", name))
				}
				return v
			}

			mustReadBytes := func(name string) []byte {
				mustReadArg(name)
				return parseBytes(name)
			}

			mustReadInt := func(name string) uint64 {
				mustReadArg(name)
				return parseUint64(name)
			}

			mustReadUint8 := func(name string) uint8 {
				mustReadArg(name)
				return parseUint8(name)
			}

			mustReadInt8 := func(name string) int8 {
				mustReadArg(name)
				return parseInt8(name)
			}

			mustRead := func(name string) string {
				mustReadArg(name)
				return args.Text()
			}

			mustReadTxnField := func(name string) TxnField {
				mustReadArg(name)
				return parseTxnField(name)
			}

			mustReadEcdsaCurveIndex := func(name string) EcdsaCurve {
				mustReadArg(name)
				return parseEcdsaCurveIndex(name)
			}

			switch v {
			case "":
				e = Empty
			case "#pragma":
				name := mustRead("name")
				switch name {
				case "version":
					version = mustReadUint8("version value")
					e = &PragmaExpr{Version: uint8(version)}
				default:
					failCurr(errors.Errorf("unexpected #pragma: %s", args.Text()))
					return
				}
			case "bnz":
				name := mustRead("label name")
				e = &BnzExpr{Label: &LabelExpr{Name: name}}
			case "bz":
				name := mustRead("label name")
				e = &BzExpr{Label: &LabelExpr{Name: name}}
			case "b":
				name := mustRead("label name")
				e = &BExpr{Label: &LabelExpr{Name: name}}
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
				curve := mustReadEcdsaCurveIndex("curve index")
				e = &EcdsaVerifyExpr{Index: curve}
			case "ecdsa_pk_decompress":
				curve := mustReadEcdsaCurveIndex("curve index")
				e = &EcdsaPkDecompressExpr{Index: curve}
			case "ecdsa_pk_recover":
				curve := mustReadEcdsaCurveIndex("curve index")
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
					value := parseUint64("value")
					values = append(values, value)
				}

				e = &IntcBlockExpr{Values: values}
			case "intc":
				value := mustReadUint8("value")
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
					b := parseBytes("value")
					values = append(values, b)
				}

				e = &BytecBlockExpr{Values: values}
			case "bytec":
				value := mustReadUint8("index")
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
				value := mustReadUint8("index")
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
				t := mustReadInt("t")
				f := mustReadTxnField("f")
				i := mustReadUint8("i")

				e = &GitxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)}
			case "gtxn":
				t := mustReadInt("t")
				f := mustReadTxnField("f")
				e = &GtxnExpr{Index: uint8(t), Field: f}
			case "txn":
				f := mustReadTxnField("f")
				e = &TxnExpr{Field: f}
			case "global":
				mustReadArg("field")

				field, err := readGlobalField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse global field: %s", args.Text()))
					return
				}

				e = &GlobalExpr{Index: field}
			case "load":
				value := mustReadUint8("i")
				e = &LoadExpr{Index: uint8(value)}
			case "gload":
				t := mustReadUint8("t")
				value := mustReadUint8("i")

				e = &GloadExpr{Group: uint8(t), Index: uint8(value)}
			case "gloads":
				value := mustReadUint8("i")
				e = &GloadsExpr{Index: uint8(value)}
			case "store":
				value := mustReadUint8("i")
				e = &StoreExpr{Index: uint8(value)}
			case "txna":
				f := mustReadTxnField("f")
				i := mustReadUint8("i")
				e = &TxnaExpr{Field: f, Index: i}
			case "gtxns":
				f := mustReadTxnField("f")
				e = &GtxnsExpr{Field: f}
			case "gaid":
				t := mustReadUint8("t")
				e = &GaidExpr{Group: uint8(t)}
			case "gtxna":
				t := mustReadUint8("t")
				f := mustReadTxnField("f")
				i := mustReadUint8("i")
				e = &GtxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)}
			case "gtxnsa":
				f := mustReadTxnField("f")
				i := mustReadUint8("i")
				e = &GtxnsaExpr{Field: f, Index: uint8(i)}
			case "txnas":
				f := mustReadTxnField("f")
				e = &TxnasExpr{Field: f}
			case "extract":
				start := mustReadUint8("s")
				length := mustReadUint8("l")

				e = &ExtractExpr{Start: uint8(start), Length: uint8(length)}
			case "substring":
				start := mustReadUint8("s")
				end := mustReadUint8("e")
				e = &SubstringExpr{Start: uint8(start), End: uint8(end)}
			case "proto":
				a := mustReadUint8("a")
				r := mustReadUint8("r")

				e = &ProtoExpr{Args: uint8(a), Results: uint8(r)}
			case "byte":
				value := mustReadBytes("value")
				e = &ByteExpr{Value: value}
			case "pushbytes":
				value := mustReadBytes("value")
				e = &PushBytesExpr{Value: value}
			case "pushint":
				value := mustReadInt("value")
				e = &PushIntExpr{Value: value}
			case "asset_params_get":
				mustReadArg("f")

				field, err := readAssetField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse asset_params_get f: %s", args.Text()))
					return
				}

				e = &AssetParamsGetExpr{Field: field}
			case "int":
				value := mustReadInt("value")
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
				value := mustReadUint8("n")
				e = &DigExpr{Index: uint8(value)}
			case "gtxnsas":
				f := mustReadTxnField("f")
				e = &GtxnsasExpr{Field: f}
			case "gitxn":
				t := mustReadUint8("t")
				f := mustReadTxnField("f")
				e = &GitxnExpr{Index: uint8(t), Field: f}
			case "asset_holding_get":
				mustReadArg("f")

				f, err := readAssetHoldingField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to read asset_holding_get f: %s", args.Text()))
					return
				}

				e = &AssetHoldingGetExpr{Field: f}
			case "acct_params_get":
				mustReadArg("f")

				f, err := readAcctParams(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse acct_params_get f: %s", args.Text()))
					return
				}

				e = &AcctParamsGetExpr{Field: f}

			case "app_params_get":
				mustReadArg("f")

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
				mustReadArg("f")
				f, err := readVrfVerifyField(args.Text())
				if err != nil {
					failCurr(errors.Wrapf(err, "failed to parse vrf_verify f: %s", args.Text()))
					return
				}

				e = &VrfVerifyExpr{Field: f}
			case "block":
				mustReadArg("f")

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
				name := mustRead("label name")
				e = &CallSubExpr{Label: &LabelExpr{Name: name}}
			case "assert":
				e = Assert
			case "dup":
				e = Dup
			case "frame_bury":
				value := mustReadInt8("index")
				e = &FrameBuryExpr{Index: value}
			case "frame_dig":
				value := mustReadInt8("index")
				e = &FrameDigExpr{Index: value}
			case "setbyte":
				e = SetByte
			case "uncover":
				value := mustReadUint8("index")
				e = &UncoverExpr{Index: uint8(value)}
			case "cover":
				value := mustReadUint8("n")
				e = &CoverExpr{Index: uint8(value)}
			case "concat":
				e = Concat
			case "itxn_begin":
				e = ItxnBegin
			case "itxn_submit":
				e = ItxnSubmit
			case "itxn":
				f := mustReadTxnField("f")
				e = &ItxnExpr{Field: f}
			case "itxn_field":
				f := mustReadTxnField("f")
				e = &ItxnFieldExpr{Field: f}
			default:
				failCurr(errors.Errorf("unexpected opcode: %s", args.Text()))
				return
			}
		}()

		if e != nil {
			res = append(res, e)
			lts = append(lts, args.ts)
		}
	}

	l := &Linter{l: res}
	l.Lint()

	for _, le := range l.res {
		var b int
		var e int

		lt := lts[le.Line()]
		if len(lt) > 0 {
			b = lt[0].b
			e = lt[len(lt)-1].e
		}

		diag = append(diag, lintError{
			error: le,
			l:     le.Line(),
			b:     b,
			e:     e,
			s:     le.Severity(),
		})
	}

	syms := []Symbol{}
	refs := []Symbol{}

	for i, op := range res {
		switch op := op.(type) {
		case *LabelExpr:
			// TODO: hack - assumes label is the first token on the line
			ts := lts[i]
			t := ts[0]
			syms = append(syms, labelSymbol{
				n: op.Name,
				l: i,
				b: t.b, // TODO: what about whitespaces before label name?
				e: t.e,
			})
		case usesLabels:
			// TODO: this is a hack that assumes label tokens start right after the op which seems to be the case currently but may be changed in the future
			ts := lts[i]
			lbls := op.Labels()
			for j, lbl := range lbls {
				t := ts[j+1]
				refs = append(refs, labelSymbol{
					n: lbl.Name,
					l: i,
					b: t.b,
					e: t.e,
				})
			}
		}
	}

	result := &ProcessResult{
		Version:     version,
		Diagnostics: diag,
		Symbols:     syms,
		SymbolRefs:  refs,
		Tokens:      ts,
		Listing:     res,
		Lines:       lts,
	}

	return result
}
