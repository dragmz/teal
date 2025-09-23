package teal

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Nop interface {
	IsNop()
}

// err

type ErrExpr struct{}

func (e *ErrExpr) String() string {
	return "err"
}

func (e *ErrExpr) IsTerminator() {}

var Err = &ErrExpr{}

// sha3_256

type Sha3256Expr struct{}

func (e *Sha3256Expr) String() string {
	return "sha3_256"
}

var Sha3256 = &Sha3256Expr{}

// sumhash512

type Sumhash512Expr struct{}

func (e *Sumhash512Expr) String() string {
	return "sumhash512"
}

var SumHash512 = &Sumhash512Expr{}

// falcon_verify

type FalconVerifyExpr struct{}

func (e *FalconVerifyExpr) String() string {
	return "falcon_verify"
}

var FalconVerify = &FalconVerifyExpr{}

// sha256

type Sha256Expr struct{}

func (e *Sha256Expr) String() string {
	return e.Name()
}

func (e *Sha256Expr) Name() string {
	return "sha256"
}

var Sha256 = &Sha256Expr{}

// keccak256

type Keccak256Expr struct{}

func (e *Keccak256Expr) String() string {
	return "keccak256"
}

var Keccak256 = &Keccak256Expr{}

// sha512_256

type Sha512256Expr struct{}

func (e *Sha512256Expr) String() string {
	return "sha512_256"
}

var Sha512256 = &Sha512256Expr{}

// ed25519verify

type ED25519VerifyExpr struct{}

func (e *ED25519VerifyExpr) String() string {
	return e.Name()
}

func (e *ED25519VerifyExpr) Name() string {
	return "ed25519verify"
}

var ED25519Verify = &ED25519VerifyExpr{}

// ecdsa_verify v

type EcdsaVerifyExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaVerifyExpr) String() string {
	return fmt.Sprintf("%s %d", e.Name(), e.Index)
}

func (e *EcdsaVerifyExpr) Name() string {
	return "ecdsa_verify"
}

// ecdsa_pk_decompress

type EcdsaPkDecompressExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaPkDecompressExpr) String() string {
	return fmt.Sprintf("ecdsa_pk_decompress %d", e.Index)
}

// ecdsa_pk_recover

type EcdsaPkRecoverExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaPkRecoverExpr) String() string {
	return fmt.Sprintf("ecdsa_pk_recover %d", e.Index)
}

// +

type PlusExpr struct {
}

func (e *PlusExpr) String() string {
	return "+"
}

var PlusOp = &PlusExpr{}

// -

type MinusExpr struct {
}

func (e *MinusExpr) String() string {
	return "-"
}

var MinusOp = &MinusExpr{}

// /

type DivExpr struct{}

func (e *DivExpr) String() string {
	return "/"
}

var Div = &DivExpr{}

// *

type MulExpr struct{}

func (e *MulExpr) String() string {
	return "*"
}

var Mul = &MulExpr{}

// <

type LtExpr struct {
}

func (e *LtExpr) String() string {
	return "<"
}

var Lt = &LtExpr{}

// >

type GtExpr struct {
}

func (e *GtExpr) String() string {
	return ">"
}

var Gt = &GtExpr{}

// <=

type LtEqExpr struct {
}

func (e *LtEqExpr) String() string {
	return "<="
}

var Le = &LtEqExpr{}

// >=

type GtEqExpr struct {
}

func (e *GtEqExpr) String() string {
	return ">="
}

var Ge = &GtEqExpr{}

// &&

type AndExpr struct {
}

func (e *AndExpr) String() string {
	return "&&"
}

var And = &AndExpr{}

// ||

type OrExpr struct {
}

func (e *OrExpr) String() string {
	return "||"
}

var Or = &OrExpr{}

// ==

type EqExpr struct {
}

func (e *EqExpr) String() string {
	return "=="
}

var Eq = &EqExpr{}

// !=

type NeqExpr struct {
}

func (e *NeqExpr) String() string {
	return "!="
}

var Neq = &NeqExpr{}

// !

type NotExpr struct{}

func (e *NotExpr) String() string {
	return "!"
}

var Not = &NotExpr{}

// len

type LenExpr struct{}

func (e *LenExpr) String() string {
	return "len"
}

var Len = &LenExpr{}

// itob

type ItobExpr struct{}

func (e *ItobExpr) String() string {
	return "itob"
}

var Itob = &ItobExpr{}

// btoi

type BtoiExpr struct{}

func (e *BtoiExpr) String() string {
	return "btoi"
}

var Btoi = &BtoiExpr{}

// %

type ModExpr struct{}

func (e *ModExpr) String() string {
	return "%"
}

var Modulo = &ModExpr{}

// |

type BitOrExpr struct{}

func (e *BitOrExpr) String() string {
	return "|"
}

var BitOr = &BitOrExpr{}

// &

type BitAndExpr struct{}

func (e *BitAndExpr) String() string {
	return "&"
}

var BitAnd = &BitAndExpr{}

// ^

type BitXorExpr struct{}

func (e *BitXorExpr) String() string {
	return "^"
}

var BitXor = &BitXorExpr{}

// ~

type BitNotExpr struct{}

func (e *BitNotExpr) String() string {
	return "~"
}

var BitNot = &BitNotExpr{}

// mulw

type MulwExpr struct{}

func (e *MulwExpr) String() string {
	return "mulw"
}

var Mulw = &MulwExpr{}

// addw

type AddwExpr struct{}

func (e *AddwExpr) String() string {
	return "addw"
}

var Addw = &AddwExpr{}

// divmodw

type DivModwExpr struct{}

func (e *DivModwExpr) String() string {
	return "divmodw"
}

var DivModw = &DivModwExpr{}

// intcblock uint ...

type IntcBlockExpr struct {
	Values []uint64
}

func (e *IntcBlockExpr) String() string {
	if len(e.Values) == 0 {
		return "intcblock"
	}

	strs := make([]string, len(e.Values))

	for i, v := range e.Values {
		strs[i] = strconv.FormatUint(uint64(v), 10)
	}

	str := strings.Join(strs, " ")

	return fmt.Sprintf("intcblock %s", str)
}

// intc i

type IntcExpr struct {
	Index uint8
}

func (e *IntcExpr) String() string {
	return fmt.Sprintf("intc %d", e.Index)
}

// intc_0

type Intc0Expr struct{}

func (e *Intc0Expr) String() string {
	return "intc_0"
}

var Intc0 = &Intc0Expr{}

// intc_1

type Intc1Expr struct{}

func (e *Intc1Expr) String() string {
	return "intc_1"
}

var Intc1 = &Intc1Expr{}

// intc_2

type Intc2Expr struct{}

func (e *Intc2Expr) String() string {
	return "intc_2"
}

var Intc2 = &Intc2Expr{}

// intc_3

type Intc3Expr struct{}

func (e *Intc3Expr) String() string {
	return "intc_3"
}

var Intc3 = &Intc3Expr{}

// bytecblock bytes ...

type BytecBlockExpr struct {
	Values [][]byte
}

func (e *BytecBlockExpr) String() string {
	if len(e.Values) == 0 {
		return "bytecblock"
	}

	strs := make([]string, len(e.Values))

	for i, v := range e.Values {
		strs[i] = fmt.Sprintf("0x%s", hex.EncodeToString(v))
	}

	str := strings.Join(strs, " ")

	return fmt.Sprintf("bytecblock %s", str)
}

// bytec 1

type BytecExpr struct {
	Index uint8
}

func (e *BytecExpr) String() string {
	return fmt.Sprintf("bytec %d", e.Index)
}

// bytec_0

type Bytec0Expr struct{}

func (e *Bytec0Expr) String() string {
	return "bytec_0"
}

var Bytec0 = &Bytec0Expr{}

// bytec_1

type Bytec1Expr struct{}

func (e *Bytec1Expr) String() string {
	return "bytec_1"
}

var Bytec1 = &Bytec1Expr{}

// bytec_2

type Bytec2Expr struct{}

func (e *Bytec2Expr) String() string {
	return "bytec_2"
}

var Bytec2 = &Bytec2Expr{}

// bytec_3

type Bytec3Expr struct{}

func (e *Bytec3Expr) String() string {
	return "bytec_3"
}

var Bytec3 = &Bytec3Expr{}

// arg n

type ArgExpr struct {
	Index uint8
}

func (e *ArgExpr) String() string {
	return fmt.Sprintf("arg %d", e.Index)
}

// arg_0

type Arg0Expr struct{}

func (e *Arg0Expr) String() string {
	return "arg_0"
}

var Arg0 = &Arg0Expr{}

// arg_1

type Arg1Expr struct{}

func (e *Arg1Expr) String() string {
	return "arg_1"
}

var Arg1 = &Arg1Expr{}

// arg_2

type Arg2Expr struct{}

func (e *Arg2Expr) String() string {
	return "arg_2"
}

var Arg2 = &Arg2Expr{}

// arg_3

type Arg3Expr struct{}

func (e *Arg3Expr) String() string {
	return "arg_3"
}

var Arg3 = &Arg3Expr{}

// txn f

type TxnExpr struct {
	Field TxnField
}

func (e *TxnExpr) String() string {
	return fmt.Sprintf("txn %s", e.Field)
}

// global f

const (
	GlobalMinTxnFee                 = GlobalField(0)
	GlobalMinBalance                = GlobalField(1)
	GlobalMaxTxnLife                = GlobalField(2)
	GlobalZeroAddress               = GlobalField(3)
	GlobalGroupSize                 = GlobalField(4)
	GlobalLogicSigVersion           = GlobalField(5)
	GlobalRound                     = GlobalField(6)
	GlobalLatestTimestamp           = GlobalField(7)
	GlobalCurrentApplicationID      = GlobalField(8)
	GlobalCreatorAddress            = GlobalField(9)
	GlobalCurrentApplicationAddress = GlobalField(10)
	GlobalGroupID                   = GlobalField(11)
	GlobalOpcodeBudget              = GlobalField(12)
	GlobalCallerApplicationID       = GlobalField(13)
	GlobalCallerApplicationAddress  = GlobalField(14)
	GlobalAssetCreateMinbalance     = GlobalField(15)
	GlobalAssetOptInMinBalance      = GlobalField(16)
)

type GlobalExpr struct {
	Field GlobalField
}

func (e *GlobalExpr) String() string {
	return fmt.Sprintf("global %s", e.Field)
}

// load i

type LoadExpr struct {
	Index uint8
}

func (e *LoadExpr) String() string {
	return fmt.Sprintf("load %d", e.Index)
}

// store i

type StoreExpr struct {
	Index uint8
}

func (e *StoreExpr) String() string {
	return fmt.Sprintf("store %d", e.Index)
}

// txna f i

type TxnaExpr struct {
	Field TxnField
	Index uint8
}

func (e *TxnaExpr) String() string {
	return fmt.Sprintf("txna %s %d", e.Field, e.Index)
}

type GtxnExpr struct {
	Group uint8
	Field TxnField
}

func (e *GtxnExpr) String() string {
	return fmt.Sprintf("gtxn %d %s", e.Group, e.Field)
}

// gtxns f

type GtxnsExpr struct {
	Field TxnField
}

func (e *GtxnsExpr) String() string {
	return fmt.Sprintf("gtxns %s", e.Field)
}

type ProtoExpr struct {
	Args    uint8
	Results uint8
}

func (e *ProtoExpr) String() string {
	return fmt.Sprintf("proto %d %d", e.Args, e.Results)
}

type LabelExpr struct {
	Name string
}

func (e *LabelExpr) IsNop() {}

func (e *LabelExpr) String() string {
	return fmt.Sprintf("%s:", e.Name)
}

func (e *LabelExpr) GetLabel() *LabelExpr {
	return e
}

type NestedExpr struct {
	Label *LabelExpr
	Body  []Expr
}

type FuncExpr struct {
	Label *LabelExpr
	Proto *ProtoExpr
	Block *NestedExpr
}

func (e *FuncExpr) Call(args ...Expr) Expr {
	var exprs []Expr

	exprs = append(exprs, args...)
	exprs = append(exprs, &CallSubExpr{Label: e.Label})

	return &NestedExpr{Body: exprs}
}

func (e *FuncExpr) GetLabel() *LabelExpr {
	return e.Label
}

const (
	TypeEnumUnknown = iota
	TypeEnumPay
	TypeEnumKeyReg
	TypeEnumAcfg
	TypeEnumAxfer
	TypeEnumAfrz
	TypeEnumAppl
)

type PushBytesExpr struct {
	Value  []byte
	Format BytesFormat
}

func (e *PushBytesExpr) String() string {
	return fmt.Sprintf("pushbytes %s", Bytes{Value: e.Value, Format: e.Format})
}

type PushIntExpr struct {
	Value uint64
}

func (e *PushIntExpr) String() string {
	return fmt.Sprintf("pushint %d", e.Value)
}

type IntExpr struct {
	Value uint64
}

func (e *IntExpr) String() string {
	return fmt.Sprintf("int %d", e.Value)
}

type BytesFormat int

const (
	BytesBase64        = 0
	BytesStringLiteral = 1
)

type ByteExpr struct {
	Value  []byte
	Format BytesFormat
}

type Bytes struct {
	Value  []byte
	Format BytesFormat
}

func (e Bytes) String() string {
	switch e.Format {
	case BytesStringLiteral:
		return fmt.Sprintf("\"%s\"", string(e.Value))
	case BytesBase64:
		return fmt.Sprintf("b64 %s", base64.StdEncoding.EncodeToString(e.Value))
	default:
		panic(fmt.Sprintf("unsupported bytes format: %d", e.Format))
	}

}

func (e *ByteExpr) String() string {
	return fmt.Sprintf("byte %s", Bytes{Value: e.Value, Format: e.Format})
}

type BoxGetExpr struct {
}

func (e *BoxGetExpr) String() string {
	return "box_get"
}

var BoxGet = &BoxGetExpr{}

type BoxPutExpr struct {
}

func (e *BoxPutExpr) String() string {
	return "box_put"
}

var BoxPut = &BoxPutExpr{}

type BoxSpliceExpr struct {
}

func (e *BoxSpliceExpr) String() string {
	return "box_splice"
}

var BoxSplice = &BoxSpliceExpr{}

type BoxResizeExpr struct {
}

func (e *BoxResizeExpr) String() string {
	return "box_resize"
}

var BoxResize = &BoxResizeExpr{}

type BoxCreateExpr struct {
}

type BoxLenExpr struct {
}

func (e *BoxLenExpr) String() string {
	return "box_len"
}

var BoxDel = &BoxDelExpr{}

type BoxDelExpr struct {
}

func (e *BoxDelExpr) String() string {
	return "box_del"
}

var BoxLen = &BoxLenExpr{}

func (e *BoxCreateExpr) String() string {
	return "box_create"
}

var BoxCreate = &BoxCreateExpr{}

type BoxReplaceExpr struct {
}

var BoxReplace = &BoxReplaceExpr{}

func (e *BoxReplaceExpr) String() string {
	return "box_replace"
}

type BoxExtractExpr struct {
}

var BoxExtract = &BoxExtractExpr{}

func (e *BoxExtractExpr) String() string {
	return "box_extract"
}

type DefineExpr struct {
	Name string
}

func (e *DefineExpr) String() string {
	// TODO
	return ""
}

func (e *DefineExpr) IsNop() {}

type PragmaExpr struct {
	Version uint64
}

func (e *PragmaExpr) IsNop() {}

func (e *PragmaExpr) String() string {
	return fmt.Sprintf("#pragma version %d", e.Version)
}

type BnzExpr struct {
	Label *LabelExpr
}

func (e *BnzExpr) Labels() []*LabelExpr {
	return []*LabelExpr{e.Label}
}

func (e *BnzExpr) String() string {
	return fmt.Sprintf("bnz %s", e.Label.Name)
}

type BzExpr struct {
	Label *LabelExpr
}

func (e *BzExpr) Labels() []*LabelExpr {
	return []*LabelExpr{e.Label}
}

func (e *BzExpr) String() string {
	return fmt.Sprintf("bz %s", e.Label.Name)
}

type BExpr struct {
	Label *LabelExpr
}

func (e *BExpr) Labels() []*LabelExpr {
	return []*LabelExpr{e.Label}
}

func (e *BExpr) String() string {
	return fmt.Sprintf("b %s", e.Label.Name)
}

type EmptyExpr struct {
}

func (e *EmptyExpr) String() string {
	return ""
}

func (e *EmptyExpr) IsNop() {}

var Empty = &EmptyExpr{}

type AppLocalGetExpr struct{}

func (e *AppLocalGetExpr) String() string {
	return "app_local_get"
}

var AppLocalGet = &AppLocalGetExpr{}

type AppLocalPutExpr struct{}

func (e *AppLocalPutExpr) String() string {
	return "app_local_put"
}

var AppLocalPut = &AppLocalPutExpr{}

type AppGlobalPutExpr struct{}

func (e *AppGlobalPutExpr) String() string {
	return "app_global_put"
}

var AppGlobalPut = &AppGlobalPutExpr{}

type AppGlobalGetExpr struct{}

func (e *AppGlobalGetExpr) String() string {
	return "app_global_get"
}

var AppGlobalGet = &AppGlobalGetExpr{}

type AppGlobalGetExExpr struct{}

func (e *AppGlobalGetExExpr) String() string {
	return "app_global_get_ex"
}

var AppGlobalGetEx = &AppGlobalGetExExpr{}

type AppLocalGetExExpr struct{}

func (e *AppLocalGetExExpr) String() string {
	return "app_local_get_ex"
}

var AppLocalGetEx = &AppLocalGetExExpr{}

type AppLocalDelExpr struct{}

func (e *AppLocalDelExpr) String() string {
	return "app_local_del"
}

var AppLocalDel = &AppLocalDelExpr{}

type RetSubExpr struct{}

var RetSub = &RetSubExpr{}

func (e *RetSubExpr) String() string {
	return "retsub"
}

func (e *RetSubExpr) IsTerminator() {}

type ReturnExpr struct {
}

func (e *ReturnExpr) String() string {
	return "return"
}

func (e *ReturnExpr) IsTerminator() {}

type SwitchExpr struct {
	Targets []*LabelExpr
}

func (e *SwitchExpr) Labels() []*LabelExpr {
	return e.Targets
}

func (e *SwitchExpr) String() string {
	var b strings.Builder

	b.WriteString("switch")

	if len(e.Targets) > 0 {
		b.WriteString(" ")

		names := make([]string, len(e.Targets))
		for i, l := range e.Targets {
			names[i] = l.Name
		}

		b.WriteString(strings.Join(names, " "))
	}

	b.WriteString("")

	return b.String()
}

type MatchExpr struct {
	Targets []*LabelExpr
}

func (e *MatchExpr) Labels() []*LabelExpr {
	return e.Targets
}

func (e *MatchExpr) String() string {
	var b strings.Builder

	b.WriteString("match")

	if len(e.Targets) > 0 {
		b.WriteString(" ")

		names := make([]string, len(e.Targets))
		for i, l := range e.Targets {
			names[i] = l.Name
		}

		b.WriteString(strings.Join(names, " "))
	}

	b.WriteString("")

	return b.String()
}

type CallSubExpr struct {
	Label *LabelExpr
}

func (e *CallSubExpr) Labels() []*LabelExpr {
	return []*LabelExpr{e.Label}
}

func (e *CallSubExpr) String() string {
	return fmt.Sprintf("callsub %s", e.Label.Name)
}

var Return = &ReturnExpr{}

type Labelled interface {
	GetLabel() *LabelExpr
}

type AssertExpr struct{}

func (e *AssertExpr) String() string {
	return "assert"
}

var Assert = &AssertExpr{}

type DupExpr struct{}

func (e *DupExpr) String() string {
	return e.Name()
}

func (e *DupExpr) Name() string {
	return "dup"
}

var Dup = &DupExpr{}

type BuryExpr struct {
	Depth uint8
}

func (e *BuryExpr) String() string {
	return fmt.Sprintf("bury %d", e.Depth)
}

type PopNExpr struct {
	Depth uint8
}

func (e *PopNExpr) String() string {
	return fmt.Sprintf("popn %d", e.Depth)
}

type DupNExpr struct {
	Count uint8
}

func (e *DupNExpr) String() string {
	return fmt.Sprintf("dupn %d", e.Count)
}

type ExtractUint16Expr struct{}

func (e *ExtractUint16Expr) String() string {
	return "extract_uint16"
}

var Extract16Bits = &ExtractUint16Expr{}

type ExtractUint32Expr struct{}

func (e *ExtractUint32Expr) String() string {
	return "extract_uint32"
}

var Extract32Bits = &ExtractUint32Expr{}

type Extract64BitsExpr struct{}

func (e *Extract64BitsExpr) String() string {
	return "extract_uint64"
}

var Extract64Bits = &Extract64BitsExpr{}

type Replace2Expr struct {
	Start uint8
}

func (e *Replace2Expr) String() string {
	return fmt.Sprintf("repace2 %d", e.Start)
}

type Replace3Expr struct{}

func (e *Replace3Expr) String() string {
	return "replace3"
}

var Replace3 = &Replace3Expr{}

type Base64DecodeExpr struct {
	Index uint8
}

func (e *Base64DecodeExpr) String() string {
	return fmt.Sprintf("base64_decode %d", e.Index)
}

type JsonRefExpr struct {
	Index uint8
}

func (e *JsonRefExpr) String() string {
	return fmt.Sprintf("json_ref %d", e.Index)
}

type Ed25519VerifyBareExpr struct{}

func (e *Ed25519VerifyBareExpr) String() string {
	return "ed25519verify_bare"
}

var Ed25519VerifyBare = &Ed25519VerifyBareExpr{}

type BitLenExpr struct{}

func (e *BitLenExpr) String() string {
	return "bitlen"
}

var BitLen = &BitLenExpr{}

type ExpwExpr struct{}

func (e *ExpwExpr) String() string {
	return "expw"
}

var Expw = &ExpwExpr{}

type ItxnaExpr struct {
	Field TxnField
	Index uint8
}

func (e *ItxnaExpr) String() string {
	return fmt.Sprintf("itxna %d %d", e.Field, e.Index)
}

type GtxnasExpr struct {
	Field TxnField
	Index uint8
}

func (e *GtxnasExpr) String() string {
	return fmt.Sprintf("gtxnas %d %d", e.Field, e.Index)
}

type ArgsExpr struct{}

func (e *ArgsExpr) String() string {
	return "args"
}

var Args = &ArgsExpr{}

type GloadssExpr struct{}

func (e *GloadssExpr) String() string {
	return "gloadss"
}

var Gloadss = &GloadssExpr{}

type ItxnasExpr struct {
	Field TxnField
}

func (e *ItxnasExpr) String() string {
	return fmt.Sprintf("itxnas %s", e.Field)
}

type GitxnasExpr struct {
	Index uint8
	Field TxnField
}

func (e *GitxnasExpr) String() string {
	return fmt.Sprintf("gitxnas %d %s", e.Index, e.Field)
}

type PushIntsExpr struct {
	Ints []uint64
}

func (e *PushIntsExpr) String() string {
	var ss []string

	for _, i := range e.Ints {
		ss = append(ss, strconv.FormatUint(i, 10))
	}

	return fmt.Sprintf("pushbytess %s", strings.Join(ss, " "))

}

type MethodExpr struct {
	Signature string
}

func (e *MethodExpr) String() string {
	return fmt.Sprintf("method \"%s\"", strings.ReplaceAll(e.Signature, "\"", "\\\""))
}

type AddrExpr struct {
	Address string
}

func (e *AddrExpr) String() string {
	return fmt.Sprintf("addr %s", e.Address)
}

type PushBytessExpr struct {
	Bytess [][]byte
}

func (e *PushBytessExpr) String() string {
	var ss []string

	for _, bs := range e.Bytess {
		ss = append(ss, Bytes{Format: BytesBase64, Value: bs}.String())
	}

	return fmt.Sprintf("pushbytess %s", strings.Join(ss, " "))
}

type EcAddExpr struct {
	Group EcGroup
}

func (e *EcAddExpr) String() string {
	return fmt.Sprintf("ec_add %d", e.Group)
}

type EcScalarMul struct {
	Group EcGroup
}

func (e *EcScalarMul) String() string {
	return fmt.Sprintf("ec_scalar_mul %d", e.Group)
}

type EcMultiScalarMulExpr struct {
	Group EcGroup
}

func (e *EcMultiScalarMulExpr) String() string {
	return fmt.Sprintf("ec_multi_scalar_mul %d", e.Group)
}

type EcPairingCheckExpr struct {
	Group EcGroup
}

func (e *EcPairingCheckExpr) String() string {
	return fmt.Sprintf("ec_pairing_check %d", e.Group)
}

type EcMultiExpExpr struct {
	Group EcGroup
}

func (e *EcMultiExpExpr) String() string {
	return fmt.Sprintf("ec_multi_exp %d", e.Group)
}

type EcSubgroupCheckExpr struct {
	Group EcGroup
}

func (e *EcSubgroupCheckExpr) String() string {
	return fmt.Sprintf("ec_subgroup_check %d", e.Group)
}

type EcMapToExpr struct {
	Group EcGroup
}

func (e *EcMapToExpr) String() string {
	return fmt.Sprintf("ec_map_to %d", e.Group)
}

type FrameBuryExpr struct {
	Index int8
}

func (e *FrameBuryExpr) String() string {
	return fmt.Sprintf("frame_bury %d", e.Index)
}

type FrameDigExpr struct {
	Index int8
}

func (e *FrameDigExpr) String() string {
	return fmt.Sprintf("frame_dig %d", e.Index)
}

type SetByteExpr struct{}

func (e *SetByteExpr) String() string {
	return "setbyte"
}

var SetByte = &SetByteExpr{}

type CoverExpr struct {
	Depth uint8
}

func (e *CoverExpr) String() string {
	return fmt.Sprintf("cover %d", e.Depth)
}

type UncoverExpr struct {
	Depth uint8
}

func (e *UncoverExpr) String() string {
	return fmt.Sprintf("uncover %d", e.Depth)
}

func Uncover(index uint8) *UncoverExpr {
	return &UncoverExpr{Depth: index}
}

type CommentExpr struct {
	Text string
}

func (e *CommentExpr) IsNop() {}

func (e *CommentExpr) String() string {
	return fmt.Sprintf("//%s", e.Text)
}

type AssetParamsGetExpr struct {
	Field AssetParamsField
}

func (e *AssetParamsGetExpr) String() string {
	return fmt.Sprintf("asset_params_get %s", e.Field)
}

type ConcatExpr struct{}

func (e *ConcatExpr) String() string {
	return "concat"
}

var Concat = &ConcatExpr{}

type ItxnBeginExpr struct{}

func (e *ItxnBeginExpr) String() string {
	return "itxn_begin"
}

var ItxnBegin = &ItxnBeginExpr{}

type ItxnSubmitExpr struct{}

func (e *ItxnSubmitExpr) String() string {
	return "itxn_submit"
}

var ItxnSubmit = &ItxnSubmitExpr{}

type ItxnExpr struct {
	Field TxnField
}

func (e *ItxnExpr) String() string {
	return fmt.Sprintf("itxn %s", e.Field)
}

type ItxnFieldExpr struct {
	Field TxnField
}

func (e *ItxnFieldExpr) String() string {
	return fmt.Sprintf("itxn_field %s", e.Field)
}

type PopExpr struct{}

func (e *PopExpr) String() string {
	return "pop"
}

var Pop = &PopExpr{}

type SwapExpr struct{}

func (e *SwapExpr) String() string {
	return "swap"
}

var Swap = &SwapExpr{}

type GtxnaExpr struct {
	Group uint8
	Field TxnField
	Index uint8
}

func (e *GtxnaExpr) String() string {
	return fmt.Sprintf("gtxna %d %s %d", e.Group, e.Field, e.Index)
}

type GtxnsaExpr struct {
	Field TxnField
	Index uint8
}

func (e *GtxnsaExpr) String() string {
	return fmt.Sprintf("gtxnsa %s %d", e.Field, e.Index)
}

type GloadExpr struct {
	Group uint8
	Index uint8
}

func (e *GloadExpr) String() string {
	return fmt.Sprintf("gload %d %d", e.Group, e.Index)
}

type GloadsExpr struct {
	Index uint8
}

func (e *GloadsExpr) String() string {
	return fmt.Sprintf("gloads %d", e.Index)
}

type SqrtExpr struct{}

func (e *SqrtExpr) String() string {
	return "sqrt"
}

var Sqrt = &SqrtExpr{}

type BalanceExpr struct{}

func (e *BalanceExpr) String() string {
	return "balance"
}

var Balance = &BalanceExpr{}

type TxnasExpr struct {
	Field TxnField
}

func (e *TxnasExpr) String() string {
	return fmt.Sprintf("txnas %s", e.Field)
}

type ExtractExpr struct {
	Start  uint8
	Length uint8
}

func (e *ExtractExpr) String() string {
	return fmt.Sprintf("extract %d %d", e.Start, e.Length)
}

type ExpExpr struct{}

func (e *ExpExpr) String() string {
	return "exp"
}

var Exp = &ExpExpr{}

type AppParamsGetExpr struct {
	Field AppParamsField
}

func (e *AppParamsGetExpr) String() string {
	return fmt.Sprintf("app_params_get %s", e.Field)
}

type MimcExpr struct {
	Field MimcConfig
}

func (e *MimcExpr) String() string {
	return fmt.Sprintf("mimc %s", e.Field)
}

type VoterParamsGetExpr struct {
	Field VoterParamsField
}

func (e *VoterParamsGetExpr) String() string {
	return fmt.Sprintf("voter_params_get %s", e.Field)
}

type AcctParamsGetExpr struct {
	Field AcctParamsField
}

func (e *AcctParamsGetExpr) String() string {
	return fmt.Sprintf("acct_params_get %s", e.Field)
}

type LogExpr struct{}

func (e *LogExpr) String() string {
	return "log"
}

var Log = &LogExpr{}

type BlockExpr struct {
	Field BlockField
}

func (e *BlockExpr) String() string {
	return fmt.Sprintf("block %s", e.Field)
}

type VrfVerifyExpr struct {
	Field VrfStandard
}

func (e *VrfVerifyExpr) String() string {
	return fmt.Sprintf("vrf_verify %s", e.Field)
}

type Extract3Expr struct{}

func (e *Extract3Expr) String() string {
	return "extract3"
}

var Extract3 = &Extract3Expr{}

type ExtractUint64Expr struct{}

func (e *ExtractUint64Expr) String() string {
	return "extract_unt64"
}

var ExtractUint64 = &ExtractUint64Expr{}

type AssetHoldingGetExpr struct {
	Field AssetHoldingField
}

func (e *AssetHoldingGetExpr) String() string {
	return fmt.Sprintf("asset_holding_get %s", e.Field)
}

type ItxnNextExpr struct{}

func (e *ItxnNextExpr) String() string {
	return "itxn_next"
}

var ItxnNext = &ItxnNextExpr{}

type MinBalanceExpr struct{}

func (e *MinBalanceExpr) String() string {
	return "min_balance"
}

var MinBalanceOp = &MinBalanceExpr{}

type OnlineStakeExpr struct{}

func (e *OnlineStakeExpr) String() string {
	return "online_stake"
}

var OnlineStakeOp = &OnlineStakeExpr{}

type GitxnExpr struct {
	Index uint8
	Field TxnField
}

func (e *GitxnExpr) String() string {
	return fmt.Sprintf("gitxn %d %s", e.Index, e.Field)
}

type BzeroExpr struct{}

func (e *BzeroExpr) String() string {
	return "bzero"
}

var Bzero = &BzeroExpr{}

type GetByteExpr struct{}

func (e *GetByteExpr) String() string {
	return "getbyte"
}

var GetByte = &GetByteExpr{}

type Substring3Expr struct{}

func (e *Substring3Expr) String() string {
	return "substring3"
}

var Substring3 = &Substring3Expr{}

type ShlExpr struct{}

func (e *ShlExpr) String() string {
	return "shl"
}

var ShiftLeft = &ShlExpr{}

type ShrExpr struct{}

func (e *ShrExpr) String() string {
	return "shr"
}

var ShiftRight = &ShrExpr{}

type GetBitExpr struct{}

func (e *GetBitExpr) String() string {
	return "getbit"
}

var GetBit = &GetBitExpr{}

type SetBitExpr struct{}

func (e *SetBitExpr) String() string {
	return "setbit"
}

var SetBit = &SetBitExpr{}

type GtxnsasExpr struct {
	Field TxnField
}

func (e *GtxnsasExpr) String() string {
	return fmt.Sprintf("gtxnsas %s", e.Field)
}

type DigExpr struct {
	Index uint8
}

func (e *DigExpr) String() string {
	return fmt.Sprintf("dig %d", e.Index)
}

type BminusExpr struct{}

func (e *BminusExpr) String() string {
	return "b-"
}

var BytesMinus = &BminusExpr{}

type BmulExpr struct{}

func (e *BmulExpr) String() string {
	return "b*"
}

var BytesMul = &BmulExpr{}

type BdivExpr struct{}

func (e *BdivExpr) String() string {
	return "b/"
}

var BytesDiv = &BdivExpr{}

type BplusExpr struct{}

func (e *BplusExpr) String() string {
	return "b+"
}

var BytesPlus = &BplusExpr{}

type GaidExpr struct {
	Group uint8
}

func (e *GaidExpr) String() string {
	return fmt.Sprintf("gaid %d", e.Group)
}

type GaidsExpr struct {
}

func (e *GaidsExpr) String() string {
	return "gaids"
}

var Gaids = &GaidsExpr{}

type LoadsExpr struct {
}

func (e *LoadsExpr) String() string {
	return "loads"
}

var Loads = &LoadsExpr{}

type StoresExpr struct {
}

func (e *StoresExpr) String() string {
	return "stores"
}

var Stores = &StoresExpr{}

type Dup2Expr struct{}

func (e *Dup2Expr) String() string {
	return "dup2"
}

var Dup2 = &Dup2Expr{}

type SubstringExpr struct {
	Start uint8
	End   uint8
}

func (e *SubstringExpr) String() string {
	return fmt.Sprintf("substring %d %d", e.Start, e.End)
}

type DivwExpr struct{}

func (e *DivwExpr) String() string {
	return "divw"
}

var Divw = &DivwExpr{}

type SelectExpr struct{}

func (e *SelectExpr) String() string {
	return "select"
}

var Select = &SelectExpr{}

type BgteqExpr struct{}

func (e *BgteqExpr) String() string {
	return "b>="
}

var Bgteq = &BgteqExpr{}

type BGtExpr struct{}

func (e *BGtExpr) String() string {
	return "b>"
}

var BytesGt = &BGtExpr{}

type BytesLeExpr struct{}

func (e *BytesLeExpr) String() string {
	return "b<="
}

var BytesLe = &BytesLeExpr{}

type BytesGeExpr struct{}

func (e *BytesGeExpr) String() string {
	return "b>="
}

var BytesGe = &BytesGeExpr{}

type BytesEqExpr struct{}

func (e *BytesEqExpr) String() string {
	return "b=="
}

var BytesEq = &BytesEqExpr{}

type BytesNeqExpr struct{}

func (e *BytesNeqExpr) String() string {
	return "b!="
}

var BytesNeq = &BytesNeqExpr{}

type BytesModuloExpr struct{}

func (e *BytesModuloExpr) String() string {
	return "b%"
}

var BytesModulo = &BytesModuloExpr{}

type BytesBitOrExpr struct{}

func (e *BytesBitOrExpr) String() string {
	return "b|"
}

var BytesBitOr = &BytesBitOrExpr{}

type BytesBitAndExpr struct{}

func (e *BytesBitAndExpr) String() string {
	return "b&"
}

var BytesBitAnd = &BytesBitAndExpr{}

type BsqrtExpr struct{}

func (e *BsqrtExpr) String() string {
	return "bsqrt"
}

var Bsqrt = &BsqrtExpr{}

type GitxnaExpr struct {
	Group uint8
	Field TxnField
	Index uint8
}

func (e *GitxnaExpr) String() string {
	return fmt.Sprintf("gitxna %d %s %d", e.Group, e.Field, e.Index)
}

type AppGlobalDelExpr struct{}

func (e *AppGlobalDelExpr) String() string {
	return "app_global_del"
}

var AppGlobalDel = &AppGlobalDelExpr{}

type BltExpr struct{}

func (e *BltExpr) String() string {
	return "b<"
}

var BytesLt = &BltExpr{}

type BandExpr struct{}

func (e *BandExpr) String() string {
	return "b&"
}

var Band = &BandExpr{}

type BytesBitXorExpr struct{}

func (e *BytesBitXorExpr) String() string {
	return "b^"
}

var BytesBitXor = &BytesBitXorExpr{}

type BytesBitNotExpr struct{}

func (e *BytesBitNotExpr) String() string {
	return "b~"
}

var BytesBitNot = &BytesBitNotExpr{}

type BytesZeroExpr struct{}

func (e *BytesZeroExpr) String() string {
	return "bzero"
}

var BytesZero = &BytesZeroExpr{}

type AppOptedInExpr struct{}

func (e *AppOptedInExpr) String() string {
	return "app_opted_in"
}

var AppOptedIn = &AppOptedInExpr{}

func Minus(a, b Expr) Expr {
	return Block(
		a, b, MinusOp,
	)
}

func Plus(a, b Expr) Expr {
	return Block(
		a, b, PlusOp,
	)
}

func Load(index uint8) *LoadExpr {
	return &LoadExpr{Index: index}
}

func Store(index uint8) *StoreExpr {
	return &StoreExpr{Index: index}
}

func Gtxns(f TxnField) *GtxnsExpr {
	return &GtxnsExpr{Field: f}
}

func Proto(args uint8, results uint8) *ProtoExpr {
	return &ProtoExpr{Args: args, Results: results}
}

func Bz(l *LabelExpr) *BzExpr {
	return &BzExpr{
		Label: l,
	}
}

func B(l *LabelExpr) *BExpr {
	return &BExpr{
		Label: l,
	}
}

func StringBytes(v string) *ByteExpr {
	return &ByteExpr{Value: []byte(v), Format: BytesStringLiteral}
}

func Label(name string) *LabelExpr {
	return &LabelExpr{Name: name}
}

func Match(labels ...*LabelExpr) *MatchExpr {
	return &MatchExpr{Targets: labels}
}

func Switch(labels ...*LabelExpr) *SwitchExpr {
	return &SwitchExpr{Targets: labels}
}

func Txn(f TxnField) *TxnExpr {
	return &TxnExpr{Field: f}
}

func Txna(f TxnField, i uint8) *TxnaExpr {
	return &TxnaExpr{Field: f, Index: i}
}

func Int(v uint64) *IntExpr {
	return &IntExpr{Value: v}
}

func Bnz(l *LabelExpr) *BnzExpr {
	return &BnzExpr{
		Label: l,
	}
}

func MatchLabels(index Expr, labels ...*LabelExpr) *NestedExpr {
	var body []Expr

	for _, l := range labels {
		body = append(body, StringBytes(l.Name))
	}

	body = append(body, index, Match(labels...))

	return &NestedExpr{Body: body}
}

func AssertEq(a, b Expr) Expr {
	return Block(
		a,
		b,
		Eq,
		Assert,
	)
}

func AssertLen(value Expr, len uint64) Expr {
	return AssertEq(
		Block(
			value,
			Len,
		),
		Int(len),
	)
}

func AssertLtEq(a, b Expr) Expr {
	return Block(
		a,
		b,
		Le,
		Assert,
	)
}

func Block(exprs ...Expr) *NestedExpr {
	return &NestedExpr{Body: exprs}
}

func FrameDig(index int8) *FrameDigExpr {
	return &FrameDigExpr{Index: index}
}

func Arg(index uint8) *TxnaExpr {
	return Txna(ApplicationArgs, index)
}

func ItxnField(field TxnField) *ItxnFieldExpr {
	return &ItxnFieldExpr{Field: field}
}

func Itxn(exprs ...Expr) Expr {
	var r []Expr

	r = append(r, ItxnBegin)
	r = append(r, exprs...)
	r = append(r, ItxnSubmit)

	return Block(r)
}

func CallSub[T Labelled](l T) *CallSubExpr {
	return &CallSubExpr{Label: l.GetLabel()}
}

func Cover(index uint8) *UncoverExpr {
	return &UncoverExpr{Depth: index}
}

func Comment(text string) *CommentExpr {
	return &CommentExpr{
		Text: text,
	}
}

func Global(index GlobalField) *GlobalExpr {
	return &GlobalExpr{Field: index}
}

func AssertInRange(expr Expr, a_inclusive uint64, b_exclusive uint64) Expr {
	return Block(
		expr,
		Int(a_inclusive),
		Ge,
		expr,
		Int(b_exclusive),
		Lt,
		And,
		Assert,
	)
}
