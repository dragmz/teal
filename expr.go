package teal

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

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

// sha256

type Sha256Expr struct{}

func (e *Sha256Expr) String() string {
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
	return "ed25519verify"
}

var ED25519Verify = &ED25519VerifyExpr{}

// ecdsa_verify v

type EcdsaCurve uint8

const (
	EcdsaVerifySecp256k1 = EcdsaCurve(0)
	EcdsaVerifySecp256r1 = EcdsaCurve(1)
)

type EcdsaVerifyExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaVerifyExpr) String() string {
	return fmt.Sprintf("ecdsa_verify %d", e.Index)
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

var LtEq = &LtEqExpr{}

// >=

type GtEqExpr struct {
}

func (e *GtEqExpr) String() string {
	return ">="
}

var GtEq = &GtEqExpr{}

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

var Mod = &ModExpr{}

// |

type BinOrExpr struct{}

func (e *BinOrExpr) String() string {
	return "|"
}

var BinOr = &BinOrExpr{}

// &

type BinAndExpr struct{}

func (e *BinAndExpr) String() string {
	return "&"
}

var BinAnd = &BinAndExpr{}

// ^

type BinXorExpr struct{}

func (e *BinXorExpr) String() string {
	return "^"
}

var BinXor = &BinXorExpr{}

// ~

type BinNotExpr struct{}

func (e *BinNotExpr) String() string {
	return "~"
}

var BinNot = &BinNotExpr{}

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

type TxnField uint8

const (
	TxnSender                    = TxnField(0)
	TxnFee                       = TxnField(1)
	TxnFirstValid                = TxnField(2)
	TxnFirstValidTime            = TxnField(3)
	TxnLastValid                 = TxnField(4)
	TxnNote                      = TxnField(5)
	TxnLease                     = TxnField(6)
	TxnReceiver                  = TxnField(7)
	TxnAmount                    = TxnField(8)
	TxnCloseRemainderTo          = TxnField(9)
	TxnVotePK                    = TxnField(10)
	TxnSelectionPK               = TxnField(11)
	TxnVoteFirst                 = TxnField(12)
	TxnVoteLast                  = TxnField(13)
	TxnVoteKeyDilution           = TxnField(14)
	TxnType                      = TxnField(15)
	TxnTypeEnum                  = TxnField(16)
	TxnXferAsset                 = TxnField(17)
	TxnAssetAmount               = TxnField(18)
	TxnAssetSender               = TxnField(19)
	TxnAssetReceiver             = TxnField(20)
	TxnAssetCloseTo              = TxnField(21)
	TxnGroupIndex                = TxnField(22)
	TxnTxID                      = TxnField(23)
	TxnApplicationID             = TxnField(24)
	TxnOnCompletion              = TxnField(25)
	TxnApplicationArgs           = TxnField(26)
	TxnNumAppArgs                = TxnField(27)
	TxnAccounts                  = TxnField(28)
	TxnNumAccounts               = TxnField(29)
	TxnApprovalProgram           = TxnField(30)
	TxnClearStateProgram         = TxnField(31)
	TxnRekeyTo                   = TxnField(32)
	TxnConfigAsset               = TxnField(33)
	TxnConfigAssetTotal          = TxnField(34)
	TxnConfigAssetDecimals       = TxnField(35)
	TxnConfigAssetDefaultFrozen  = TxnField(36)
	TxnConfigAssetUnitName       = TxnField(37)
	TxnConfigAssetName           = TxnField(38)
	TxnConfigAssetURL            = TxnField(39)
	TxnConfigAssetMetadataHash   = TxnField(40)
	TxnConfigAssetManager        = TxnField(41)
	TxnConfigAssetReserve        = TxnField(42)
	TxnConfigAssetFreeze         = TxnField(43)
	TxnConfigAssetClawback       = TxnField(44)
	TxnFreezeAsset               = TxnField(45)
	TxnFreezeAssetAccounts       = TxnField(46)
	TxnFreezeAssetFrozen         = TxnField(47)
	TxnAssets                    = TxnField(48)
	TxnNumAssets                 = TxnField(49)
	TxnApplications              = TxnField(50)
	TxnNumApplications           = TxnField(51)
	TxnGlobalNumUint             = TxnField(52)
	TxnGlobalNumByteSlice        = TxnField(53)
	TxnLocalNumUint              = TxnField(54)
	TxnLocalNumByteSlice         = TxnField(55)
	TxnExtraProgramPages         = TxnField(56)
	TxnNonparticipation          = TxnField(57)
	TxnLogs                      = TxnField(58)
	TxnNumLogs                   = TxnField(59)
	TxnCreatedAssetID            = TxnField(60)
	TxnCreatedApplicationID      = TxnField(61)
	TxnLastLog                   = TxnField(62)
	TxnStateProofPK              = TxnField(63)
	TxnApprovalProgramPages      = TxnField(64)
	TxnNumApprovalProgramPages   = TxnField(65)
	TxnClearStateProgramPages    = TxnField(66)
	TxnNumClearStateProgramPages = TxnField(67)
)

func (f TxnField) String() string {
	switch f {
	case TxnSender:
		return "Sender"
	case TxnFee:
		return "TxnFee"
	case TxnFirstValid:
		return "FirstValid"
	case TxnFirstValidTime:
		return "FirstValidTime"
	case TxnLastValid:
		return "LastValid"
	case TxnNote:
		return "Note"
	case TxnLease:
		return "Lease"
	case TxnReceiver:
		return "Receiver"
	case TxnAmount:
		return "Amount"
	case TxnCloseRemainderTo:
		return "CloseRemainderTo"
	case TxnVotePK:
		return "VotePK"
	case TxnSelectionPK:
		return "SelectionPK"
	case TxnVoteFirst:
		return "VoteFirst"
	case TxnVoteLast:
		return "VoteLast"
	case TxnVoteKeyDilution:
		return "VoteKeyDilution"
	case TxnType:
		return "Type"
	case TxnTypeEnum:
		return "TypeEnum"
	case TxnXferAsset:
		return "XferAsset"
	case TxnAssetAmount:
		return "AssetAmount"
	case TxnAssetSender:
		return "AssetSender"
	case TxnAssetReceiver:
		return "AssetReceiver"
	case TxnAssetCloseTo:
		return "AssetCloseTo"
	case TxnGroupIndex:
		return "GroupIndex"
	case TxnTxID:
		return "TxID"
	case TxnApplicationID:
		return "ApplicationID"
	case TxnOnCompletion:
		return "OnCompletion"
	case TxnApplicationArgs:
		return "ApplicationArgs"
	case TxnNumAppArgs:
		return "NumAppArgs"
	case TxnAccounts:
		return "Accounts"
	case TxnNumAccounts:
		return "NumAccounts"
	case TxnApprovalProgram:
		return "ApprovalProgram"
	case TxnClearStateProgram:
		return "ClearStateProgram"
	case TxnRekeyTo:
		return "RekeyTo"
	case TxnConfigAsset:
		return "ConfigAsset"
	case TxnConfigAssetTotal:
		return "ConfigAssetTotal"
	case TxnConfigAssetDecimals:
		return "ConfigAssetDecimals"
	case TxnConfigAssetDefaultFrozen:
		return "ConfigAssetDefaultFrozen"
	case TxnConfigAssetUnitName:
		return "ConfigAssetUnitName"
	case TxnConfigAssetName:
		return "ConfigAssetName"
	case TxnConfigAssetURL:
		return "ConfigAssetURL"
	case TxnConfigAssetMetadataHash:
		return "ConfigAssetMetadataHash"
	case TxnConfigAssetManager:
		return "ConfigAssetManager"
	case TxnConfigAssetReserve:
		return "ConfigAssetReserve"
	case TxnConfigAssetFreeze:
		return "ConfigAssetFreeze"
	case TxnConfigAssetClawback:
		return "ConfigAssetClawback"
	case TxnFreezeAsset:
		return "FreezeAsset"
	case TxnFreezeAssetAccounts:
		return "FreezeAssetAccounts"
	case TxnFreezeAssetFrozen:
		return "FreezeAssetFrozen"
	case TxnAssets:
		return "Assets"
	case TxnNumAssets:
		return "NumAssets"
	case TxnApplications:
		return "Applications"
	case TxnNumApplications:
		return "NumApplications"
	case TxnGlobalNumUint:
		return "GlobalNumUint"
	case TxnGlobalNumByteSlice:
		return "GlobalNumByteSlice"
	case TxnLocalNumUint:
		return "LocalNumUint"
	case TxnLocalNumByteSlice:
		return "LocalNumByteSlice"
	case TxnExtraProgramPages:
		return "ExtraProgramPages"
	case TxnNonparticipation:
		return "Nonparticipation"
	case TxnLogs:
		return "Logs"
	case TxnNumLogs:
		return "NumLogs"
	case TxnCreatedAssetID:
		return "CreatedAssetID"
	case TxnCreatedApplicationID:
		return "CreatedApplicationID"
	case TxnLastLog:
		return "LastLog"
	case TxnStateProofPK:
		return "StateProofPK"
	case TxnApprovalProgramPages:
		return "ApprovalProgramPages"
	case TxnNumApprovalProgramPages:
		return "NumApprovalProgramPages"
	case TxnClearStateProgramPages:
		return "ClearStateProgramPages"
	case TxnNumClearStateProgramPages:
		return "NumClearStateProgramPages"
	default:
		return fmt.Sprintf("%d", f)
	}
}

type TxnExpr struct {
	Field TxnField
}

func (e *TxnExpr) String() string {
	return fmt.Sprintf("txn %s", e.Field)
}

// global f

type GlobalField uint8

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
)

func (f GlobalField) String() string {
	switch f {
	case GlobalMinTxnFee:
		return "MinTxnFee"
	case GlobalMinBalance:
		return "MinBalance"
	case GlobalMaxTxnLife:
		return "MaxTxnLife"
	case GlobalZeroAddress:
		return "ZeroAddress"
	case GlobalGroupSize:
		return "GroupSize"
	case GlobalLogicSigVersion:
		return "LogicSigVersion"
	case GlobalRound:
		return "Round"
	case GlobalLatestTimestamp:
		return "LatestTimestamp"
	case GlobalCurrentApplicationID:
		return "CurrentApplicationID"
	case GlobalCreatorAddress:
		return "CreatorAddress"
	case GlobalCurrentApplicationAddress:
		return "CurrentApplicationAddress"
	case GlobalGroupID:
		return "GroupID"
	case GlobalOpcodeBudget:
		return "OpcodeBudget"
	case GlobalCallerApplicationID:
		return "CallerApplicationID"
	case GlobalCallerApplicationAddress:
		return "CallerApplicationAddress"
	default:
		return fmt.Sprintf("%d", f)
	}
}

type GlobalExpr struct {
	Index GlobalField
}

func (e *GlobalExpr) String() string {
	return fmt.Sprintf("global %s", e.Index)
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
	Index uint8
	Field TxnField
}

func (e *GtxnExpr) String() string {
	return fmt.Sprintf("gtxn %d %s", e.Index, e.Field)
}

// gtxna t f i TODO

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

type compiledLabelExpr struct {
	Name string
}

func (e *compiledLabelExpr) String() string {
	return fmt.Sprintf("%s:", e.Name)
}

type LabelExpr struct {
	Name string
}

func (e *LabelExpr) Compile(c *compiler) []Op {
	return []Op{c.getLabel(e.Name)}
}

func (e *LabelExpr) GetLabel() *LabelExpr {
	return e
}

type NestedExpr struct {
	Label *LabelExpr
	Body  []Expr
}

func (e *NestedExpr) Compile(c *compiler) []Op {
	var ops []Op

	if e.Label != nil {
		ops = c.compile(ops, c.getLabel(e.Label.Name))
	}

	if e.Body != nil {
		ops = c.compile(ops, e.Body)
	}

	return ops
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

func (e *FuncExpr) Compile(c *compiler) []Op {
	var ops []Op

	if e.Label != nil {
		ops = c.compile(ops, c.getLabel(e.Label.Name))
	}

	if e.Proto != nil {
		ops = c.compile(ops, e.Proto)
	}

	if e.Block != nil {
		ops = c.compile(ops, e.Block)
	}

	ops = append(ops, RetSub)

	return ops
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
	return Bytes{Value: e.Value, Format: e.Format}.String()
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
		return fmt.Sprintf("byte \"%s\"", string(e.Value))
	case BytesBase64:
		return fmt.Sprintf("byte b64 %s", base64.StdEncoding.EncodeToString(e.Value))
	default:
		panic(fmt.Sprintf("unsupported bytes format: %d", e.Format))
	}

}

func (e *ByteExpr) String() string {
	return Bytes{
		Value:  e.Value,
		Format: e.Format,
	}.String()
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

type PragmaExpr struct {
	Version uint8
}

func (e *PragmaExpr) String() string {
	return fmt.Sprintf("#pragma version %d", e.Version)
}

type BnzExpr struct {
	Label *LabelExpr
}

type compiledBnzExpr struct {
	Label *compiledLabelExpr
}

func (e *compiledBnzExpr) IsBranch() {}

func (e *BnzExpr) Compile(c *compiler) []Op {
	return []Op{&compiledBnzExpr{Label: c.getLabel(e.Label.Name)}}
}

func (e *compiledBnzExpr) Labels() []*compiledLabelExpr {
	return []*compiledLabelExpr{e.Label}
}

func (e *compiledBnzExpr) String() string {
	return fmt.Sprintf("bnz %s", e.Label.Name)
}

type BzExpr struct {
	Label *LabelExpr
}

type compiledBzExpr struct {
	Label *compiledLabelExpr
}

func (e *compiledBzExpr) IsBranch() {}

func (e *compiledBzExpr) Labels() []*compiledLabelExpr {
	return []*compiledLabelExpr{e.Label}
}

func (e *BzExpr) Compile(c *compiler) []Op {
	return []Op{&compiledBzExpr{Label: c.getLabel(e.Label.Name)}}
}

func (e *compiledBzExpr) String() string {
	return fmt.Sprintf("bz %s", e.Label.Name)
}

type BExpr struct {
	Label *LabelExpr
}

type compiledBExpr struct {
	Label *compiledLabelExpr
}

func (e *compiledBExpr) IsBranch() {}

func (e *compiledBExpr) Labels() []*compiledLabelExpr {
	return []*compiledLabelExpr{e.Label}
}

func (e *BExpr) Compile(c *compiler) []Op {
	return []Op{&compiledBExpr{Label: c.getLabel(e.Label.Name)}}
}

func (e *compiledBExpr) String() string {
	return fmt.Sprintf("b %s", e.Label.Name)
}

type Branch interface {
	IsBranch()
}

type Nop interface {
	IsNop()
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
	return "app_local_del_expr"
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
	Labels []*LabelExpr
}

type compiledSwitchExpr struct {
	labels []*compiledLabelExpr
}

func (e *compiledSwitchExpr) IsBranch() {}

func (e *compiledSwitchExpr) Labels() []*compiledLabelExpr {
	return e.labels
}

func (e *SwitchExpr) Compile(c *compiler) []Op {
	labels := make([]*compiledLabelExpr, len(e.Labels))

	for i, l := range e.Labels {
		labels[i] = c.getLabel(l.Name)
	}

	return []Op{&compiledSwitchExpr{labels: labels}}
}

func (e *compiledSwitchExpr) String() string {
	var b strings.Builder

	b.WriteString("switch")

	if len(e.labels) > 0 {
		b.WriteString(" ")

		names := make([]string, len(e.labels))
		for i, l := range e.labels {
			names[i] = l.Name
		}

		b.WriteString(strings.Join(names, " "))
	}

	b.WriteString("")

	return b.String()
}

type MatchExpr struct {
	Labels []*LabelExpr
}

type compiledMatchExpr struct {
	labels []*compiledLabelExpr
}

func (e *compiledMatchExpr) IsBranch() {}

func (e *compiledMatchExpr) Labels() []*compiledLabelExpr {
	return e.labels
}

func (e *MatchExpr) Compile(c *compiler) []Op {
	labels := make([]*compiledLabelExpr, len(e.Labels))

	for i, l := range e.Labels {
		labels[i] = c.getLabel(l.Name)
	}

	return []Op{&compiledMatchExpr{labels}}
}

func (e *compiledMatchExpr) String() string {
	var b strings.Builder

	b.WriteString("match")

	if len(e.labels) > 0 {
		b.WriteString(" ")

		names := make([]string, len(e.labels))
		for i, l := range e.labels {
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

type compiledCallSubExpr struct {
	Label *compiledLabelExpr
}

func (e *compiledCallSubExpr) IsBranch() {}

func (e *compiledCallSubExpr) Labels() []*compiledLabelExpr {
	return []*compiledLabelExpr{e.Label}
}

func (e *CallSubExpr) Compile(c *compiler) []Op {
	return []Op{&compiledCallSubExpr{Label: c.getLabel(e.Label.Name)}}
}

func (e *compiledCallSubExpr) String() string {
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
	return "dup"
}

var Dup = &DupExpr{}

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
	Index uint8
}

func (e *CoverExpr) String() string {
	return fmt.Sprintf("cover %d", e.Index)
}

type UncoverExpr struct {
	Index uint8
}

func (e *UncoverExpr) String() string {
	return fmt.Sprintf("uncover %d", e.Index)
}

func Uncover(index uint8) *UncoverExpr {
	return &UncoverExpr{Index: index}
}

type CommentExpr struct {
	Text string
}

func (e *CommentExpr) IsNop() {}

func (e *CommentExpr) String() string {
	return fmt.Sprintf("//%s", e.Text)
}

type AssetParamsField uint8

const (
	AssetTotal         = AssetParamsField(0)
	AssetDecimals      = AssetParamsField(1)
	AssetDefaultFrozen = AssetParamsField(2)
	AssetUnitName      = AssetParamsField(3)
	AssetName          = AssetParamsField(4)
	AssetURL           = AssetParamsField(5)
	AssetMetadataHash  = AssetParamsField(6)
	AssetManager       = AssetParamsField(7)
	AssetReserve       = AssetParamsField(8)
	AssetFreeze        = AssetParamsField(9)
	AssetClawback      = AssetParamsField(10)
	AssetCreator       = AssetParamsField(11)
)

func (f AssetParamsField) String() string {
	switch f {
	case AssetTotal:
		return "AssetTotal"
	case AssetUnitName:
		return "AssetUnitName"
	case AssetName:
		return "AssetName"
	case AssetURL:
		return "AssetURL"
	case AssetMetadataHash:
		return "AssetMetadataHash"
	case AssetManager:
		return "AssetManager"
	case AssetReserve:
		return "AssetReserve"
	case AssetFreeze:
		return "AssetFreeze"
	case AssetClawback:
		return "AssetClawback"
	case AssetCreator:
		return "AssetCreator"
	case AssetDecimals:
		return "AssetDecimals"
	case AssetDefaultFrozen:
		return "AssetDefaultFrozen"
	default:
		return fmt.Sprintf("%d", f)
	}
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

type AppParamsField uint8

const (
	AppApprovalProgram    = AppParamsField(0)
	AppClearStateProgram  = AppParamsField(1)
	AppGlobalNumUint      = AppParamsField(2)
	AppGlobalNumByteSlice = AppParamsField(3)
	AppLocalNumUint       = AppParamsField(4)
	AppLocalNumByteSlice  = AppParamsField(5)
	AppExtraProgramPages  = AppParamsField(6)
	AppCreator            = AppParamsField(7)
	AppAddress            = AppParamsField(8)
)

func (f AppParamsField) String() string {
	switch f {
	case AppAddress:
		return "AppAddress"
	case AppApprovalProgram:
		return "AppApprovalProgram"
	case AppClearStateProgram:
		return "AppClearStateProgram"
	case AppCreator:
		return "AppCreator"
	case AppExtraProgramPages:
		return "AppExtraProgramPages"
	case AppGlobalNumByteSlice:
		return "AppGlobalNumByteSlice"
	case AppGlobalNumUint:
		return "AppGlobalNumUint"
	case AppLocalNumByteSlice:
		return "AppLocalNumByteSlice"
	case AppLocalNumUint:
		return "AppLocalNumUint"
	default:
		return fmt.Sprintf("%d", f)
	}
}

type AppParamsGetExpr struct {
	Field AppParamsField
}

func (e *AppParamsGetExpr) String() string {
	return fmt.Sprintf("app_params_get %s", e.Field)
}

type AcctParamsField uint8

const (
	AcctBalance            = AcctParamsField(0)
	AcctMinBalance         = AcctParamsField(1)
	AcctAuthAddr           = AcctParamsField(2)
	AcctTotalNumUint       = AcctParamsField(3)
	AcctTotalNumByteSlice  = AcctParamsField(4)
	AcctTotalExtraAppPages = AcctParamsField(5)
	AcctTotalAppsCreated   = AcctParamsField(6)
	AcctTotalAppsOptedIn   = AcctParamsField(7)
	AcctTotalAssetsCreated = AcctParamsField(8)
	AcctTotalAssets        = AcctParamsField(9)
	AcctTotalBoxes         = AcctParamsField(10)
	AcctTotalBoxBytes      = AcctParamsField(11)
)

func (f AcctParamsField) String() string {
	switch f {
	case AcctBalance:
		return "AcctBalance"
	case AcctMinBalance:
		return "AcctMinBalance"
	case AcctAuthAddr:
		return "AcctAuthAddr"
	case AcctTotalNumUint:
		return "AcctTotalNumUint"
	case AcctTotalNumByteSlice:
		return "AcctTotalNumByteSlice"
	case AcctTotalExtraAppPages:
		return "AcctTotalExtraAppPages"
	case AcctTotalAppsCreated:
		return "AcctTotalAppsCreated"
	case AcctTotalAppsOptedIn:
		return "AcctTotalAppsOptedIn"
	case AcctTotalAssetsCreated:
		return "AcctTotalAssetsCreated"
	case AcctTotalAssets:
		return "AcctTotalAssets"
	case AcctTotalBoxes:
		return "AcctTotalBoxes"
	case AcctTotalBoxBytes:
		return "AcctTotalBoxBytes"
	default:
		return fmt.Sprintf("%d", f)
	}
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

type BlockField uint8

const (
	BlkSeed      = BlockField(0)
	BlkTimestamp = BlockField(1)
)

func (f BlockField) String() string {
	switch f {
	case BlkSeed:
		return "BlkSeed"
	case BlkTimestamp:
		return "BlkTimestamp"
	default:
		return fmt.Sprintf("%d", f)
	}
}

type BlockExpr struct {
	Field BlockField
}

func (e *BlockExpr) String() string {
	return fmt.Sprintf("block %s", e.Field)
}

type VrfVerifyField uint8

const (
	VrfAlgorand = VrfVerifyField(0)
)

func (f VrfVerifyField) String() string {
	switch f {
	case VrfAlgorand:
		return "VrfAglorand"
	default:
		return fmt.Sprintf("%d", f)
	}
}

type VrfVerifyExpr struct {
	Field VrfVerifyField
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

type AssetHoldingField uint8

const (
	AssetBalance = AssetHoldingField(0)
	AssetFrozen  = AssetHoldingField(1)
)

func (f AssetHoldingField) String() string {
	switch f {
	case AssetBalance:
		return "AssetBalance"
	case AssetFrozen:
		return "AssetFrozen"
	default:
		return fmt.Sprintf("%d", f)
	}
}

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

var MinBalance = &MinBalanceExpr{}

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

type ShrExpr struct{}

func (e *ShrExpr) String() string {
	return "shr"
}

var Shr = &ShrExpr{}

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

var Bminus = &BminusExpr{}

type BmulExpr struct{}

func (e *BmulExpr) String() string {
	return "b*"
}

var Bmul = &BmulExpr{}

type BdivExpr struct{}

func (e *BdivExpr) String() string {
	return "b/"
}

var Bdiv = &BdivExpr{}

type BplusExpr struct{}

func (e *BplusExpr) String() string {
	return "b+"
}

var Bplus = &BplusExpr{}

type GaidExpr struct {
	Group uint8
}

func (e *GaidExpr) String() string {
	return fmt.Sprintf("gaid %d", e.Group)
}

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

var Blt = &BltExpr{}

type BandExpr struct{}

func (e *BandExpr) String() string {
	return "b&"
}

var Band = &BandExpr{}

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
	return &MatchExpr{Labels: labels}
}

func Switch(labels ...*LabelExpr) *SwitchExpr {
	return &SwitchExpr{Labels: labels}
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
		LtEq,
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
	return Txna(TxnApplicationArgs, index)
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
	return &UncoverExpr{Index: index}
}

func Comment(text string) *CommentExpr {
	return &CommentExpr{
		Text: text,
	}
}

func Global(index GlobalField) *GlobalExpr {
	return &GlobalExpr{Index: index}
}

func AssertInRange(expr Expr, a_inclusive uint64, b_exclusive uint64) Expr {
	return Block(
		expr,
		Int(a_inclusive),
		GtEq,
		expr,
		Int(b_exclusive),
		Lt,
		And,
		Assert,
	)
}
