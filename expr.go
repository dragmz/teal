package teal

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Branch interface {
	IsBranch()
}

type Nop interface {
	IsNop()
}

// err

type ErrExpr struct{}

func (e *ErrExpr) Execute(b *VmBranch) error {
	b.exit()
	return nil
}

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

func (e *Sha3256Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Sha3256 = &Sha3256Expr{}

// sha256

type Sha256Expr struct{}

func (e *Sha256Expr) String() string {
	return e.Name()
}

func (e *Sha256Expr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes, src: vmOpSource{e: e, args: []vmSource{v}}})

	b.Line++
	return nil
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

func (e *Keccak256Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Keccak256 = &Keccak256Expr{}

// sha512_256

type Sha512256Expr struct{}

func (e *Sha512256Expr) String() string {
	return "sha512_256"
}

func (e *Sha512256Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Sha512256 = &Sha512256Expr{}

// ed25519verify

type ED25519VerifyExpr struct{}

func (e *ED25519VerifyExpr) String() string {
	return e.Name()
}

func (e *ED25519VerifyExpr) Execute(b *VmBranch) error {
	v1 := b.pop(VmTypeBytes)
	v2 := b.pop(VmTypeBytes)
	v3 := b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64, src: vmOpSource{e: e, args: []vmSource{
		v1, v2, v3,
	}}})

	b.Line++
	return nil
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

func (e *EcdsaVerifyExpr) Execute(b *VmBranch) error {
	v1 := b.pop(VmTypeBytes)
	v2 := b.pop(VmTypeBytes)
	v3 := b.pop(VmTypeBytes)
	v4 := b.pop(VmTypeBytes)
	v5 := b.pop(VmTypeBytes)

	b.push(VmValue{T: VmTypeUint64, src: vmOpSource{e: e, args: []vmSource{v1, v2, v3, v4, v5}}})

	b.Line++
	return nil
}

// ecdsa_pk_decompress

type EcdsaPkDecompressExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaPkDecompressExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

func (e *EcdsaPkDecompressExpr) String() string {
	return fmt.Sprintf("ecdsa_pk_decompress %d", e.Index)
}

// ecdsa_pk_recover

type EcdsaPkRecoverExpr struct {
	Index EcdsaCurve
}

func (e *EcdsaPkRecoverExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
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

func (e *PlusExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var PlusOp = &PlusExpr{}

// -

type MinusExpr struct {
}

func (e *MinusExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
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

func (e *DivExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Div = &DivExpr{}

// *

type MulExpr struct{}

func (e *MulExpr) String() string {
	return "*"
}

func (e *MulExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Mul = &MulExpr{}

// <

type LtExpr struct {
}

func (e *LtExpr) String() string {
	return "<"
}

func (e *LtExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Lt = &LtExpr{}

// >

type GtExpr struct {
}

func (e *GtExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

func (e *GtExpr) String() string {
	return ">"
}

var Gt = &GtExpr{}

// <=

type LtEqExpr struct {
}

func (e *LtEqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
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

func (e *GtEqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Ge = &GtEqExpr{}

// &&

type AndExpr struct {
}

func (e *AndExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
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

func (e *OrExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Or = &OrExpr{}

// ==

type EqExpr struct {
}

func (e *EqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

func (e *EqExpr) String() string {
	return "=="
}

var Eq = &EqExpr{}

// !=

type NeqExpr struct {
}

func (e *NeqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
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

func (e *NotExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Not = &NotExpr{}

// len

type LenExpr struct{}

func (e *LenExpr) String() string {
	return "len"
}

func (e *LenExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Len = &LenExpr{}

// itob

type ItobExpr struct{}

func (e *ItobExpr) String() string {
	return "itob"
}

func (e *ItobExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Itob = &ItobExpr{}

// btoi

type BtoiExpr struct{}

func (e *BtoiExpr) String() string {
	return "btoi"
}

func (e *BtoiExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Btoi = &BtoiExpr{}

// %

type ModExpr struct{}

func (e *ModExpr) String() string {
	return "%"
}

func (e *ModExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Modulo = &ModExpr{}

// |

type BitOrExpr struct{}

func (e *BitOrExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

func (e *BitOrExpr) String() string {
	return "|"
}

var BitOr = &BitOrExpr{}

// &

type BitAndExpr struct{}

func (e *BitAndExpr) String() string {
	return "&"
}

func (e *BitAndExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var BitAnd = &BitAndExpr{}

// ^

type BitXorExpr struct{}

func (e *BitXorExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

func (e *BitXorExpr) String() string {
	return "^"
}

var BitXor = &BitXorExpr{}

// ~

type BitNotExpr struct{}

func (e *BitNotExpr) String() string {
	return "~"
}

func (e *BitNotExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var BitNot = &BitNotExpr{}

// mulw

type MulwExpr struct{}

func (e *MulwExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

func (e *MulwExpr) String() string {
	return "mulw"
}

var Mulw = &MulwExpr{}

// addw

type AddwExpr struct{}

func (e *AddwExpr) String() string {
	return "addw"
}

func (e *AddwExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Addw = &AddwExpr{}

// divmodw

type DivModwExpr struct{}

func (e *DivModwExpr) String() string {
	return "divmodw"
}

var DivModw = &DivModwExpr{}

func (e *DivModwExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

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

func (e *IntcExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

// intc_0

type Intc0Expr struct{}

func (e *Intc0Expr) String() string {
	return "intc_0"
}

func (e *Intc0Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Intc0 = &Intc0Expr{}

// intc_1

type Intc1Expr struct{}

func (e *Intc1Expr) String() string {
	return "intc_1"
}
func (e *Intc1Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Intc1 = &Intc1Expr{}

// intc_2

type Intc2Expr struct{}

func (e *Intc2Expr) String() string {
	return "intc_2"
}
func (e *Intc2Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Intc2 = &Intc2Expr{}

// intc_3

type Intc3Expr struct{}

func (e *Intc3Expr) String() string {
	return "intc_3"
}
func (e *Intc3Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
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

func (e *BytecExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

func (e *BytecExpr) String() string {
	return fmt.Sprintf("bytec %d", e.Index)
}

// bytec_0

type Bytec0Expr struct{}

func (e *Bytec0Expr) String() string {
	return "bytec_0"
}

func (e *Bytec0Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Bytec0 = &Bytec0Expr{}

// bytec_1

type Bytec1Expr struct{}

func (e *Bytec1Expr) String() string {
	return "bytec_1"
}

func (e *Bytec1Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Bytec1 = &Bytec1Expr{}

// bytec_2

type Bytec2Expr struct{}

func (e *Bytec2Expr) String() string {
	return "bytec_2"
}
func (e *Bytec2Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Bytec2 = &Bytec2Expr{}

// bytec_3

type Bytec3Expr struct{}

func (e *Bytec3Expr) String() string {
	return "bytec_3"
}
func (e *Bytec3Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Bytec3 = &Bytec3Expr{}

// arg n

type ArgExpr struct {
	Index uint8
}

func (e *ArgExpr) String() string {
	return fmt.Sprintf("arg %d", e.Index)
}

func (e *ArgExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

// arg_0

type Arg0Expr struct{}

func (e *Arg0Expr) String() string {
	return "arg_0"
}
func (e *Arg0Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Arg0 = &Arg0Expr{}

// arg_1

type Arg1Expr struct{}

func (e *Arg1Expr) String() string {
	return "arg_1"
}
func (e *Arg1Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}

var Arg1 = &Arg1Expr{}

// arg_2

type Arg2Expr struct{}

func (e *Arg2Expr) String() string {
	return "arg_2"
}
func (e *Arg2Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
}
func (e *Arg3Expr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})

	b.Line++
	return nil
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

func (e *TxnExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
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
)

type GlobalExpr struct {
	Field GlobalField
}

func (e *GlobalExpr) String() string {
	return fmt.Sprintf("global %s", e.Field)
}

func (e *GlobalExpr) Execute(b *VmBranch) error {
	spec, ok := globalFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

// load i

type LoadExpr struct {
	Index uint8
}

func (e *LoadExpr) String() string {
	return fmt.Sprintf("load %d", e.Index)
}

func (e *LoadExpr) Execute(b *VmBranch) error {
	v := b.vm.Scratch.Items[e.Index]
	b.push(VmValue{T: VmTypeAny, src: v})

	b.Line++
	return nil
}

// store i

type StoreExpr struct {
	Index uint8
}

func (e *StoreExpr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeAny)

	b.store(VmValue{T: VmTypeUint64, src: vmUint64Const{uint64(e.Index)}}, v)

	b.Line++
	return nil
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

func (e *TxnaExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

type GtxnExpr struct {
	Group uint8
	Field TxnField
}

func (e *GtxnExpr) String() string {
	return fmt.Sprintf("gtxn %d %s", e.Group, e.Field)
}

func (e *GtxnExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

// gtxns f

type GtxnsExpr struct {
	Field TxnField
}

func (e *GtxnsExpr) String() string {
	return fmt.Sprintf("gtxns %s", e.Field)
}

func (e *GtxnsExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.pop(VmTypeUint64)
	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

type ProtoExpr struct {
	Args    uint8
	Results uint8
}

func (e *ProtoExpr) Execute(b *VmBranch) error {
	b.prepare(e.Args, e.Results)
	b.Line++
	return nil
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

func (e *PushBytesExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type PushIntExpr struct {
	Value uint64
}

func (e *PushIntExpr) String() string {
	return fmt.Sprintf("pushint %d", e.Value)
}

func (e *PushIntExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type IntExpr struct {
	Value uint64
}

func (e *IntExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64, src: vmConst{v: strconv.FormatUint(e.Value, 10)}})
	b.Line++

	return nil
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

func (e *ByteExpr) Execute(b *VmBranch) error {
	b.push(VmValue{
		T: VmTypeBytes,
		src: vmConst{
			v: Bytes{Value: e.Value, Format: BytesStringLiteral}.String(),
		}})
	b.Line++
	return nil
}

type BoxGetExpr struct {
}

func (e *BoxGetExpr) String() string {
	return "box_get"
}

func (e *BoxGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BoxGet = &BoxGetExpr{}

type BoxPutExpr struct {
}

func (e *BoxPutExpr) String() string {
	return "box_put"
}

func (e *BoxPutExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.Line++
	return nil
}

var BoxPut = &BoxPutExpr{}

type BoxCreateExpr struct {
}

func (e *BoxCreateExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type BoxLenExpr struct {
}

func (e *BoxLenExpr) String() string {
	return "box_len"
}

func (e *BoxLenExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BoxDel = &BoxDelExpr{}

type BoxDelExpr struct {
}

func (e *BoxDelExpr) String() string {
	return "box_del"
}

func (e *BoxDelExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
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

func (e *BoxReplaceExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.Line++
	return nil
}

type BoxExtractExpr struct {
}

var BoxExtract = &BoxExtractExpr{}

func (e *BoxExtractExpr) String() string {
	return "box_extract"
}

func (e *BoxExtractExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type PragmaExpr struct {
	Version uint8
}

func (e *PragmaExpr) IsNop() {}

func (e *PragmaExpr) String() string {
	return fmt.Sprintf("#pragma version %d", e.Version)
}

type BnzExpr struct {
	Label *LabelExpr
}

func (e *BnzExpr) IsBranch() {}

func (e *BnzExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.fork(e.Label.Name)
	b.Line++
	return nil
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

func (e *BzExpr) IsBranch() {}

func (e *BzExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.fork(e.Label.Name)
	b.Line++
	return nil
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

func (e *BExpr) IsBranch() {}

func (e *BExpr) Execute(b *VmBranch) error {
	b.jump(e.Label.Name)
	return nil
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

func (e *AppLocalGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

var AppLocalGet = &AppLocalGetExpr{}

type AppLocalPutExpr struct{}

func (e *AppLocalPutExpr) String() string {
	return "app_local_put"
}

func (e *AppLocalPutExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.pop(VmTypeBytes)
	b.pop(VmTypeAny)
	b.Line++
	return nil
}

var AppLocalPut = &AppLocalPutExpr{}

type AppGlobalPutExpr struct{}

func (e *AppGlobalPutExpr) String() string {
	return "app_global_put"
}

func (e *AppGlobalPutExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.pop(VmTypeBytes)
	b.Line++
	return nil
}

var AppGlobalPut = &AppGlobalPutExpr{}

type AppGlobalGetExpr struct{}

func (e *AppGlobalGetExpr) String() string {
	return "app_global_get"
}

func (e *AppGlobalGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

var AppGlobalGet = &AppGlobalGetExpr{}

type AppGlobalGetExExpr struct{}

func (e *AppGlobalGetExExpr) String() string {
	return "app_global_get_ex"
}

func (e *AppGlobalGetExExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var AppGlobalGetEx = &AppGlobalGetExExpr{}

type AppLocalGetExExpr struct{}

func (e *AppLocalGetExExpr) String() string {
	return "app_local_get_ex"
}

func (e *AppLocalGetExExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeUint64)
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var AppLocalGetEx = &AppLocalGetExExpr{}

type AppLocalDelExpr struct{}

func (e *AppLocalDelExpr) String() string {
	return "app_local_del"
}

func (e *AppLocalDelExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeAny)
	b.Line++
	return nil
}

var AppLocalDel = &AppLocalDelExpr{}

type RetSubExpr struct{}

var RetSub = &RetSubExpr{}

func (e *RetSubExpr) Execute(b *VmBranch) error {
	if len(b.Frames) > 0 {
		f := b.Frames[len(b.Frames)-1]

		rs := []VmValue{}

		for i := uint8(0); i < f.NumReturns; i++ {
			rs = append(rs, b.pop(VmTypeAny))
		}

		for i := uint8(0); i < f.NumArgs; i++ {
			b.pop(VmTypeAny)
		}

		for i := len(rs) - 1; i >= 0; i-- {
			b.push(rs[i])
		}

		b.Line = f.Return + 1
		b.Name = f.Name

		b.Frames = b.Frames[:len(b.Frames)-1]
	} else {
		b.Line++
	}

	return nil
}

func (e *RetSubExpr) String() string {
	return "retsub"
}

func (e *RetSubExpr) IsTerminator() {}

type ReturnExpr struct {
}

func (e *ReturnExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.exit()
	return nil
}

func (e *ReturnExpr) String() string {
	return "return"
}

func (e *ReturnExpr) IsTerminator() {}

type SwitchExpr struct {
	Targets []*LabelExpr
}

func (e *SwitchExpr) IsBranch() {}

func (e *SwitchExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)

	for _, t := range e.Targets {
		b.fork(t.Name)
	}

	b.Line++

	return nil
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

func (e *MatchExpr) IsBranch() {}

func (e *MatchExpr) Execute(b *VmBranch) error {
	for range e.Targets {
		b.pop(VmTypeAny)
	}

	b.pop(VmTypeAny)

	for _, t := range e.Targets {
		b.fork(t.Name)
	}

	b.Line++

	return nil
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

func (e *CallSubExpr) IsBranch() {}

func (e *CallSubExpr) Execute(b *VmBranch) error {
	b.call(e.Label.Name)
	return nil
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

func (e *AssertExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.Line++
	return nil
}

var Assert = &AssertExpr{}

type DupExpr struct{}

func (e *DupExpr) String() string {
	return e.Name()
}

func (e *DupExpr) Name() string {
	return "dup"
}

func (e *DupExpr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeAny)
	b.push(v)
	b.push(v)
	b.Line++
	return nil
}

var Dup = &DupExpr{}

type BuryExpr struct {
	Depth uint8
}

func (e *BuryExpr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeAny)
	b.replace(e.Depth, v)
	b.Line++
	return nil
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

func (e *PopNExpr) Execute(b *VmBranch) error {
	for i := uint8(0); i < e.Depth; i++ {
		b.pop(VmTypeAny)
	}
	b.Line++
	return nil
}

type DupNExpr struct {
	Count uint8
}

func (e *DupNExpr) String() string {
	return fmt.Sprintf("dupn %d", e.Count)
}

func (e *DupNExpr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeAny)
	for i := uint8(0); i <= e.Count; i++ {
		b.push(v)
	}
	b.Line++
	return nil
}

type ExtractUint16Expr struct{}

func (e *ExtractUint16Expr) String() string {
	return "extract_uint16"
}

func (e *ExtractUint16Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Extract16Bits = &ExtractUint16Expr{}

type ExtractUint32Expr struct{}

func (e *ExtractUint32Expr) String() string {
	return "extract_uint32"
}

func (e *ExtractUint32Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Extract32Bits = &ExtractUint32Expr{}

type Extract64BitsExpr struct{}

func (e *Extract64BitsExpr) String() string {
	return "extract_uint64"
}

func (e *ExtractUint64Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Extract64Bits = &Extract64BitsExpr{}

type Replace2Expr struct {
	Start uint8
}

func (e *Replace2Expr) String() string {
	return fmt.Sprintf("repace2 %d", e.Start)
}

func (e *Replace2Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type Replace3Expr struct{}

func (e *Replace3Expr) String() string {
	return "replace3"
}

func (e *Replace3Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Replace3 = &Replace3Expr{}

type Base64DecodeExpr struct {
	Index uint8
}

func (e *Base64DecodeExpr) String() string {
	return fmt.Sprintf("base64_decode %d", e.Index)
}

func (e *Base64DecodeExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type JsonRefExpr struct {
	Index uint8
}

func (e *JsonRefExpr) String() string {
	return fmt.Sprintf("json_ref %d", e.Index)
}

func (e *JsonRefExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

type Ed25519VerifyBareExpr struct{}

func (e *Ed25519VerifyBareExpr) String() string {
	return "ed25519verify_bare"
}

func (e *Ed25519VerifyBareExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Ed25519VerifyBare = &Ed25519VerifyBareExpr{}

type BitLenExpr struct{}

func (e *BitLenExpr) String() string {
	return "bitlen"
}

func (e *BitLenExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++

	return nil
}

var BitLen = &BitLenExpr{}

type ExpwExpr struct{}

func (e *ExpwExpr) String() string {
	return "expw"
}

func (e *ExpwExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++

	return nil
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

func (e *GtxnasExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

func (e *GtxnasExpr) String() string {
	return fmt.Sprintf("gtxnas %d %d", e.Field, e.Index)
}

type ArgsExpr struct{}

func (e *ArgsExpr) String() string {
	return "args"
}

func (e *ArgsExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Args = &ArgsExpr{}

type GloadssExpr struct{}

func (e *GloadssExpr) String() string {
	return "gloadss"
}

func (e *GloadssExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

var Gloadss = &GloadssExpr{}

type ItxnasExpr struct {
	Field TxnField
}

func (e *ItxnasExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

func (e *ItxnasExpr) String() string {
	return fmt.Sprintf("itxnas %s", e.Field)
}

type GitxnasExpr struct {
	Index uint8
	Field TxnField
}

func (e *GitxnasExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

func (e *GitxnasExpr) String() string {
	return fmt.Sprintf("gitxnas %d %s", e.Index, e.Field)
}

type PushIntsExpr struct {
	Ints []uint64
}

func (e *PushIntsExpr) Execute(b *VmBranch) error {
	for range e.Ints {
		b.push(VmValue{T: VmTypeBytes})
	}
	b.Line++
	return nil
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

func (e *PushBytessExpr) Execute(b *VmBranch) error {
	for range e.Bytess {
		b.push(VmValue{T: VmTypeBytes})
	}
	b.Line++
	return nil
}

func (e *PushBytessExpr) String() string {
	var ss []string

	for _, bs := range e.Bytess {
		ss = append(ss, Bytes{Format: BytesBase64, Value: bs}.String())
	}

	return fmt.Sprintf("pushbytess %s", strings.Join(ss, " "))
}

type Bn256AddExpr struct{}

func (e *Bn256AddExpr) String() string {
	return "bn256_add"
}

func (e *Bn256AddExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Bn256Add = &Bn256AddExpr{}

type Bn256ScalarMulExpr struct{}

func (e *Bn256ScalarMulExpr) String() string {
	return "bn256_scalar_mul"
}

func (e *Bn256ScalarMulExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Bn256ScalarMul = &Bn256ScalarMulExpr{}

type Bn256PairingExpr struct{}

func (e *Bn256PairingExpr) String() string {
	return "bn256_pairing"
}

func (e *Bn256PairingExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Bn256Pairing = &Bn256PairingExpr{}

type FrameBuryExpr struct {
	Index int8
}

func (e *FrameBuryExpr) Execute(b *VmBranch) error {
	f := b.Frames[len(b.Frames)-1]

	v := b.pop(VmTypeAny)
	b.replace(uint8(int8(f.p)+e.Index), v)

	b.Line++
	return nil
}

func (e *FrameBuryExpr) String() string {
	return fmt.Sprintf("frame_bury %d", e.Index)
}

type FrameDigExpr struct {
	Index int8
}

func (e *FrameDigExpr) Execute(b *VmBranch) error {
	f := b.Frames[len(b.Frames)-1]

	v := b.Stack.Items[uint8(int8(f.p)+e.Index)]
	b.push(v)

	b.Line++
	return nil
}

func (e *FrameDigExpr) String() string {
	return fmt.Sprintf("frame_dig %d", e.Index)
}

type SetByteExpr struct{}

func (e *SetByteExpr) String() string {
	return "setbyte"
}

func (e *SetByteExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var SetByte = &SetByteExpr{}

type CoverExpr struct {
	Depth uint8
}

func (e *CoverExpr) String() string {
	return fmt.Sprintf("cover %d", e.Depth)
}

func (e *CoverExpr) Execute(b *VmBranch) error {
	is := []VmValue{}

	v := b.pop(VmTypeAny)

	for i := uint8(0); i < e.Depth; i++ {
		is = append(is, b.pop(VmTypeAny))
	}

	b.push(v)

	for i := len(is) - 1; i >= 0; i-- {
		b.push(is[i])
	}

	b.Line++
	return nil
}

type UncoverExpr struct {
	Depth uint8
}

func (e *UncoverExpr) Execute(b *VmBranch) error {
	is := []VmValue{}

	for i := uint8(0); i < e.Depth; i++ {
		is = append(is, b.pop(VmTypeAny))
	}

	v := b.pop(VmTypeAny)

	for i := len(is) - 1; i >= 0; i-- {
		b.push(is[i])
	}

	b.push(v)

	b.Line++
	return nil
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

func (e *AssetParamsGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type ConcatExpr struct{}

func (e *ConcatExpr) String() string {
	return "concat"
}

func (e *ConcatExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
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

func (e *ItxnExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
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

func (e *ItxnFieldExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.Line++
	return nil
}

type PopExpr struct{}

func (e *PopExpr) String() string {
	return "pop"
}

func (e *PopExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.Line++
	return nil
}

var Pop = &PopExpr{}

type SwapExpr struct{}

func (e *SwapExpr) String() string {
	return "swap"
}

func (e *SwapExpr) Execute(b *VmBranch) error {
	v1 := b.pop(VmTypeAny)
	v2 := b.pop(VmTypeAny)

	b.push(v1)
	b.push(v2)

	b.Line++
	return nil
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

func (e *GtxnaExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

type GtxnsaExpr struct {
	Field TxnField
	Index uint8
}

func (e *GtxnsaExpr) String() string {
	return fmt.Sprintf("gtxnsa %s %d", e.Field, e.Index)
}

func (e *GtxnsaExpr) Execute(b *VmBranch) error {
	spec, ok := txnFieldSpecByField(e.Field)
	if !ok {
		panic("unknown field")
	}

	b.pop(VmTypeUint64)
	b.push(VmValue{T: spec.Type().Vm()})

	b.Line++
	return nil
}

type GloadExpr struct {
	Group uint8
	Index uint8
}

func (e *GloadExpr) String() string {
	return fmt.Sprintf("gload %d %d", e.Group, e.Index)
}

func (e *GloadExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeAny})

	b.Line++
	return nil
}

type GloadsExpr struct {
	Index uint8
}

func (e *GloadsExpr) String() string {
	return fmt.Sprintf("gloads %d", e.Index)
}

func (e *GloadsExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})

	b.Line++
	return nil
}

type SqrtExpr struct{}

func (e *SqrtExpr) String() string {
	return "sqrt"
}

func (e *SqrtExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Sqrt = &SqrtExpr{}

type BalanceExpr struct{}

func (e *BalanceExpr) String() string {
	return "balance"
}

func (e *BalanceExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Balance = &BalanceExpr{}

type TxnasExpr struct {
	Field TxnField
}

func (e *TxnasExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
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

func (e *ExtractExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type ExpExpr struct{}

func (e *ExpExpr) String() string {
	return "exp"
}

func (e *ExpExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Exp = &ExpExpr{}

type AppParamsGetExpr struct {
	Field AppParamsField
}

func (e *AppParamsGetExpr) String() string {
	return fmt.Sprintf("app_params_get %s", e.Field)
}

func (e *AppParamsGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type AcctParamsGetExpr struct {
	Field AcctParamsField
}

func (e *AcctParamsGetExpr) String() string {
	return fmt.Sprintf("acct_params_get %s", e.Field)
}

func (e *AcctParamsGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type LogExpr struct{}

func (e *LogExpr) String() string {
	return "log"
}

func (e *LogExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.Line++
	return nil
}

var Log = &LogExpr{}

type BlockExpr struct {
	Field BlockField
}

func (e *BlockExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
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

func (e *VrfVerifyExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

type Extract3Expr struct{}

func (e *Extract3Expr) String() string {
	return "extract3"
}

func (e *Extract3Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
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

func (e *AssetHoldingGetExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeAny})
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
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

func (e *MinBalanceExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeAny)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var MinBalanceOp = &MinBalanceExpr{}

type GitxnExpr struct {
	Index uint8
	Field TxnField
}

func (e *GitxnExpr) String() string {
	return fmt.Sprintf("gitxn %d %s", e.Index, e.Field)
}

func (e *GitxnExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
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

func (e *GetByteExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var GetByte = &GetByteExpr{}

type Substring3Expr struct{}

func (e *Substring3Expr) String() string {
	return "substring3"
}

func (e *Substring3Expr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var Substring3 = &Substring3Expr{}

type ShlExpr struct{}

func (e *ShlExpr) String() string {
	return "shl"
}

func (e *ShlExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var ShiftLeft = &ShlExpr{}

type ShrExpr struct{}

func (e *ShrExpr) String() string {
	return "shr"
}

func (e *ShrExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var ShiftRight = &ShrExpr{}

type GetBitExpr struct{}

func (e *GetBitExpr) String() string {
	return "getbit"
}

func (e *GetBitExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var GetBit = &GetBitExpr{}

type SetBitExpr struct{}

func (e *SetBitExpr) String() string {
	return "setbit"
}

func (e *SetBitExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

var SetBit = &SetBitExpr{}

type GtxnsasExpr struct {
	Field TxnField
}

func (e *GtxnsasExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

func (e *GtxnsasExpr) String() string {
	return fmt.Sprintf("gtxnsas %s", e.Field)
}

type DigExpr struct {
	Index uint8
}

func (e *DigExpr) Execute(b *VmBranch) error {
	v := b.Stack.Items[uint8(len(b.Stack.Items))-1-e.Index]
	b.push(v)
	b.Line++
	return nil
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

func (e *BmulExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesMul = &BmulExpr{}

type BdivExpr struct{}

func (e *BdivExpr) String() string {
	return "b/"
}

func (e *BdivExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesDiv = &BdivExpr{}

type BplusExpr struct{}

func (e *BplusExpr) String() string {
	return "b+"
}

func (e *BplusExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesPlus = &BplusExpr{}

type GaidExpr struct {
	Group uint8
}

func (e *GaidExpr) String() string {
	return fmt.Sprintf("gaid %d", e.Group)
}

func (e *GaidExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

type GaidsExpr struct {
}

func (e *GaidsExpr) String() string {
	return "gaids"
}

func (e *GaidsExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})

	b.Line++
	return nil
}

var Gaids = &GaidsExpr{}

type LoadsExpr struct {
}

func (e *LoadsExpr) String() string {
	return "loads"
}

func (e *LoadsExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeAny})

	b.Line++
	return nil
}

var Loads = &LoadsExpr{}

type StoresExpr struct {
}

func (e *StoresExpr) Execute(b *VmBranch) error {
	v := b.pop(VmTypeAny)
	index := b.pop(VmTypeUint64)

	b.store(index, v)

	b.Line++
	return nil
}

func (e *StoresExpr) String() string {
	return "stores"
}

var Stores = &StoresExpr{}

type Dup2Expr struct{}

func (e *Dup2Expr) String() string {
	return "dup2"
}

func (e *Dup2Expr) Execute(b *VmBranch) error {
	v1 := b.pop(VmTypeAny)
	v2 := b.pop(VmTypeAny)
	b.push(v1)
	b.push(v2)
	b.push(v1)
	b.push(v2)
	b.Line++
	return nil
}

var Dup2 = &Dup2Expr{}

type SubstringExpr struct {
	Start uint8
	End   uint8
}

func (e *SubstringExpr) String() string {
	return fmt.Sprintf("substring %d %d", e.Start, e.End)
}

func (e *SubstringExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

type DivwExpr struct{}

func (e *DivwExpr) String() string {
	return "divw"
}

func (e *DivwExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var Divw = &DivwExpr{}

type SelectExpr struct{}

func (e *SelectExpr) String() string {
	return "select"
}

var Select = &SelectExpr{}

func (e *SelectExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeAny)
	b.pop(VmTypeAny)

	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

type BgteqExpr struct{}

func (e *BgteqExpr) String() string {
	return "b>="
}

var Bgteq = &BgteqExpr{}

type BGtExpr struct{}

func (e *BGtExpr) String() string {
	return "b>"
}

func (e *BGtExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BytesGt = &BGtExpr{}

type BytesLeExpr struct{}

func (e *BytesLeExpr) String() string {
	return "b<="
}

func (e *BytesLeExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BytesLe = &BytesLeExpr{}

type BytesGeExpr struct{}

func (e *BytesGeExpr) String() string {
	return "b>="
}

func (e *BytesGeExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BytesGe = &BytesGeExpr{}

type BytesEqExpr struct{}

func (e *BytesEqExpr) String() string {
	return "b=="
}

func (e *BytesEqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BytesEq = &BytesEqExpr{}

type BytesNeqExpr struct{}

func (e *BytesNeqExpr) String() string {
	return "b!="
}

func (e *BytesNeqExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
}

var BytesNeq = &BytesNeqExpr{}

type BytesModuloExpr struct{}

func (e *BytesModuloExpr) String() string {
	return "b%"
}

func (e *BytesModuloExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
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

func (e *BytesBitAndExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesBitAnd = &BytesBitAndExpr{}

type BsqrtExpr struct{}

func (e *BsqrtExpr) String() string {
	return "bsqrt"
}

func (e *BsqrtExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
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

func (e *GitxnaExpr) Execute(b *VmBranch) error {
	b.push(VmValue{T: VmTypeAny})
	b.Line++
	return nil
}

type AppGlobalDelExpr struct{}

func (e *AppGlobalDelExpr) String() string {
	return "app_global_del"
}

func (e *AppGlobalDelExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.Line++
	return nil
}

var AppGlobalDel = &AppGlobalDelExpr{}

type BltExpr struct{}

func (e *BltExpr) String() string {
	return "b<"
}

func (e *BltExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
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

func (e *BytesBitXorExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesBitXor = &BytesBitXorExpr{}

type BytesBitNotExpr struct{}

func (e *BytesBitNotExpr) String() string {
	return "b~"
}

func (e *BytesBitNotExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeBytes)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesBitNot = &BytesBitNotExpr{}

type BytesZeroExpr struct{}

func (e *BytesZeroExpr) String() string {
	return "bzero"
}

func (e *BytesZeroExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeBytes})
	b.Line++
	return nil
}

var BytesZero = &BytesZeroExpr{}

type AppOptedInExpr struct{}

func (e *AppOptedInExpr) String() string {
	return "app_opted_in"
}

func (e *AppOptedInExpr) Execute(b *VmBranch) error {
	b.pop(VmTypeUint64)
	b.pop(VmTypeUint64)
	b.push(VmValue{T: VmTypeUint64})
	b.Line++
	return nil
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
