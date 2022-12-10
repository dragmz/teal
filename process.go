package teal

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var PseudoOps = pseudoOps
var OpDocByName = opDocByName
var OpDocExtras = opDocExtras
var opSpecByName = func() map[string]OpSpec {
	res := map[string]OpSpec{}
	for _, spec := range OpSpecs {
		res[spec.Name] = spec
	}
	return res
}()

type recoverable struct{}

type parserContext struct {
	ops  []Op
	args *arguments
	diag []Diagnostic
}

func (c *parserContext) emit(e Op) {
	c.ops = append(c.ops, e)
}

func (c *parserContext) failAt(l int, b int, e int, err error) {
	c.diag = append(c.diag, parseError{l: l, b: b, e: e, error: err})
	panic(recoverable{})
}

func (c *parserContext) failToken(t Token, err error) {
	c.failAt(t.l, t.b, t.e, err)
}

func (c *parserContext) failCurr(err error) {
	c.failToken(c.args.Curr(), err)
}

func (c *parserContext) failPrev(err error) {
	c.failToken(c.args.Prev(), err)
}

func (c *parserContext) mustReadArg(name string) {
	if !c.args.Scan() {
		c.failPrev(errors.Errorf("missing arg: %s", name))
	}
}

func (c *parserContext) parseUint64(name string) uint64 {
	v, err := readInt(c.args)
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse uint64: %s", name))
	}

	return v
}

func (c *parserContext) parseUint8(name string) uint8 {
	v, err := readUint8(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse uint8: %s", name))
	}
	return v
}

func (c *parserContext) parseInt8(name string) int8 {
	v, err := readInt8(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse int8: %s", name))
	}
	return v
}

func (c *parserContext) parseTxnField(name string) TxnField {
	v, err := readTxnField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse txn field: %s", name))
	}
	return v
}

func (c *parserContext) parseBytes(name string) []byte {
	arg := c.args.Curr().String()

	if strings.HasPrefix(arg, "base32(") || strings.HasPrefix(arg, "b32(") {
		close := strings.IndexRune(arg, ')')
		if close == -1 {
			c.failCurr(errors.New("byte base32 arg lacks close paren"))
		}

		open := strings.IndexRune(arg, '(')
		val, err := base32DecodeAnyPadding(arg[open+1 : close])
		if err != nil {
			c.failCurr(err)
		}
		return val

	}

	if strings.HasPrefix(arg, "base64(") || strings.HasPrefix(arg, "b64(") {
		close := strings.IndexRune(arg, ')')
		if close == -1 {
			c.failCurr(errors.New("byte base64 arg lacks close paren"))
		}

		open := strings.IndexRune(arg, '(')
		val, err := base64.StdEncoding.DecodeString(arg[open+1 : close])
		if err != nil {
			c.failCurr(err)
		}
		return val
	}

	if strings.HasPrefix(arg, "0x") {
		val, err := hex.DecodeString(arg[2:])
		if err != nil {
			c.failCurr(err)
		}
		return val
	}

	if arg == "base32" || arg == "b32" {
		l := c.mustRead("literal")
		val, err := base32DecodeAnyPadding(l)
		if err != nil {
			c.failCurr(err)
		}
		return val
	}

	if arg == "base64" || arg == "b64" {
		l := c.mustRead("literal")
		val, err := base64.StdEncoding.DecodeString(l)
		if err != nil {
			c.failCurr(err)
		}
		return val
	}

	if len(arg) > 1 && arg[0] == '"' && arg[len(arg)-1] == '"' {
		val, err := parseStringLiteral(arg)
		if err != nil {
			c.failCurr(err)
		}
		return val
	}

	c.failCurr(fmt.Errorf("byte arg did not parse: %v", arg))

	return nil
}

func (c *parserContext) parseEcdsaCurveIndex(name string) EcdsaCurve {
	v, err := readEcdsaCurveIndex(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse ESCDS curve index: %s", name))
	}
	return v
}

func (c *parserContext) mustReadBytes(name string) []byte {
	c.mustReadArg(name)
	return c.parseBytes(name)
}

func (c *parserContext) mustReadInt(name string) uint64 {
	c.mustReadArg(name)
	return c.parseUint64(name)
}

func (c *parserContext) mustReadUint8(name string) uint8 {
	c.mustReadArg(name)
	return c.parseUint8(name)
}

func (c *parserContext) mustReadInt8(name string) int8 {
	c.mustReadArg(name)
	return c.parseInt8(name)
}

func (c *parserContext) mustRead(name string) string {
	c.mustReadArg(name)
	return c.args.Text()
}

func (c *parserContext) mustReadTxnField(name string) TxnField {
	c.mustReadArg(name)
	return c.parseTxnField(name)
}

func (c *parserContext) mustReadEcdsaCurveIndex(name string) EcdsaCurve {
	c.mustReadArg(name)
	return c.parseEcdsaCurveIndex(name)
}

type parserFunc func(c *parserContext)

type OpSpecProto struct{}
type OpSpecDetails struct {
	NamesMap map[string]bool
	Names    []string
	Cost     int
}

type OpSpec struct {
	Code      byte
	Name      string
	Parse     parserFunc
	Proto     OpSpecProto
	Version   uint8
	OpDetails OpSpecDetails
}

func assembler(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func costly(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func detDefault() OpSpecDetails {
	return OpSpecDetails{}
}

func detSwitch() OpSpecDetails {
	return OpSpecDetails{}
}

func costByField(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

func typed(vs ...interface{}) OpSpecDetails {
	return OpSpecDetails{}
}

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

var asmPushBytes = []interface{}{}
var immBytes = []interface{}{}
var asmPushInt = []interface{}{}
var immInt = []interface{}{}
var asmPushBytess = []interface{}{}
var asmPushInts = []interface{}{}

func proto(vs ...interface{}) OpSpecProto {
	return OpSpecProto{}
}

func opErr(c *parserContext) {
	c.emit(Err)
}

func opSHA256(c *parserContext) {
	c.emit(Sha256)
}

func opKeccak256(c *parserContext) {
	c.emit(Keccak256)
}

func opSHA512_256(c *parserContext) {
	c.emit(Sha512256)
}
func opEd25519Verify(c *parserContext) {
	c.emit(ED25519Verify)
}

func opEcdsaVerify(c *parserContext) {
	curve := c.mustReadEcdsaCurveIndex("curve index")
	c.emit(&EcdsaVerifyExpr{Index: curve})
}
func opEcdsaPkDecompress(c *parserContext) {
	curve := c.mustReadEcdsaCurveIndex("curve index")
	c.emit(&EcdsaPkDecompressExpr{Index: curve})
}

func opEcdsaPkRecover(c *parserContext) {
	curve := c.mustReadEcdsaCurveIndex("curve index")
	c.emit(&EcdsaPkRecoverExpr{Index: curve})
}

func opPlus(c *parserContext) {
	c.emit(PlusOp)
}
func opMinus(c *parserContext) {
	c.emit(MinusOp)
}
func opDiv(c *parserContext) {
	c.emit(Div)
}
func opMul(c *parserContext) {
	c.emit(Mul)
}
func opLt(c *parserContext) {
	c.emit(Lt)
}
func opGt(c *parserContext) {
	c.emit(Gt)
}
func opLe(c *parserContext) {
	c.emit(Le)
}
func opGe(c *parserContext) {
	c.emit(Ge)
}
func opAnd(c *parserContext) {
	c.emit(And)
}
func opOr(c *parserContext) {
	c.emit(Or)
}
func opEq(c *parserContext) {
	c.emit(Eq)
}
func opNeq(c *parserContext) {
	c.emit(Neq)
}
func opNot(c *parserContext) {
	c.emit(Not)
}
func opLen(c *parserContext) {
	c.emit(Len)
}
func opItob(c *parserContext) {
	c.emit(Itob)
}
func opBtoi(c *parserContext) {
	c.emit(Btoi)
}
func opModulo(c *parserContext) {
	c.emit(Modulo)
}
func opBitOr(c *parserContext) {
	c.emit(Bitr)
}
func opBitAnd(c *parserContext) {
	c.emit(BitAnd)
}
func opBitXor(c *parserContext) {
	c.emit(BitXor)
}
func opBitNot(c *parserContext) {
	c.emit(BitNot)
}
func opMulw(c *parserContext) {
	c.emit(Mulw)
}
func opAddw(c *parserContext) {
	c.emit(Addw)
}
func opDivModw(c *parserContext) {
	c.emit(DivModw)
}
func opIntConstBlock(c *parserContext) {
	var values []uint64

	for c.args.Scan() {
		value := c.parseUint64("value")
		values = append(values, value)
	}

	c.emit(&IntcBlockExpr{Values: values})
}

func opIntConstLoad(c *parserContext) {
	value := c.mustReadUint8("value")
	c.emit(&IntcExpr{Index: uint8(value)})
}

func opIntConst0(c *parserContext) {
	c.emit(Intc0)
}
func opIntConst1(c *parserContext) {
	c.emit(Intc1)
}
func opIntConst2(c *parserContext) {
	c.emit(Intc2)
}
func opIntConst3(c *parserContext) {
	c.emit(Intc3)
}
func opByteConstBlock(c *parserContext) {
	var values [][]byte

	for c.args.Scan() {
		b := c.parseBytes("value")
		values = append(values, b)
	}

	c.emit(&BytecBlockExpr{Values: values})
}

func opByteConstLoad(c *parserContext) {
	value := c.mustReadUint8("index")
	c.emit(&BytecExpr{Index: uint8(value)})
}

func opByteConst0(c *parserContext) {
	c.emit(Bytec0)
}
func opByteConst1(c *parserContext) {
	c.emit(Bytec1)
}
func opByteConst2(c *parserContext) {
	c.emit(Bytec2)
}
func opByteConst3(c *parserContext) {
	c.emit(Bytec3)
}
func opArg(c *parserContext) {
	value := c.mustReadUint8("index")
	c.emit(&ArgExpr{Index: uint8(value)})
}
func opArg0(c *parserContext) {
	c.emit(Arg0)
}
func opArg1(c *parserContext) {
	c.emit(Arg1)
}
func opArg2(c *parserContext) {
	c.emit(Arg2)
}
func opArg3(c *parserContext) {
	c.emit(Arg3)
}
func opTxn(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&TxnExpr{Field: f})
}
func opGlobal(c *parserContext) {
	c.mustReadArg("field")

	field, err := readGlobalField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse global field: %s", c.args.Text()))
		return
	}

	c.emit(&GlobalExpr{Index: field})
}

func opGtxn(c *parserContext) {
	t := c.mustReadInt("t")
	f := c.mustReadTxnField("f")
	c.emit(&GtxnExpr{Index: uint8(t), Field: f})
}

func opLoad(c *parserContext) {
	value := c.mustReadUint8("i")
	c.emit(&LoadExpr{Index: uint8(value)})
}
func opStore(c *parserContext) {
	value := c.mustReadUint8("i")
	c.emit(&StoreExpr{Index: uint8(value)})
}

func opTxna(c *parserContext) {
	f := c.mustReadTxnField("f")
	i := c.mustReadUint8("i")
	c.emit(&TxnaExpr{Field: f, Index: i})
}

func opGtxna(c *parserContext) {
	t := c.mustReadUint8("t")
	f := c.mustReadTxnField("f")
	i := c.mustReadUint8("i")
	c.emit(&GtxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)})
}
func opGtxns(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&GtxnsExpr{Field: f})
}

func opGtxnsa(c *parserContext) {
	f := c.mustReadTxnField("f")
	i := c.mustReadUint8("i")
	c.emit(&GtxnsaExpr{Field: f, Index: uint8(i)})
}
func opGload(c *parserContext) {
	t := c.mustReadUint8("t")
	value := c.mustReadUint8("i")

	c.emit(&GloadExpr{Group: uint8(t), Index: uint8(value)})
}

func opGloads(c *parserContext) {
	value := c.mustReadUint8("i")
	c.emit(&GloadsExpr{Index: uint8(value)})
}

func opGaid(c *parserContext) {
	t := c.mustReadUint8("t")
	c.emit(&GaidExpr{Group: uint8(t)})
}
func opGaids(c *parserContext) {
	c.emit(Gaids)
}
func opLoads(c *parserContext) {
	c.emit(Loads)
}
func opStores(c *parserContext) {
	c.emit(Stores)
}
func opBnz(c *parserContext) {
	name := c.mustRead("label name")
	c.emit(&BnzExpr{Label: &LabelExpr{Name: name}})
}
func opBz(c *parserContext) {
	name := c.mustRead("label name")
	c.emit(&BzExpr{Label: &LabelExpr{Name: name}})
}
func opB(c *parserContext) {
	name := c.mustRead("label name")
	c.emit(&BExpr{Label: &LabelExpr{Name: name}})
}
func opReturn(c *parserContext) {
	c.emit(Return)
}
func opAssert(c *parserContext) {
	c.emit(Assert)
}
func opBury(c *parserContext) {
	n := c.mustReadUint8("n")
	c.emit(&BuryExpr{Index: n})
}
func opPopN(c *parserContext) {
	n := c.mustReadUint8("n")
	c.emit(&PopNExpr{Index: n})
}
func opDupN(c *parserContext) {
	n := c.mustReadUint8("n")
	c.emit(&DupNExpr{Index: n})
}

func opPop(c *parserContext) {
	c.emit(Pop)
}
func opDup(c *parserContext) {
	c.emit(Dup)
}
func opDup2(c *parserContext) {
	c.emit(Dup2)
}
func opDig(c *parserContext) {
	value := c.mustReadUint8("n")
	c.emit(&DigExpr{Index: uint8(value)})
}
func opSwap(c *parserContext) {
	c.emit(Swap)
}
func opSelect(c *parserContext) {
	c.emit(Select)
}
func opCover(c *parserContext) {
	value := c.mustReadUint8("n")
	c.emit(&CoverExpr{Index: uint8(value)})
}
func opUncover(c *parserContext) {
	value := c.mustReadUint8("index")
	c.emit(&UncoverExpr{Index: uint8(value)})
}
func opConcat(c *parserContext) {
	c.emit(Concat)
}
func opSubstring(c *parserContext) {
	start := c.mustReadUint8("s")
	end := c.mustReadUint8("e")
	c.emit(&SubstringExpr{Start: uint8(start), End: uint8(end)})
}
func opSubstring3(c *parserContext) {
	c.emit(Substring3)
}
func opGetBit(c *parserContext) {
	c.emit(GetBit)
}
func opSetBit(c *parserContext) {
	c.emit(SetBit)
}
func opGetByte(c *parserContext) {
	c.emit(GetByte)
}
func opSetByte(c *parserContext) {
	c.emit(SetByte)
}
func opExtract(c *parserContext) {
	start := c.mustReadUint8("s")
	length := c.mustReadUint8("l")

	c.emit(&ExtractExpr{Start: uint8(start), Length: uint8(length)})
}

func opExtract3(c *parserContext) {
	c.emit(Extract3)
}
func opExtract16Bits(c *parserContext) {
	c.emit(Extract16Bits)
}
func opExtract32Bits(c *parserContext) {
	c.emit(Extract32Bits)
}
func opExtract64Bits(c *parserContext) {
	c.emit(Extract64Bits)
}
func opReplace2(c *parserContext) {
	value := c.mustReadUint8("s")
	c.emit(&Replace2Expr{Index: uint8(value)})
}
func opReplace3(c *parserContext) {
	c.emit(Replace3)
}
func opBase64Decode(c *parserContext) {
	value := c.mustReadUint8("e")
	c.emit(&Base64DecodeExpr{Index: uint8(value)})
}
func opJSONRef(c *parserContext) {
	value := c.mustReadUint8("r")
	c.emit(&JsonRefExpr{Index: uint8(value)})
}
func opBalance(c *parserContext) {
	c.emit(Balance)
}
func opAppOptedIn(c *parserContext) {
	c.emit(AppOptedIn)
}
func opAppLocalGet(c *parserContext) {
	c.emit(AppLocalGet)
}
func opAppLocalGetEx(c *parserContext) {
	c.emit(AppLocalGetEx)
}
func opAppGlobalGet(c *parserContext) {
	c.emit(AppGlobalGet)
}
func opAppGlobalGetEx(c *parserContext) {
	c.emit(AppGlobalGetEx)
}
func opAppLocalPut(c *parserContext) {
	c.emit(AppLocalPut)
}
func opAppGlobalPut(c *parserContext) {
	c.emit(AppGlobalPut)
}
func opAppLocalDel(c *parserContext) {
	c.emit(AppLocalDel)
}
func opAppGlobalDel(c *parserContext) {
	c.emit(AppGlobalDel)
}
func opAssetHoldingGet(c *parserContext) {
	c.mustReadArg("f")

	f, err := readAssetHoldingField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to read asset_holding_get f: %s", c.args.Text()))
		return
	}

	c.emit(&AssetHoldingGetExpr{Field: f})
}
func opAssetParamsGet(c *parserContext) {
	c.mustReadArg("f")

	field, err := readAssetField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse asset_params_get f: %s", c.args.Text()))
		return
	}

	c.emit(&AssetParamsGetExpr{Field: field})
}
func opAppParamsGet(c *parserContext) {
	c.mustReadArg("f")

	f, err := readAppField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse app_params_get f: %s", c.args.Text()))
		return
	}

	c.emit(&AppParamsGetExpr{Field: f})
}
func opAcctParamsGet(c *parserContext) {
	c.mustReadArg("f")

	f, err := readAcctParams(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse acct_params_get f: %s", c.args.Text()))
		return
	}

	c.emit(&AcctParamsGetExpr{Field: f})
}
func opMinBalance(c *parserContext) {
	c.emit(MinBalanceOp)
}
func opPushBytes(c *parserContext) {
	value := c.mustReadBytes("value")
	c.emit(&PushBytesExpr{Value: value})
}
func opPushInt(c *parserContext) {
	value := c.mustReadInt("value")
	c.emit(&PushIntExpr{Value: value})
}
func opPushBytess(c *parserContext) {
	// TODO
}
func opEd25519VerifyBare(c *parserContext) {
	c.emit(Ed25519VerifyBare)
}
func opPushInts(c *parserContext) {
	// TODO
}
func opCallSub(c *parserContext) {
	name := c.mustRead("label name")
	c.emit(&CallSubExpr{Label: &LabelExpr{Name: name}})
}
func opRetSub(c *parserContext) {
	c.emit(RetSub)
}
func opProto(c *parserContext) {
	a := c.mustReadUint8("a")
	r := c.mustReadUint8("r")

	c.emit(&ProtoExpr{Args: uint8(a), Results: uint8(r)})
}
func opFrameDig(c *parserContext) {
	value := c.mustReadInt8("index")
	c.emit(&FrameDigExpr{Index: value})
}
func opFrameBury(c *parserContext) {
	value := c.mustReadInt8("index")
	c.emit(&FrameBuryExpr{Index: value})
}
func opSwitch(c *parserContext) {
	var labels []*LabelExpr
	for c.args.Scan() {
		labels = append(labels, &LabelExpr{Name: c.args.Text()})
	}
	c.emit(&SwitchExpr{Targets: labels})
}
func opMatch(c *parserContext) {
	var labels []*LabelExpr
	for c.args.Scan() {
		labels = append(labels, &LabelExpr{Name: c.args.Text()})
	}
	c.emit(&MatchExpr{Targets: labels})
}
func opShiftLeft(c *parserContext) {
	c.emit(ShiftLeft)
}
func opShiftRight(c *parserContext) {
	c.emit(ShiftRight)
}
func opSqrt(c *parserContext) {
	c.emit(Sqrt)
}
func opBitLen(c *parserContext) {
	c.emit(BitLen)
}
func opExp(c *parserContext) {
	c.emit(Exp)
}
func opExpw(c *parserContext) {
	c.emit(Expw)
}
func opBytesSqrt(c *parserContext) {
	c.emit(Bsqrt)
}
func opDivw(c *parserContext) {
	c.emit(Divw)
}
func opSHA3_256(c *parserContext) {
	c.emit(Sha3256)
}
func opBn256Add(c *parserContext) {
	// TODO
}
func opBn256ScalarMul(c *parserContext) {
	// TODO
}
func opBn256Pairing(c *parserContext) {
	// TODO
}
func opBytesPlus(c *parserContext) {
	c.emit(BytesPlus)
}
func opBytesMinus(c *parserContext) {
	c.emit(BytesMinus)
}
func opBytesDiv(c *parserContext) {
	c.emit(BytesDiv)
}
func opBytesMul(c *parserContext) {
	c.emit(BytesMul)
}
func opBytesLt(c *parserContext) {
	c.emit(BytesLt)
}
func opBytesGt(c *parserContext) {
	c.emit(BytesGt)
}
func opBytesLe(c *parserContext) {
	c.emit(BytesLe)
}
func opBytesGe(c *parserContext) {
	c.emit(BytesGe)
}
func opBytesEq(c *parserContext) {
	c.emit(BytesEq)
}
func opBytesNeq(c *parserContext) {
	c.emit(BytesNeq)
}
func opBytesModulo(c *parserContext) {
	c.emit(BytesModulo)
}
func opBytesBitOr(c *parserContext) {
	c.emit(BytesBitOr)
}
func opBytesBitAnd(c *parserContext) {
	c.emit(BytesBitAnd)
}
func opBytesBitXor(c *parserContext) {
	c.emit(BytesBitXor)
}
func opBytesBitNot(c *parserContext) {
	c.emit(BytesBitNot)
}
func opBytesZero(c *parserContext) {
	c.emit(BytesZero)
}
func opLog(c *parserContext) {
	c.emit(Log)
}
func opTxBegin(c *parserContext) {
	c.emit(ItxnBegin)
}
func opItxnField(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&ItxnFieldExpr{Field: f})
}
func opItxnSubmit(c *parserContext) {
	c.emit(ItxnSubmit)
}
func opItxn(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&ItxnExpr{Field: f})
}
func opItxna(c *parserContext) {
	f := c.mustReadUint8("f")
	i := c.mustReadUint8("i")
	c.emit(&ItxnaExpr{Field: f, Index: i})
}
func opItxnNext(c *parserContext) {
	c.emit(ItxnNext)
}
func opGitxn(c *parserContext) {
	t := c.mustReadUint8("t")
	f := c.mustReadTxnField("f")
	c.emit(&GitxnExpr{Index: uint8(t), Field: f})
}
func opGitxna(c *parserContext) {
	t := c.mustReadInt("t")
	f := c.mustReadTxnField("f")
	i := c.mustReadUint8("i")

	c.emit(&GitxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)})

}
func opBoxCreate(c *parserContext) {
	c.emit(BoxCreate)
}
func opBoxExtract(c *parserContext) {
	c.emit(BoxExtract)
}
func opBoxReplace(c *parserContext) {
	c.emit(BoxReplace)
}
func opBoxDel(c *parserContext) {
	c.emit(BoxDel)
}
func opBoxLen(c *parserContext) {
	c.emit(BoxLen)
}
func opBoxGet(c *parserContext) {
	c.emit(BoxGet)
}
func opBoxPut(c *parserContext) {
	c.emit(BoxPut)
}
func opTxnas(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&TxnasExpr{Field: f})
}
func opGtxnas(c *parserContext) {
	t := c.mustReadUint8("t")
	f := c.mustReadUint8("f")
	c.emit(&GtxnasExpr{Index: t, Field: f})
}
func opGtxnsas(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&GtxnsasExpr{Field: f})
}
func opArgs(c *parserContext) {
	c.emit(Args)
}
func opGloadss(c *parserContext) {
	c.emit(Gloadss)
}
func opItxnas(c *parserContext) {
	f := c.mustReadTxnField("f")
	c.emit(&ItxnasExpr{Field: f})
}
func opGitxnas(c *parserContext) {
	t := c.mustReadUint8("t")
	f := c.mustReadTxnField("f")
	c.emit(&GitxnasExpr{Index: t, Field: f})
}
func opVrfVerify(c *parserContext) {
	c.mustReadArg("f")
	f, err := readVrfVerifyField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse vrf_verify f: %s", c.args.Text()))
		return
	}

	c.emit(&VrfVerifyExpr{Field: f})
}
func opBlock(c *parserContext) {
	c.mustReadArg("f")

	f, err := readBlockField(c.args.Text())
	if err != nil {
		c.failCurr(errors.Wrapf(err, "failed to parse block f: %s", c.args.Text()))
		return
	}

	c.emit(&BlockExpr{Field: f})
}

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
	c := &parserContext{
		ops: []Op{},
	}

	var ts []Token
	ts, c.diag = readTokens(source)

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
	version := uint8(1)

	for _, l := range lines {
		c.args = &arguments{ts: l}
		func() {
			defer func() {
				switch v := recover().(type) {
				case recoverable:
					c.emit(Empty) // consider replacing with Raw string expr
				case nil:
				default:
					fmt.Printf("unrecoverable: %v", v)
					panic(v)
				}
			}()

			if !c.args.Scan() {
				c.emit(Empty)
				return
			}

			if strings.HasSuffix(c.args.Text(), ":") {
				name := c.args.Text()
				name = name[:len(name)-1]
				if len(name) == 0 {
					c.failCurr(errors.New("missing label name"))
					return
				}
				c.emit(&LabelExpr{Name: name})
				return
			}

			name := c.args.Text()
			switch c.args.Text() {
			case "":
				c.emit(Empty)
			case "#pragma":
				name := c.mustRead("name")
				switch name {
				case "version":
					version = c.mustReadUint8("version value")
					c.emit(&PragmaExpr{Version: uint8(version)})
				default:
					c.failCurr(errors.Errorf("unexpected #pragma: %s", c.args.Text()))
					return
				}
			case "byte":
				value := c.mustReadBytes("value")
				c.emit(&ByteExpr{Value: value})
			case "int":
				value := c.mustReadInt("value")
				c.emit(&IntExpr{Value: value})
			default:
				spec, ok := opSpecByName[name]
				if ok {
					spec.Parse(c)
				} else {
					c.failCurr(errors.Errorf("unexpected opcode: %s", c.args.Text()))
				}
				return
			}
		}()

		lts = append(lts, c.args.ts)
	}

	l := &Linter{l: c.ops}
	l.Lint()

	for _, le := range l.res {
		var b int
		var e int

		lt := lts[le.Line()]
		if len(lt) > 0 {
			b = lt[0].b
			e = lt[len(lt)-1].e
		}

		c.diag = append(c.diag, lintError{
			error: le,
			l:     le.Line(),
			b:     b,
			e:     e,
			s:     le.Severity(),
		})
	}

	syms := []Symbol{}
	refs := []Symbol{}

	for i, op := range c.ops {
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
		Diagnostics: c.diag,
		Symbols:     syms,
		SymbolRefs:  refs,
		Tokens:      ts,
		Listing:     c.ops,
		Lines:       lts,
	}

	return result
}
