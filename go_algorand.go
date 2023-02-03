package teal

import (
	"encoding/base32"
	"fmt"
	"strconv"

	"github.com/dragmz/teal/internal/protocol"
	"github.com/dragmz/teal/internal/transactions"
)

// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

const protoByte = 0x8a

// directRefEnabledVersion is the version of TEAL where opcodes
// that reference accounts, asas, and apps may do so directly, not requiring
// using an index into arrays.
const directRefEnabledVersion = 4

const fidoVersion = 7       // base64, json, secp256r1
const randomnessVersion = 7 // vrf_verify, block
const fpVersion = 8         // changes for frame pointers and simpler function discipline

// EXPERIMENTAL. These should be revisited whenever a new LogicSigVersion is
// moved from vFuture to a new consensus version. If they remain unready, bump
// their version, and fixup TestAssemble() in assembler_test.go.
const pairingVersion = 9 // bn256 opcodes. will add bls12-381, and unify the available opcodes.

// Unlimited Global Storage opcodes
const boxVersion = 8 // box_*

const anyImmediates = -1

func parseStringLiteral(input string) (result []byte, err error) {
	start := 0
	end := len(input) - 1
	if input[start] != '"' || input[end] != '"' {
		return nil, fmt.Errorf("no quotes")
	}
	start++

	escapeSeq := false
	hexSeq := false
	result = make([]byte, 0, end-start+1)

	// skip first and last quotes
	pos := start
	for pos < end {
		char := input[pos]
		if char == '\\' && !escapeSeq {
			if hexSeq {
				return nil, fmt.Errorf("escape seq inside hex number")
			}
			escapeSeq = true
			pos++
			continue
		}
		if escapeSeq {
			escapeSeq = false
			switch char {
			case 'n':
				char = '\n'
			case 'r':
				char = '\r'
			case 't':
				char = '\t'
			case '\\':
				char = '\\'
			case '"':
				char = '"'
			case 'x':
				hexSeq = true
				pos++
				continue
			default:
				return nil, fmt.Errorf("invalid escape seq \\%c", char)
			}
		}
		if hexSeq {
			hexSeq = false
			if pos >= len(input)-2 { // count a closing quote
				return nil, fmt.Errorf("non-terminated hex seq")
			}
			num, err := strconv.ParseUint(input[pos:pos+2], 16, 8)
			if err != nil {
				return nil, err
			}
			char = uint8(num)
			pos++
		}

		result = append(result, char)
		pos++
	}
	if escapeSeq || hexSeq {
		return nil, fmt.Errorf("non-terminated escape seq")
	}

	return
}

func base32DecodeAnyPadding(x string) (val []byte, err error) {
	val, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(x)
	if err != nil {
		// try again with standard padding
		var e2 error
		val, e2 = base32.StdEncoding.DecodeString(x)
		if e2 == nil {
			err = nil
		}
	}
	return
}

// fields

type StackType byte

const (
	// StackNone in an OpSpec shows that the op pops or yields nothing
	StackNone StackType = iota

	// StackAny in an OpSpec shows that the op pops or yield any type
	StackAny

	// StackUint64 in an OpSpec shows that the op pops or yields a uint64
	StackUint64

	// StackBytes in an OpSpec shows that the op pops or yields a []byte
	StackBytes
)

// StackTypes is an alias for a list of StackType with syntactic sugar
type StackTypes []StackType

type runMode uint64

const (
	// modeSig is LogicSig execution
	modeSig runMode = 1 << iota

	// modeApp is application/contract execution
	modeApp

	// local constant, run in any mode
	modeAny = modeSig | modeApp
)

// Copyright (C) 2019-2022 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

//go:generate stringer -type=TxnField,GlobalField,AssetParamsField,AppParamsField,AcctParamsField,AssetHoldingField,OnCompletionConstType,EcdsaCurve,Base64Encoding,JSONRefType,VrfStandard,BlockField -output=fields_string.go

// FieldSpec unifies the various specs for assembly, disassembly, and doc generation.
type FieldSpec interface {
	Field() byte
	Type() StackType
	OpVersion() uint64
	Note() string
	Version() uint64
}

// fieldSpecMap is something that yields a FieldSpec, given a name for the field
type fieldSpecMap interface {
	get(name string) (FieldSpec, bool)
}

// FieldGroup binds all the info for a field (names, int value, spec access) so
// they can be attached to opcodes and used by doc generation
type FieldGroup struct {
	Name  string
	Doc   string
	Names []string
	specs fieldSpecMap
}

// SpecByName returns a FieldsSpec for a name, respecting the "sparseness" of
// the Names array to hide some names
func (fg *FieldGroup) SpecByName(name string) (FieldSpec, bool) {
	if fs, ok := fg.specs.get(name); ok {
		if fg.Names[fs.Field()] != "" {
			return fs, true
		}
	}
	return nil, false
}

// TxnField is an enum type for `txn` and `gtxn`
type TxnField int

const (
	// Sender Transaction.Sender
	Sender TxnField = iota
	// Fee Transaction.Fee
	Fee
	// FirstValid Transaction.FirstValid
	FirstValid
	// FirstValidTime timestamp of block(FirstValid-1)
	FirstValidTime
	// LastValid Transaction.LastValid
	LastValid
	// Note Transaction.Note
	Note
	// Lease Transaction.Lease
	Lease
	// Receiver Transaction.Receiver
	Receiver
	// Amount Transaction.Amount
	Amount
	// CloseRemainderTo Transaction.CloseRemainderTo
	CloseRemainderTo
	// VotePK Transaction.VotePK
	VotePK
	// SelectionPK Transaction.SelectionPK
	SelectionPK
	// VoteFirst Transaction.VoteFirst
	VoteFirst
	// VoteLast Transaction.VoteLast
	VoteLast
	// VoteKeyDilution Transaction.VoteKeyDilution
	VoteKeyDilution
	// Type Transaction.Type
	Type
	// TypeEnum int(Transaction.Type)
	TypeEnum
	// XferAsset Transaction.XferAsset
	XferAsset
	// AssetAmount Transaction.AssetAmount
	AssetAmount
	// AssetSender Transaction.AssetSender
	AssetSender
	// AssetReceiver Transaction.AssetReceiver
	AssetReceiver
	// AssetCloseTo Transaction.AssetCloseTo
	AssetCloseTo
	// GroupIndex i for txngroup[i] == Txn
	GroupIndex
	// TxID Transaction.ID()
	TxID
	// ApplicationID basics.AppIndex
	ApplicationID
	// OnCompletion OnCompletion
	OnCompletion
	// ApplicationArgs  [][]byte
	ApplicationArgs
	// NumAppArgs len(ApplicationArgs)
	NumAppArgs
	// Accounts []basics.Address
	Accounts
	// NumAccounts len(Accounts)
	NumAccounts
	// ApprovalProgram []byte
	ApprovalProgram
	// ClearStateProgram []byte
	ClearStateProgram
	// RekeyTo basics.Address
	RekeyTo
	// ConfigAsset basics.AssetIndex
	ConfigAsset
	// ConfigAssetTotal AssetParams.Total
	ConfigAssetTotal
	// ConfigAssetDecimals AssetParams.Decimals
	ConfigAssetDecimals
	// ConfigAssetDefaultFrozen AssetParams.AssetDefaultFrozen
	ConfigAssetDefaultFrozen
	// ConfigAssetUnitName AssetParams.UnitName
	ConfigAssetUnitName
	// ConfigAssetName AssetParams.AssetName
	ConfigAssetName
	// ConfigAssetURL AssetParams.URL
	ConfigAssetURL
	// ConfigAssetMetadataHash AssetParams.MetadataHash
	ConfigAssetMetadataHash
	// ConfigAssetManager AssetParams.Manager
	ConfigAssetManager
	// ConfigAssetReserve AssetParams.Reserve
	ConfigAssetReserve
	// ConfigAssetFreeze AssetParams.Freeze
	ConfigAssetFreeze
	// ConfigAssetClawback AssetParams.Clawback
	ConfigAssetClawback
	//FreezeAsset  basics.AssetIndex
	FreezeAsset
	// FreezeAssetAccount basics.Address
	FreezeAssetAccount
	// FreezeAssetFrozen bool
	FreezeAssetFrozen
	// Assets []basics.AssetIndex
	Assets
	// NumAssets len(ForeignAssets)
	NumAssets
	// Applications []basics.AppIndex
	Applications
	// NumApplications len(ForeignApps)
	NumApplications

	// GlobalNumUint uint64
	GlobalNumUint
	// GlobalNumByteSlice uint64
	GlobalNumByteSlice
	// LocalNumUint uint64
	LocalNumUint
	// LocalNumByteSlice uint64
	LocalNumByteSlice

	// ExtraProgramPages AppParams.ExtraProgramPages
	ExtraProgramPages

	// Nonparticipation Transaction.Nonparticipation
	Nonparticipation

	// Logs Transaction.ApplyData.EvalDelta.Logs
	Logs

	// NumLogs len(Logs)
	NumLogs

	// CreatedAssetID Transaction.ApplyData.EvalDelta.ConfigAsset
	CreatedAssetID

	// CreatedApplicationID Transaction.ApplyData.EvalDelta.ApplicationID
	CreatedApplicationID

	// LastLog Logs[len(Logs)-1]
	LastLog

	// StateProofPK Transaction.StateProofPK
	StateProofPK

	// ApprovalProgramPages [][]byte
	ApprovalProgramPages

	// NumApprovalProgramPages = len(ApprovalProgramPages) // 4096
	NumApprovalProgramPages

	// ClearStateProgramPages [][]byte
	ClearStateProgramPages

	// NumClearStateProgramPages = len(ClearStateProgramPages) // 4096
	NumClearStateProgramPages

	invalidTxnField // compile-time constant for number of fields
)

func txnFieldSpecByField(f TxnField) (txnFieldSpec, bool) {
	if int(f) >= len(txnFieldSpecs) {
		return txnFieldSpec{}, false
	}
	return txnFieldSpecs[f], true
}

// TxnFieldNames are arguments to the 'txn' family of opcodes.
var TxnFieldNames [invalidTxnField]string

var txnFieldSpecByName = make(tfNameSpecMap, len(TxnFieldNames))

// simple interface used by doc generator for fields versioning
type tfNameSpecMap map[string]txnFieldSpec

func (s tfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

type txnFieldSpec struct {
	field      TxnField
	ftype      StackType
	array      bool   // Is this an array field?
	version    uint64 // When this field become available to txn/gtxn. 0=always
	itxVersion uint64 // When this field become available to itxn_field. 0=never
	effects    bool   // Is this a field on the "effects"? That is, something in ApplyData
	doc        string
}

func (fs txnFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs txnFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs txnFieldSpec) OpVersion() uint64 {
	return 0
}
func (fs txnFieldSpec) Version() uint64 {
	return fs.version
}
func (fs txnFieldSpec) Note() string {
	note := fs.doc
	if fs.effects {
		note = addExtra(note, "Application mode only")
	}
	return note
}

var txnFieldSpecs = [...]txnFieldSpec{
	{Sender, StackBytes, false, 0, 5, false, "32 byte address"},
	{Fee, StackUint64, false, 0, 5, false, "microalgos"},
	{FirstValid, StackUint64, false, 0, 0, false, "round number"},
	{FirstValidTime, StackUint64, false, randomnessVersion, 0, false, "UNIX timestamp of block before txn.FirstValid. Fails if negative"},
	{LastValid, StackUint64, false, 0, 0, false, "round number"},
	{Note, StackBytes, false, 0, 6, false, "Any data up to 1024 bytes"},
	{Lease, StackBytes, false, 0, 0, false, "32 byte lease value"},
	{Receiver, StackBytes, false, 0, 5, false, "32 byte address"},
	{Amount, StackUint64, false, 0, 5, false, "microalgos"},
	{CloseRemainderTo, StackBytes, false, 0, 5, false, "32 byte address"},
	{VotePK, StackBytes, false, 0, 6, false, "32 byte address"},
	{SelectionPK, StackBytes, false, 0, 6, false, "32 byte address"},
	{VoteFirst, StackUint64, false, 0, 6, false, "The first round that the participation key is valid."},
	{VoteLast, StackUint64, false, 0, 6, false, "The last round that the participation key is valid."},
	{VoteKeyDilution, StackUint64, false, 0, 6, false, "Dilution for the 2-level participation key"},
	{Type, StackBytes, false, 0, 5, false, "Transaction type as bytes"},
	{TypeEnum, StackUint64, false, 0, 5, false, "Transaction type as integer"},
	{XferAsset, StackUint64, false, 0, 5, false, "Asset ID"},
	{AssetAmount, StackUint64, false, 0, 5, false, "value in Asset's units"},
	{AssetSender, StackBytes, false, 0, 5, false,
		"32 byte address. Source of assets if Sender is the Asset's Clawback address."},
	{AssetReceiver, StackBytes, false, 0, 5, false, "32 byte address"},
	{AssetCloseTo, StackBytes, false, 0, 5, false, "32 byte address"},
	{GroupIndex, StackUint64, false, 0, 0, false,
		"Position of this transaction within an atomic transaction group. A stand-alone transaction is implicitly element 0 in a group of 1"},
	{TxID, StackBytes, false, 0, 0, false, "The computed ID for this transaction. 32 bytes."},
	{ApplicationID, StackUint64, false, 2, 6, false, "ApplicationID from ApplicationCall transaction"},
	{OnCompletion, StackUint64, false, 2, 6, false, "ApplicationCall transaction on completion action"},
	{ApplicationArgs, StackBytes, true, 2, 6, false,
		"Arguments passed to the application in the ApplicationCall transaction"},
	{NumAppArgs, StackUint64, false, 2, 0, false, "Number of ApplicationArgs"},
	{Accounts, StackBytes, true, 2, 6, false, "Accounts listed in the ApplicationCall transaction"},
	{NumAccounts, StackUint64, false, 2, 0, false, "Number of Accounts"},
	{ApprovalProgram, StackBytes, false, 2, 6, false, "Approval program"},
	{ClearStateProgram, StackBytes, false, 2, 6, false, "Clear state program"},
	{RekeyTo, StackBytes, false, 2, 6, false, "32 byte Sender's new AuthAddr"},
	{ConfigAsset, StackUint64, false, 2, 5, false, "Asset ID in asset config transaction"},
	{ConfigAssetTotal, StackUint64, false, 2, 5, false, "Total number of units of this asset created"},
	{ConfigAssetDecimals, StackUint64, false, 2, 5, false,
		"Number of digits to display after the decimal place when displaying the asset"},
	{ConfigAssetDefaultFrozen, StackUint64, false, 2, 5, false,
		"Whether the asset's slots are frozen by default or not, 0 or 1"},
	{ConfigAssetUnitName, StackBytes, false, 2, 5, false, "Unit name of the asset"},
	{ConfigAssetName, StackBytes, false, 2, 5, false, "The asset name"},
	{ConfigAssetURL, StackBytes, false, 2, 5, false, "URL"},
	{ConfigAssetMetadataHash, StackBytes, false, 2, 5, false,
		"32 byte commitment to unspecified asset metadata"},
	{ConfigAssetManager, StackBytes, false, 2, 5, false, "32 byte address"},
	{ConfigAssetReserve, StackBytes, false, 2, 5, false, "32 byte address"},
	{ConfigAssetFreeze, StackBytes, false, 2, 5, false, "32 byte address"},
	{ConfigAssetClawback, StackBytes, false, 2, 5, false, "32 byte address"},
	{FreezeAsset, StackUint64, false, 2, 5, false, "Asset ID being frozen or un-frozen"},
	{FreezeAssetAccount, StackBytes, false, 2, 5, false,
		"32 byte address of the account whose asset slot is being frozen or un-frozen"},
	{FreezeAssetFrozen, StackUint64, false, 2, 5, false, "The new frozen value, 0 or 1"},
	{Assets, StackUint64, true, 3, 6, false, "Foreign Assets listed in the ApplicationCall transaction"},
	{NumAssets, StackUint64, false, 3, 0, false, "Number of Assets"},
	{Applications, StackUint64, true, 3, 6, false, "Foreign Apps listed in the ApplicationCall transaction"},
	{NumApplications, StackUint64, false, 3, 0, false, "Number of Applications"},
	{GlobalNumUint, StackUint64, false, 3, 6, false, "Number of global state integers in ApplicationCall"},
	{GlobalNumByteSlice, StackUint64, false, 3, 6, false, "Number of global state byteslices in ApplicationCall"},
	{LocalNumUint, StackUint64, false, 3, 6, false, "Number of local state integers in ApplicationCall"},
	{LocalNumByteSlice, StackUint64, false, 3, 6, false, "Number of local state byteslices in ApplicationCall"},
	{ExtraProgramPages, StackUint64, false, 4, 6, false,
		"Number of additional pages for each of the application's approval and clear state programs. An ExtraProgramPages of 1 means 2048 more total bytes, or 1024 for each program."},
	{Nonparticipation, StackUint64, false, 5, 6, false, "Marks an account nonparticipating for rewards"},

	// "Effects" Last two things are always going to: 0, true
	{Logs, StackBytes, true, 5, 0, true, "Log messages emitted by an application call (only with `itxn` in v5)"},
	{NumLogs, StackUint64, false, 5, 0, true, "Number of Logs (only with `itxn` in v5)"},
	{CreatedAssetID, StackUint64, false, 5, 0, true,
		"Asset ID allocated by the creation of an ASA (only with `itxn` in v5)"},
	{CreatedApplicationID, StackUint64, false, 5, 0, true,
		"ApplicationID allocated by the creation of an application (only with `itxn` in v5)"},
	{LastLog, StackBytes, false, 6, 0, true, "The last message emitted. Empty bytes if none were emitted"},

	// Not an effect. Just added after the effects fields.
	{StateProofPK, StackBytes, false, 6, 6, false, "64 byte state proof public key"},

	// Pseudo-fields to aid access to large programs (bigger than TEAL values)
	// reading in a txn seems not *super* useful, but setting in `itxn` is critical to inner app factories
	{ApprovalProgramPages, StackBytes, true, 7, 7, false, "Approval Program as an array of pages"},
	{NumApprovalProgramPages, StackUint64, false, 7, 0, false, "Number of Approval Program pages"},
	{ClearStateProgramPages, StackBytes, true, 7, 7, false, "ClearState Program as an array of pages"},
	{NumClearStateProgramPages, StackUint64, false, 7, 0, false, "Number of ClearState Program pages"},
}

// TxnFields contains info on the arguments to the txn* family of opcodes
var TxnFields = FieldGroup{
	"txn", "",
	TxnFieldNames[:],
	txnFieldSpecByName,
}

// TxnScalarFields narrows TxnFields to only have the names of scalar fetching opcodes
var TxnScalarFields = FieldGroup{
	"txn", "Fields (see [transaction reference](https://developer.algorand.org/docs/reference/transactions/))",
	txnScalarFieldNames(),
	txnFieldSpecByName,
}

// txnScalarFieldNames are txn field names that return scalars. Return value is
// a "sparse" slice, the names appear at their usual index, array slots are set
// to "".  They are laid out this way so that it is possible to get the name
// from the index value.
func txnScalarFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.array {
			names[i] = ""
		} else {
			names[i] = fs.field.String()
		}
	}
	return names
}

// TxnArrayFields narows TxnFields to only have the names of array fetching opcodes
var TxnArrayFields = FieldGroup{
	"txna", "Fields (see [transaction reference](https://developer.algorand.org/docs/reference/transactions/))",
	txnaFieldNames(),
	txnFieldSpecByName,
}

// txnaFieldNames are txn field names that return arrays. Return value is a
// "sparse" slice, the names appear at their usual index, non-array slots are
// set to "".  They are laid out this way so that it is possible to get the name
// from the index value.
func txnaFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.array {
			names[i] = fs.field.String()
		} else {
			names[i] = ""
		}
	}
	return names
}

// ItxnSettableFields collects info for itxn_field opcode
var ItxnSettableFields = FieldGroup{
	"itxn_field", "",
	itxnSettableFieldNames(),
	txnFieldSpecByName,
}

// itxnSettableFieldNames are txn field names that can be set by
// itxn_field. Return value is a "sparse" slice, the names appear at their usual
// index, unsettable slots are set to "".  They are laid out this way so that it is
// possible to get the name from the index value.
func itxnSettableFieldNames() []string {
	names := make([]string, len(txnFieldSpecs))
	for i, fs := range txnFieldSpecs {
		if fs.itxVersion == 0 {
			names[i] = ""
		} else {
			names[i] = fs.field.String()
		}
	}
	return names
}

var innerTxnTypes = map[string]uint64{
	string(protocol.PaymentTx):         5,
	string(protocol.KeyRegistrationTx): 6,
	string(protocol.AssetTransferTx):   5,
	string(protocol.AssetConfigTx):     5,
	string(protocol.AssetFreezeTx):     5,
	string(protocol.ApplicationCallTx): 6,
}

// TxnTypeNames is the values of Txn.Type in enum order
var TxnTypeNames = [...]string{
	string(protocol.UnknownTx),
	string(protocol.PaymentTx),
	string(protocol.KeyRegistrationTx),
	string(protocol.AssetConfigTx),
	string(protocol.AssetTransferTx),
	string(protocol.AssetFreezeTx),
	string(protocol.ApplicationCallTx),
}

// map txn type names (long and short) to index/enum value
var txnTypeMap = make(map[string]uint64)

// OnCompletionConstType is the same as transactions.OnCompletion
type OnCompletionConstType transactions.OnCompletion

const (
	// NoOp = transactions.NoOpOC
	NoOp = OnCompletionConstType(transactions.NoOpOC)
	// OptIn = transactions.OptInOC
	OptIn = OnCompletionConstType(transactions.OptInOC)
	// CloseOut = transactions.CloseOutOC
	CloseOut = OnCompletionConstType(transactions.CloseOutOC)
	// ClearState = transactions.ClearStateOC
	ClearState = OnCompletionConstType(transactions.ClearStateOC)
	// UpdateApplication = transactions.UpdateApplicationOC
	UpdateApplication = OnCompletionConstType(transactions.UpdateApplicationOC)
	// DeleteApplication = transactions.DeleteApplicationOC
	DeleteApplication = OnCompletionConstType(transactions.DeleteApplicationOC)
	// end of constants
	invalidOnCompletionConst = DeleteApplication + 1
)

// OnCompletionNames is the string names of Txn.OnCompletion, array index is the const value
var OnCompletionNames [invalidOnCompletionConst]string

// onCompletionMap maps symbolic name to uint64 for assembleInt
var onCompletionMap map[string]uint64

// GlobalField is an enum for `global` opcode
type GlobalField uint64

const (
	// MinTxnFee ConsensusParams.MinTxnFee
	MinTxnFee GlobalField = iota
	// MinBalance ConsensusParams.MinBalance
	MinBalance
	// MaxTxnLife ConsensusParams.MaxTxnLife
	MaxTxnLife
	// ZeroAddress [32]byte{0...}
	ZeroAddress
	// GroupSize len(txn group)
	GroupSize

	// v2

	// LogicSigVersion ConsensusParams.LogicSigVersion
	LogicSigVersion
	// Round basics.Round
	Round
	// LatestTimestamp uint64
	LatestTimestamp
	// CurrentApplicationID uint64
	CurrentApplicationID

	// v3

	// CreatorAddress [32]byte
	CreatorAddress

	// v5

	// CurrentApplicationAddress [32]byte
	CurrentApplicationAddress
	// GroupID [32]byte
	GroupID

	// v6

	// OpcodeBudget The remaining budget available for execution
	OpcodeBudget

	// CallerApplicationID The ID of the caller app, else 0
	CallerApplicationID

	// CallerApplicationAddress The Address of the caller app, else ZeroAddress
	CallerApplicationAddress

	invalidGlobalField // compile-time constant for number of fields
)

// GlobalFieldNames are arguments to the 'global' opcode
var GlobalFieldNames [invalidGlobalField]string

type globalFieldSpec struct {
	field   GlobalField
	ftype   StackType
	mode    runMode
	version uint64
	doc     string
}

func (fs globalFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs globalFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs globalFieldSpec) OpVersion() uint64 {
	return 0
}
func (fs globalFieldSpec) Version() uint64 {
	return fs.version
}
func (fs globalFieldSpec) Note() string {
	note := fs.doc
	if fs.mode == modeApp {
		note = addExtra(note, "Application mode only.")
	}
	// There are no Signature mode only globals
	return note
}

var globalFieldSpecs = [...]globalFieldSpec{
	// version 0 is the same as v1 (initial release)
	{MinTxnFee, StackUint64, modeAny, 0, "microalgos"},
	{MinBalance, StackUint64, modeAny, 0, "microalgos"},
	{MaxTxnLife, StackUint64, modeAny, 0, "rounds"},
	{ZeroAddress, StackBytes, modeAny, 0, "32 byte address of all zero bytes"},
	{GroupSize, StackUint64, modeAny, 0,
		"Number of transactions in this atomic transaction group. At least 1"},
	{LogicSigVersion, StackUint64, modeAny, 2, "Maximum supported version"},
	{Round, StackUint64, modeApp, 2, "Current round number"},
	{LatestTimestamp, StackUint64, modeApp, 2,
		"Last confirmed block UNIX timestamp. Fails if negative"},
	{CurrentApplicationID, StackUint64, modeApp, 2, "ID of current application executing"},
	{CreatorAddress, StackBytes, modeApp, 3,
		"Address of the creator of the current application"},
	{CurrentApplicationAddress, StackBytes, modeApp, 5,
		"Address that the current application controls"},
	{GroupID, StackBytes, modeAny, 5,
		"ID of the transaction group. 32 zero bytes if the transaction is not part of a group."},
	{OpcodeBudget, StackUint64, modeAny, 6,
		"The remaining cost that can be spent by opcodes in this program."},
	{CallerApplicationID, StackUint64, modeApp, 6,
		"The application ID of the application that called this application. 0 if this application is at the top-level."},
	{CallerApplicationAddress, StackBytes, modeApp, 6,
		"The application address of the application that called this application. ZeroAddress if this application is at the top-level."},
}

func globalFieldSpecByField(f GlobalField) (globalFieldSpec, bool) {
	if int(f) >= len(globalFieldSpecs) {
		return globalFieldSpec{}, false
	}
	return globalFieldSpecs[f], true
}

var globalFieldSpecByName = make(gfNameSpecMap, len(GlobalFieldNames))

type gfNameSpecMap map[string]globalFieldSpec

func (s gfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// GlobalFields has info on the global opcode's immediate
var GlobalFields = FieldGroup{
	"global", "Fields",
	GlobalFieldNames[:],
	globalFieldSpecByName,
}

// EcdsaCurve is an enum for `ecdsa_` opcodes
type EcdsaCurve int

const (
	// Secp256k1 curve for bitcoin/ethereum
	Secp256k1 EcdsaCurve = iota
	// Secp256r1 curve
	Secp256r1
	invalidEcdsaCurve // compile-time constant for number of fields
)

var ecdsaCurveNames [invalidEcdsaCurve]string

type ecdsaCurveSpec struct {
	field   EcdsaCurve
	version uint64
	doc     string
}

func (fs ecdsaCurveSpec) Field() byte {
	return byte(fs.field)
}
func (fs ecdsaCurveSpec) Type() StackType {
	return StackNone // Will not show, since all are untyped
}
func (fs ecdsaCurveSpec) OpVersion() uint64 {
	return 5
}
func (fs ecdsaCurveSpec) Version() uint64 {
	return fs.version
}
func (fs ecdsaCurveSpec) Note() string {
	return fs.doc
}

var ecdsaCurveSpecs = [...]ecdsaCurveSpec{
	{Secp256k1, 5, "secp256k1 curve, used in Bitcoin"},
	{Secp256r1, fidoVersion, "secp256r1 curve, NIST standard"},
}

func ecdsaCurveSpecByField(c EcdsaCurve) (ecdsaCurveSpec, bool) {
	if int(c) >= len(ecdsaCurveSpecs) {
		return ecdsaCurveSpec{}, false
	}
	return ecdsaCurveSpecs[c], true
}

var ecdsaCurveSpecByName = make(ecdsaCurveNameSpecMap, len(ecdsaCurveNames))

type ecdsaCurveNameSpecMap map[string]ecdsaCurveSpec

func (s ecdsaCurveNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// EcdsaCurves collects details about the constants used to describe EcdsaCurves
var EcdsaCurves = FieldGroup{
	"ECDSA", "Curves",
	ecdsaCurveNames[:],
	ecdsaCurveSpecByName,
}

// Base64Encoding is an enum for the `base64decode` opcode
type Base64Encoding int

const (
	// URLEncoding represents the base64url encoding defined in https://www.rfc-editor.org/rfc/rfc4648.html
	URLEncoding Base64Encoding = iota
	// StdEncoding represents the standard encoding of the RFC
	StdEncoding
	invalidBase64Encoding // compile-time constant for number of fields
)

var base64EncodingNames [invalidBase64Encoding]string

type base64EncodingSpec struct {
	field   Base64Encoding
	version uint64
}

var base64EncodingSpecs = [...]base64EncodingSpec{
	{URLEncoding, 6},
	{StdEncoding, 6},
}

func base64EncodingSpecByField(e Base64Encoding) (base64EncodingSpec, bool) {
	if int(e) >= len(base64EncodingSpecs) {
		return base64EncodingSpec{}, false
	}
	return base64EncodingSpecs[e], true
}

var base64EncodingSpecByName = make(base64EncodingSpecMap, len(base64EncodingNames))

type base64EncodingSpecMap map[string]base64EncodingSpec

func (fs base64EncodingSpec) Field() byte {
	return byte(fs.field)
}
func (fs base64EncodingSpec) Type() StackType {
	return StackAny // Will not show in docs, since all are untyped
}
func (fs base64EncodingSpec) OpVersion() uint64 {
	return 6
}
func (fs base64EncodingSpec) Version() uint64 {
	return fs.version
}
func (fs base64EncodingSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s base64EncodingSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// Base64Encodings describes the base64_encode immediate
var Base64Encodings = FieldGroup{
	"base64", "Encodings",
	base64EncodingNames[:],
	base64EncodingSpecByName,
}

// JSONRefType is an enum for the `json_ref` opcode
type JSONRefType int

const (
	// JSONString represents string json value
	JSONString JSONRefType = iota
	// JSONUint64 represents uint64 json value
	JSONUint64
	// JSONObject represents json object
	JSONObject
	invalidJSONRefType // compile-time constant for number of fields
)

var jsonRefTypeNames [invalidJSONRefType]string

type jsonRefSpec struct {
	field   JSONRefType
	ftype   StackType
	version uint64
}

var jsonRefSpecs = [...]jsonRefSpec{
	{JSONString, StackBytes, fidoVersion},
	{JSONUint64, StackUint64, fidoVersion},
	{JSONObject, StackBytes, fidoVersion},
}

func jsonRefSpecByField(r JSONRefType) (jsonRefSpec, bool) {
	if int(r) >= len(jsonRefSpecs) {
		return jsonRefSpec{}, false
	}
	return jsonRefSpecs[r], true
}

var jsonRefSpecByName = make(jsonRefSpecMap, len(jsonRefTypeNames))

type jsonRefSpecMap map[string]jsonRefSpec

func (fs jsonRefSpec) Field() byte {
	return byte(fs.field)
}
func (fs jsonRefSpec) Type() StackType {
	return fs.ftype
}
func (fs jsonRefSpec) OpVersion() uint64 {
	return fidoVersion
}
func (fs jsonRefSpec) Version() uint64 {
	return fs.version
}
func (fs jsonRefSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s jsonRefSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// JSONRefTypes describes the json_ref immediate
var JSONRefTypes = FieldGroup{
	"json_ref", "Types",
	jsonRefTypeNames[:],
	jsonRefSpecByName,
}

// VrfStandard is an enum for the `vrf_verify` opcode
type VrfStandard int

const (
	// VrfAlgorand is the built-in VRF of the Algorand chain
	VrfAlgorand        VrfStandard = iota
	invalidVrfStandard             // compile-time constant for number of fields
)

var vrfStandardNames [invalidVrfStandard]string

type vrfStandardSpec struct {
	field   VrfStandard
	version uint64
}

var vrfStandardSpecs = [...]vrfStandardSpec{
	{VrfAlgorand, randomnessVersion},
}

func vrfStandardSpecByField(r VrfStandard) (vrfStandardSpec, bool) {
	if int(r) >= len(vrfStandardSpecs) {
		return vrfStandardSpec{}, false
	}
	return vrfStandardSpecs[r], true
}

var vrfStandardSpecByName = make(vrfStandardSpecMap, len(vrfStandardNames))

type vrfStandardSpecMap map[string]vrfStandardSpec

func (s vrfStandardSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

func (fs vrfStandardSpec) Field() byte {
	return byte(fs.field)
}

func (fs vrfStandardSpec) Type() StackType {
	return StackNone // Will not show, since all are the same
}

func (fs vrfStandardSpec) OpVersion() uint64 {
	return randomnessVersion
}

func (fs vrfStandardSpec) Version() uint64 {
	return fs.version
}

func (fs vrfStandardSpec) Note() string {
	note := "" // no doc list?
	return note
}

func (s vrfStandardSpecMap) SpecByName(name string) FieldSpec {
	return s[name]
}

// VrfStandards describes the json_ref immediate
var VrfStandards = FieldGroup{
	"vrf_verify", "Standards",
	vrfStandardNames[:],
	vrfStandardSpecByName,
}

// BlockField is an enum for the `block` opcode
type BlockField int

const (
	// BlkSeed is the Block's vrf seed
	BlkSeed BlockField = iota
	// BlkTimestamp is the Block's timestamp, seconds from epoch
	BlkTimestamp
	invalidBlockField // compile-time constant for number of fields
)

var blockFieldNames [invalidBlockField]string

type blockFieldSpec struct {
	field   BlockField
	ftype   StackType
	version uint64
}

var blockFieldSpecs = [...]blockFieldSpec{
	{BlkSeed, StackBytes, randomnessVersion},
	{BlkTimestamp, StackUint64, randomnessVersion},
}

func blockFieldSpecByField(r BlockField) (blockFieldSpec, bool) {
	if int(r) >= len(blockFieldSpecs) {
		return blockFieldSpec{}, false
	}
	return blockFieldSpecs[r], true
}

var blockFieldSpecByName = make(blockFieldSpecMap, len(blockFieldNames))

type blockFieldSpecMap map[string]blockFieldSpec

func (s blockFieldSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

func (fs blockFieldSpec) Field() byte {
	return byte(fs.field)
}

func (fs blockFieldSpec) Type() StackType {
	return fs.ftype
}

func (fs blockFieldSpec) OpVersion() uint64 {
	return randomnessVersion
}

func (fs blockFieldSpec) Version() uint64 {
	return fs.version
}

func (fs blockFieldSpec) Note() string {
	return ""
}

func (s blockFieldSpecMap) SpecByName(name string) FieldSpec {
	return s[name]
}

// BlockFields describes the json_ref immediate
var BlockFields = FieldGroup{
	"block", "Fields",
	blockFieldNames[:],
	blockFieldSpecByName,
}

// AssetHoldingField is an enum for `asset_holding_get` opcode
type AssetHoldingField int

const (
	// AssetBalance AssetHolding.Amount
	AssetBalance AssetHoldingField = iota
	// AssetFrozen AssetHolding.Frozen
	AssetFrozen
	invalidAssetHoldingField // compile-time constant for number of fields
)

var assetHoldingFieldNames [invalidAssetHoldingField]string

type assetHoldingFieldSpec struct {
	field   AssetHoldingField
	ftype   StackType
	version uint64
	doc     string
}

func (fs assetHoldingFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs assetHoldingFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs assetHoldingFieldSpec) OpVersion() uint64 {
	return 2
}
func (fs assetHoldingFieldSpec) Version() uint64 {
	return fs.version
}
func (fs assetHoldingFieldSpec) Note() string {
	return fs.doc
}

var assetHoldingFieldSpecs = [...]assetHoldingFieldSpec{
	{AssetBalance, StackUint64, 2, "Amount of the asset unit held by this account"},
	{AssetFrozen, StackUint64, 2, "Is the asset frozen or not"},
}

func assetHoldingFieldSpecByField(f AssetHoldingField) (assetHoldingFieldSpec, bool) {
	if int(f) >= len(assetHoldingFieldSpecs) {
		return assetHoldingFieldSpec{}, false
	}
	return assetHoldingFieldSpecs[f], true
}

var assetHoldingFieldSpecByName = make(ahfNameSpecMap, len(assetHoldingFieldNames))

type ahfNameSpecMap map[string]assetHoldingFieldSpec

func (s ahfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AssetHoldingFields describes asset_holding_get's immediates
var AssetHoldingFields = FieldGroup{
	"asset_holding", "Fields",
	assetHoldingFieldNames[:],
	assetHoldingFieldSpecByName,
}

// AssetParamsField is an enum for `asset_params_get` opcode
type AssetParamsField int

const (
	// AssetTotal AssetParams.Total
	AssetTotal AssetParamsField = iota
	// AssetDecimals AssetParams.Decimals
	AssetDecimals
	// AssetDefaultFrozen AssetParams.AssetDefaultFrozen
	AssetDefaultFrozen
	// AssetUnitName AssetParams.UnitName
	AssetUnitName
	// AssetName AssetParams.AssetName
	AssetName
	// AssetURL AssetParams.URL
	AssetURL
	// AssetMetadataHash AssetParams.MetadataHash
	AssetMetadataHash
	// AssetManager AssetParams.Manager
	AssetManager
	// AssetReserve AssetParams.Reserve
	AssetReserve
	// AssetFreeze AssetParams.Freeze
	AssetFreeze
	// AssetClawback AssetParams.Clawback
	AssetClawback

	// AssetCreator is not *in* the Params, but it is uniquely determined.
	AssetCreator

	invalidAssetParamsField // compile-time constant for number of fields
)

var assetParamsFieldNames [invalidAssetParamsField]string

type assetParamsFieldSpec struct {
	field   AssetParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs assetParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs assetParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs assetParamsFieldSpec) OpVersion() uint64 {
	return 2
}
func (fs assetParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs assetParamsFieldSpec) Note() string {
	return fs.doc
}

var assetParamsFieldSpecs = [...]assetParamsFieldSpec{
	{AssetTotal, StackUint64, 2, "Total number of units of this asset"},
	{AssetDecimals, StackUint64, 2, "See AssetParams.Decimals"},
	{AssetDefaultFrozen, StackUint64, 2, "Frozen by default or not"},
	{AssetUnitName, StackBytes, 2, "Asset unit name"},
	{AssetName, StackBytes, 2, "Asset name"},
	{AssetURL, StackBytes, 2, "URL with additional info about the asset"},
	{AssetMetadataHash, StackBytes, 2, "Arbitrary commitment"},
	{AssetManager, StackBytes, 2, "Manager address"},
	{AssetReserve, StackBytes, 2, "Reserve address"},
	{AssetFreeze, StackBytes, 2, "Freeze address"},
	{AssetClawback, StackBytes, 2, "Clawback address"},
	{AssetCreator, StackBytes, 5, "Creator address"},
}

func assetParamsFieldSpecByField(f AssetParamsField) (assetParamsFieldSpec, bool) {
	if int(f) >= len(assetParamsFieldSpecs) {
		return assetParamsFieldSpec{}, false
	}
	return assetParamsFieldSpecs[f], true
}

var assetParamsFieldSpecByName = make(apfNameSpecMap, len(assetParamsFieldNames))

type apfNameSpecMap map[string]assetParamsFieldSpec

func (s apfNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AssetParamsFields describes asset_params_get's immediates
var AssetParamsFields = FieldGroup{
	"asset_params", "Fields",
	assetParamsFieldNames[:],
	assetParamsFieldSpecByName,
}

// AppParamsField is an enum for `app_params_get` opcode
type AppParamsField int

const (
	// AppApprovalProgram AppParams.ApprovalProgram
	AppApprovalProgram AppParamsField = iota
	// AppClearStateProgram AppParams.ClearStateProgram
	AppClearStateProgram
	// AppGlobalNumUint AppParams.StateSchemas.GlobalStateSchema.NumUint
	AppGlobalNumUint
	// AppGlobalNumByteSlice AppParams.StateSchemas.GlobalStateSchema.NumByteSlice
	AppGlobalNumByteSlice
	// AppLocalNumUint AppParams.StateSchemas.LocalStateSchema.NumUint
	AppLocalNumUint
	// AppLocalNumByteSlice AppParams.StateSchemas.LocalStateSchema.NumByteSlice
	AppLocalNumByteSlice
	// AppExtraProgramPages AppParams.ExtraProgramPages
	AppExtraProgramPages

	// AppCreator is not *in* the Params, but it is uniquely determined.
	AppCreator

	// AppAddress is also not *in* the Params, but can be derived
	AppAddress

	invalidAppParamsField // compile-time constant for number of fields
)

var appParamsFieldNames [invalidAppParamsField]string

type appParamsFieldSpec struct {
	field   AppParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs appParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs appParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs appParamsFieldSpec) OpVersion() uint64 {
	return 5
}
func (fs appParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs appParamsFieldSpec) Note() string {
	return fs.doc
}

var appParamsFieldSpecs = [...]appParamsFieldSpec{
	{AppApprovalProgram, StackBytes, 5, "Bytecode of Approval Program"},
	{AppClearStateProgram, StackBytes, 5, "Bytecode of Clear State Program"},
	{AppGlobalNumUint, StackUint64, 5, "Number of uint64 values allowed in Global State"},
	{AppGlobalNumByteSlice, StackUint64, 5, "Number of byte array values allowed in Global State"},
	{AppLocalNumUint, StackUint64, 5, "Number of uint64 values allowed in Local State"},
	{AppLocalNumByteSlice, StackUint64, 5, "Number of byte array values allowed in Local State"},
	{AppExtraProgramPages, StackUint64, 5, "Number of Extra Program Pages of code space"},
	{AppCreator, StackBytes, 5, "Creator address"},
	{AppAddress, StackBytes, 5, "Address for which this application has authority"},
}

func appParamsFieldSpecByField(f AppParamsField) (appParamsFieldSpec, bool) {
	if int(f) >= len(appParamsFieldSpecs) {
		return appParamsFieldSpec{}, false
	}
	return appParamsFieldSpecs[f], true
}

var appParamsFieldSpecByName = make(appNameSpecMap, len(appParamsFieldNames))

// simple interface used by doc generator for fields versioning
type appNameSpecMap map[string]appParamsFieldSpec

func (s appNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AppParamsFields describes app_params_get's immediates
var AppParamsFields = FieldGroup{
	"app_params", "Fields",
	appParamsFieldNames[:],
	appParamsFieldSpecByName,
}

// AcctParamsField is an enum for `acct_params_get` opcode
type AcctParamsField int

const (
	// AcctBalance is the balance, with pending rewards
	AcctBalance AcctParamsField = iota
	// AcctMinBalance is algos needed for this accounts apps and assets
	AcctMinBalance
	// AcctAuthAddr is the rekeyed address if any, else ZeroAddress
	AcctAuthAddr

	// AcctTotalNumUint is the count of all uints from created global apps or opted in locals
	AcctTotalNumUint
	// AcctTotalNumByteSlice is the count of all byte slices from created global apps or opted in locals
	AcctTotalNumByteSlice

	// AcctTotalExtraAppPages is the extra code pages across all apps
	AcctTotalExtraAppPages

	// AcctTotalAppsCreated is the number of apps created by this account
	AcctTotalAppsCreated
	// AcctTotalAppsOptedIn is the number of apps opted in by this account
	AcctTotalAppsOptedIn
	// AcctTotalAssetsCreated is the number of ASAs created by this account
	AcctTotalAssetsCreated
	// AcctTotalAssets is the number of ASAs opted in by this account (always includes AcctTotalAssetsCreated)
	AcctTotalAssets
	// AcctTotalBoxes is the number of boxes created by the app this account is associated with
	AcctTotalBoxes
	// AcctTotalBoxBytes is the number of bytes in all boxes of this app account
	AcctTotalBoxBytes

	// AcctTotalAppSchema - consider how to expose

	invalidAcctParamsField // compile-time constant for number of fields
)

var acctParamsFieldNames [invalidAcctParamsField]string

type acctParamsFieldSpec struct {
	field   AcctParamsField
	ftype   StackType
	version uint64
	doc     string
}

func (fs acctParamsFieldSpec) Field() byte {
	return byte(fs.field)
}
func (fs acctParamsFieldSpec) Type() StackType {
	return fs.ftype
}
func (fs acctParamsFieldSpec) OpVersion() uint64 {
	return 6
}
func (fs acctParamsFieldSpec) Version() uint64 {
	return fs.version
}
func (fs acctParamsFieldSpec) Note() string {
	return fs.doc
}

var acctParamsFieldSpecs = [...]acctParamsFieldSpec{
	{AcctBalance, StackUint64, 6, "Account balance in microalgos"},
	{AcctMinBalance, StackUint64, 6, "Minimum required balance for account, in microalgos"},
	{AcctAuthAddr, StackBytes, 6, "Address the account is rekeyed to."},

	{AcctTotalNumUint, StackUint64, 8, "The total number of uint64 values allocated by this account in Global and Local States."},
	{AcctTotalNumByteSlice, StackUint64, 8, "The total number of byte array values allocated by this account in Global and Local States."},
	{AcctTotalExtraAppPages, StackUint64, 8, "The number of extra app code pages used by this account."},
	{AcctTotalAppsCreated, StackUint64, 8, "The number of existing apps created by this account."},
	{AcctTotalAppsOptedIn, StackUint64, 8, "The number of apps this account is opted into."},
	{AcctTotalAssetsCreated, StackUint64, 8, "The number of existing ASAs created by this account."},
	{AcctTotalAssets, StackUint64, 8, "The numbers of ASAs held by this account (including ASAs this account created)."},
	{AcctTotalBoxes, StackUint64, boxVersion, "The number of existing boxes created by this account's app."},
	{AcctTotalBoxBytes, StackUint64, boxVersion, "The total number of bytes used by this account's app's box keys and values."},
}

func acctParamsFieldSpecByField(f AcctParamsField) (acctParamsFieldSpec, bool) {
	if int(f) >= len(acctParamsFieldSpecs) {
		return acctParamsFieldSpec{}, false
	}
	return acctParamsFieldSpecs[f], true
}

var acctParamsFieldSpecByName = make(acctNameSpecMap, len(acctParamsFieldNames))

type acctNameSpecMap map[string]acctParamsFieldSpec

func (s acctNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// AcctParamsFields describes acct_params_get's immediates
var AcctParamsFields = FieldGroup{
	"acct_params", "Fields",
	acctParamsFieldNames[:],
	acctParamsFieldSpecByName,
}

// TypeNameDescriptions contains extra description about a low level
// protocol transaction Type string, and provide a friendlier type
// constant name in assembler.
var TypeNameDescriptions = map[string]string{
	string(protocol.UnknownTx):         "Unknown type. Invalid transaction",
	string(protocol.PaymentTx):         "Payment",
	string(protocol.KeyRegistrationTx): "KeyRegistration",
	string(protocol.AssetConfigTx):     "AssetConfig",
	string(protocol.AssetTransferTx):   "AssetTransfer",
	string(protocol.AssetFreezeTx):     "AssetFreeze",
	string(protocol.ApplicationCallTx): "ApplicationCall",
}

func init() {
	equal := func(x int, y int) {
		if x != y {
			panic(fmt.Sprintf("%d != %d", x, y))
		}
	}

	equal(len(txnFieldSpecs), len(TxnFieldNames))
	for i, s := range txnFieldSpecs {
		equal(int(s.field), i)
		TxnFieldNames[s.field] = s.field.String()
		txnFieldSpecByName[s.field.String()] = s
	}

	equal(len(globalFieldSpecs), len(GlobalFieldNames))
	for i, s := range globalFieldSpecs {
		equal(int(s.field), i)
		GlobalFieldNames[s.field] = s.field.String()
		globalFieldSpecByName[s.field.String()] = s
	}

	equal(len(ecdsaCurveSpecs), len(ecdsaCurveNames))
	for i, s := range ecdsaCurveSpecs {
		equal(int(s.field), i)
		ecdsaCurveNames[s.field] = s.field.String()
		ecdsaCurveSpecByName[s.field.String()] = s
	}

	equal(len(ecGroupSpecs), len(ecGroupNames))
	for i, s := range ecGroupSpecs {
		equal(int(s.field), i)
		ecGroupNames[s.field] = s.field.String()
		ecGroupSpecByName[s.field.String()] = s
	}

	equal(len(base64EncodingSpecs), len(base64EncodingNames))
	for i, s := range base64EncodingSpecs {
		equal(int(s.field), i)
		base64EncodingNames[i] = s.field.String()
		base64EncodingSpecByName[s.field.String()] = s
	}

	equal(len(jsonRefSpecs), len(jsonRefTypeNames))
	for i, s := range jsonRefSpecs {
		equal(int(s.field), i)
		jsonRefTypeNames[i] = s.field.String()
		jsonRefSpecByName[s.field.String()] = s
	}

	equal(len(vrfStandardSpecs), len(vrfStandardNames))
	for i, s := range vrfStandardSpecs {
		equal(int(s.field), i)
		vrfStandardNames[i] = s.field.String()
		vrfStandardSpecByName[s.field.String()] = s
	}

	equal(len(blockFieldSpecs), len(blockFieldNames))
	for i, s := range blockFieldSpecs {
		equal(int(s.field), i)
		blockFieldNames[i] = s.field.String()
		blockFieldSpecByName[s.field.String()] = s
	}

	equal(len(assetHoldingFieldSpecs), len(assetHoldingFieldNames))
	for i, s := range assetHoldingFieldSpecs {
		equal(int(s.field), i)
		assetHoldingFieldNames[i] = s.field.String()
		assetHoldingFieldSpecByName[s.field.String()] = s
	}

	equal(len(assetParamsFieldSpecs), len(assetParamsFieldNames))
	for i, s := range assetParamsFieldSpecs {
		equal(int(s.field), i)
		assetParamsFieldNames[i] = s.field.String()
		assetParamsFieldSpecByName[s.field.String()] = s
	}

	equal(len(appParamsFieldSpecs), len(appParamsFieldNames))
	for i, s := range appParamsFieldSpecs {
		equal(int(s.field), i)
		appParamsFieldNames[i] = s.field.String()
		appParamsFieldSpecByName[s.field.String()] = s
	}

	equal(len(acctParamsFieldSpecs), len(acctParamsFieldNames))
	for i, s := range acctParamsFieldSpecs {
		equal(int(s.field), i)
		acctParamsFieldNames[i] = s.field.String()
		acctParamsFieldSpecByName[s.field.String()] = s
	}

	txnTypeMap = make(map[string]uint64)
	for i, tt := range TxnTypeNames {
		txnTypeMap[tt] = uint64(i)
	}
	for k, v := range TypeNameDescriptions {
		txnTypeMap[v] = txnTypeMap[k]
	}

	onCompletionMap = make(map[string]uint64, len(OnCompletionNames))
	for oc := NoOp; oc < invalidOnCompletionConst; oc++ {
		symbol := oc.String()
		OnCompletionNames[oc] = symbol
		onCompletionMap[symbol] = uint64(oc)
	}

}

func addExtra(original string, extra string) string {
	if len(original) == 0 {
		return extra
	}
	if len(extra) == 0 {
		return original
	}
	sep := ". "
	if original[len(original)-1] == '.' {
		sep = " "
	}
	return original + sep + extra
}

// !!!

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Sender-0]
	_ = x[Fee-1]
	_ = x[FirstValid-2]
	_ = x[FirstValidTime-3]
	_ = x[LastValid-4]
	_ = x[Note-5]
	_ = x[Lease-6]
	_ = x[Receiver-7]
	_ = x[Amount-8]
	_ = x[CloseRemainderTo-9]
	_ = x[VotePK-10]
	_ = x[SelectionPK-11]
	_ = x[VoteFirst-12]
	_ = x[VoteLast-13]
	_ = x[VoteKeyDilution-14]
	_ = x[Type-15]
	_ = x[TypeEnum-16]
	_ = x[XferAsset-17]
	_ = x[AssetAmount-18]
	_ = x[AssetSender-19]
	_ = x[AssetReceiver-20]
	_ = x[AssetCloseTo-21]
	_ = x[GroupIndex-22]
	_ = x[TxID-23]
	_ = x[ApplicationID-24]
	_ = x[OnCompletion-25]
	_ = x[ApplicationArgs-26]
	_ = x[NumAppArgs-27]
	_ = x[Accounts-28]
	_ = x[NumAccounts-29]
	_ = x[ApprovalProgram-30]
	_ = x[ClearStateProgram-31]
	_ = x[RekeyTo-32]
	_ = x[ConfigAsset-33]
	_ = x[ConfigAssetTotal-34]
	_ = x[ConfigAssetDecimals-35]
	_ = x[ConfigAssetDefaultFrozen-36]
	_ = x[ConfigAssetUnitName-37]
	_ = x[ConfigAssetName-38]
	_ = x[ConfigAssetURL-39]
	_ = x[ConfigAssetMetadataHash-40]
	_ = x[ConfigAssetManager-41]
	_ = x[ConfigAssetReserve-42]
	_ = x[ConfigAssetFreeze-43]
	_ = x[ConfigAssetClawback-44]
	_ = x[FreezeAsset-45]
	_ = x[FreezeAssetAccount-46]
	_ = x[FreezeAssetFrozen-47]
	_ = x[Assets-48]
	_ = x[NumAssets-49]
	_ = x[Applications-50]
	_ = x[NumApplications-51]
	_ = x[GlobalNumUint-52]
	_ = x[GlobalNumByteSlice-53]
	_ = x[LocalNumUint-54]
	_ = x[LocalNumByteSlice-55]
	_ = x[ExtraProgramPages-56]
	_ = x[Nonparticipation-57]
	_ = x[Logs-58]
	_ = x[NumLogs-59]
	_ = x[CreatedAssetID-60]
	_ = x[CreatedApplicationID-61]
	_ = x[LastLog-62]
	_ = x[StateProofPK-63]
	_ = x[ApprovalProgramPages-64]
	_ = x[NumApprovalProgramPages-65]
	_ = x[ClearStateProgramPages-66]
	_ = x[NumClearStateProgramPages-67]
	_ = x[invalidTxnField-68]
}

const _TxnField_name = "SenderFeeFirstValidFirstValidTimeLastValidNoteLeaseReceiverAmountCloseRemainderToVotePKSelectionPKVoteFirstVoteLastVoteKeyDilutionTypeTypeEnumXferAssetAssetAmountAssetSenderAssetReceiverAssetCloseToGroupIndexTxIDApplicationIDOnCompletionApplicationArgsNumAppArgsAccountsNumAccountsApprovalProgramClearStateProgramRekeyToConfigAssetConfigAssetTotalConfigAssetDecimalsConfigAssetDefaultFrozenConfigAssetUnitNameConfigAssetNameConfigAssetURLConfigAssetMetadataHashConfigAssetManagerConfigAssetReserveConfigAssetFreezeConfigAssetClawbackFreezeAssetFreezeAssetAccountFreezeAssetFrozenAssetsNumAssetsApplicationsNumApplicationsGlobalNumUintGlobalNumByteSliceLocalNumUintLocalNumByteSliceExtraProgramPagesNonparticipationLogsNumLogsCreatedAssetIDCreatedApplicationIDLastLogStateProofPKApprovalProgramPagesNumApprovalProgramPagesClearStateProgramPagesNumClearStateProgramPagesinvalidTxnField"

var _TxnField_index = [...]uint16{0, 6, 9, 19, 33, 42, 46, 51, 59, 65, 81, 87, 98, 107, 115, 130, 134, 142, 151, 162, 173, 186, 198, 208, 212, 225, 237, 252, 262, 270, 281, 296, 313, 320, 331, 347, 366, 390, 409, 424, 438, 461, 479, 497, 514, 533, 544, 562, 579, 585, 594, 606, 621, 634, 652, 664, 681, 698, 714, 718, 725, 739, 759, 766, 778, 798, 821, 843, 868, 883}

func (i TxnField) String() string {
	if i < 0 || i >= TxnField(len(_TxnField_index)-1) {
		return "TxnField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TxnField_name[_TxnField_index[i]:_TxnField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[MinTxnFee-0]
	_ = x[MinBalance-1]
	_ = x[MaxTxnLife-2]
	_ = x[ZeroAddress-3]
	_ = x[GroupSize-4]
	_ = x[LogicSigVersion-5]
	_ = x[Round-6]
	_ = x[LatestTimestamp-7]
	_ = x[CurrentApplicationID-8]
	_ = x[CreatorAddress-9]
	_ = x[CurrentApplicationAddress-10]
	_ = x[GroupID-11]
	_ = x[OpcodeBudget-12]
	_ = x[CallerApplicationID-13]
	_ = x[CallerApplicationAddress-14]
	_ = x[invalidGlobalField-15]
}

const _GlobalField_name = "MinTxnFeeMinBalanceMaxTxnLifeZeroAddressGroupSizeLogicSigVersionRoundLatestTimestampCurrentApplicationIDCreatorAddressCurrentApplicationAddressGroupIDOpcodeBudgetCallerApplicationIDCallerApplicationAddressinvalidGlobalField"

var _GlobalField_index = [...]uint8{0, 9, 19, 29, 40, 49, 64, 69, 84, 104, 118, 143, 150, 162, 181, 205, 223}

func (i GlobalField) String() string {
	if i >= GlobalField(len(_GlobalField_index)-1) {
		return "GlobalField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _GlobalField_name[_GlobalField_index[i]:_GlobalField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetTotal-0]
	_ = x[AssetDecimals-1]
	_ = x[AssetDefaultFrozen-2]
	_ = x[AssetUnitName-3]
	_ = x[AssetName-4]
	_ = x[AssetURL-5]
	_ = x[AssetMetadataHash-6]
	_ = x[AssetManager-7]
	_ = x[AssetReserve-8]
	_ = x[AssetFreeze-9]
	_ = x[AssetClawback-10]
	_ = x[AssetCreator-11]
	_ = x[invalidAssetParamsField-12]
}

const _AssetParamsField_name = "AssetTotalAssetDecimalsAssetDefaultFrozenAssetUnitNameAssetNameAssetURLAssetMetadataHashAssetManagerAssetReserveAssetFreezeAssetClawbackAssetCreatorinvalidAssetParamsField"

var _AssetParamsField_index = [...]uint8{0, 10, 23, 41, 54, 63, 71, 88, 100, 112, 123, 136, 148, 171}

func (i AssetParamsField) String() string {
	if i < 0 || i >= AssetParamsField(len(_AssetParamsField_index)-1) {
		return "AssetParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetParamsField_name[_AssetParamsField_index[i]:_AssetParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AppApprovalProgram-0]
	_ = x[AppClearStateProgram-1]
	_ = x[AppGlobalNumUint-2]
	_ = x[AppGlobalNumByteSlice-3]
	_ = x[AppLocalNumUint-4]
	_ = x[AppLocalNumByteSlice-5]
	_ = x[AppExtraProgramPages-6]
	_ = x[AppCreator-7]
	_ = x[AppAddress-8]
	_ = x[invalidAppParamsField-9]
}

const _AppParamsField_name = "AppApprovalProgramAppClearStateProgramAppGlobalNumUintAppGlobalNumByteSliceAppLocalNumUintAppLocalNumByteSliceAppExtraProgramPagesAppCreatorAppAddressinvalidAppParamsField"

var _AppParamsField_index = [...]uint8{0, 18, 38, 54, 75, 90, 110, 130, 140, 150, 171}

func (i AppParamsField) String() string {
	if i < 0 || i >= AppParamsField(len(_AppParamsField_index)-1) {
		return "AppParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AppParamsField_name[_AppParamsField_index[i]:_AppParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AcctBalance-0]
	_ = x[AcctMinBalance-1]
	_ = x[AcctAuthAddr-2]
	_ = x[AcctTotalNumUint-3]
	_ = x[AcctTotalNumByteSlice-4]
	_ = x[AcctTotalExtraAppPages-5]
	_ = x[AcctTotalAppsCreated-6]
	_ = x[AcctTotalAppsOptedIn-7]
	_ = x[AcctTotalAssetsCreated-8]
	_ = x[AcctTotalAssets-9]
	_ = x[AcctTotalBoxes-10]
	_ = x[AcctTotalBoxBytes-11]
	_ = x[invalidAcctParamsField-12]
}

const _AcctParamsField_name = "AcctBalanceAcctMinBalanceAcctAuthAddrAcctTotalNumUintAcctTotalNumByteSliceAcctTotalExtraAppPagesAcctTotalAppsCreatedAcctTotalAppsOptedInAcctTotalAssetsCreatedAcctTotalAssetsAcctTotalBoxesAcctTotalBoxBytesinvalidAcctParamsField"

var _AcctParamsField_index = [...]uint8{0, 11, 25, 37, 53, 74, 96, 116, 136, 158, 173, 187, 204, 226}

func (i AcctParamsField) String() string {
	if i < 0 || i >= AcctParamsField(len(_AcctParamsField_index)-1) {
		return "AcctParamsField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AcctParamsField_name[_AcctParamsField_index[i]:_AcctParamsField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AssetBalance-0]
	_ = x[AssetFrozen-1]
	_ = x[invalidAssetHoldingField-2]
}

const _AssetHoldingField_name = "AssetBalanceAssetFrozeninvalidAssetHoldingField"

var _AssetHoldingField_index = [...]uint8{0, 12, 23, 47}

func (i AssetHoldingField) String() string {
	if i < 0 || i >= AssetHoldingField(len(_AssetHoldingField_index)-1) {
		return "AssetHoldingField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AssetHoldingField_name[_AssetHoldingField_index[i]:_AssetHoldingField_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NoOp-0]
	_ = x[OptIn-1]
	_ = x[CloseOut-2]
	_ = x[ClearState-3]
	_ = x[UpdateApplication-4]
	_ = x[DeleteApplication-5]
}

const _OnCompletionConstType_name = "NoOpOptInCloseOutClearStateUpdateApplicationDeleteApplication"

var _OnCompletionConstType_index = [...]uint8{0, 4, 9, 17, 27, 44, 61}

func (i OnCompletionConstType) String() string {
	if i >= OnCompletionConstType(len(_OnCompletionConstType_index)-1) {
		return "OnCompletionConstType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _OnCompletionConstType_name[_OnCompletionConstType_index[i]:_OnCompletionConstType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Secp256k1-0]
	_ = x[Secp256r1-1]
	_ = x[invalidEcdsaCurve-2]
}

const _EcdsaCurve_name = "Secp256k1Secp256r1invalidEcdsaCurve"

var _EcdsaCurve_index = [...]uint8{0, 9, 18, 35}

func (i EcdsaCurve) String() string {
	if i < 0 || i >= EcdsaCurve(len(_EcdsaCurve_index)-1) {
		return "EcdsaCurve(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcdsaCurve_name[_EcdsaCurve_index[i]:_EcdsaCurve_index[i+1]]
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BN254_G1-0]
	_ = x[BN254_G2-1]
	_ = x[BLS12_381_G1-2]
	_ = x[BLS12_381_G2-3]
	_ = x[invalidEcgroup-4]
}

const _EcGroup_name = "BN254_G1BN254_G2BLS12_381_G1BLS12_381_G2invalidEcgroup"

var _EcGroup_index = [...]uint8{0, 8, 16, 28, 40, 54}

func (i EcGroup) String() string {
	if i < 0 || i >= EcGroup(len(_EcGroup_index)-1) {
		return "EcGroup(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EcGroup_name[_EcGroup_index[i]:_EcGroup_index[i+1]]
}

// Ecgroup is an enum for `ec_` opcodes
type EcGroup int

const (
	// BN254_G1 is the G1 group of BN254
	BN254_G1 EcGroup = iota
	// BN254_G2 is the G2 group of BN254
	BN254_G2
	// BL12_381_G1 specifies the G1 group of BLS 12-381
	BLS12_381_G1
	// BL12_381_G2 specifies the G2 group of BLS 12-381
	BLS12_381_G2
	invalidEcgroup // compile-time constant for number of fields
)

var ecGroupNames [invalidEcgroup]string

type ecGroupSpec struct {
	field EcGroup
	doc   string
}

func (fs ecGroupSpec) Field() byte {
	return byte(fs.field)
}
func (fs ecGroupSpec) Type() StackType {
	return StackNone // Will not show, since all are untyped
}
func (fs ecGroupSpec) OpVersion() uint64 {
	return pairingVersion
}
func (fs ecGroupSpec) Version() uint64 {
	return pairingVersion
}
func (fs ecGroupSpec) Note() string {
	return fs.doc
}

var ecGroupSpecs = [...]ecGroupSpec{
	{BN254_G1, "G1 of the BN254 curve. Points encoded as 32 byte X following by 32 byte Y"},
	{BN254_G2, "G2 of the BN254 curve. Points encoded as 64 byte X following by 64 byte Y"},
	{BLS12_381_G1, "G1 of the BLS 12-381 curve. Points encoded as 48 byte X following by 48 byte Y"},
	{BLS12_381_G2, "G2 of the BLS 12-381 curve. Points encoded as 96 byte X following by 48 byte Y"},
}

func ecGroupSpecByField(c EcGroup) (ecGroupSpec, bool) {
	if int(c) >= len(ecGroupSpecs) {
		return ecGroupSpec{}, false
	}
	return ecGroupSpecs[c], true
}

var ecGroupSpecByName = make(ecGroupNameSpecMap, len(ecGroupNames))

type ecGroupNameSpecMap map[string]ecGroupSpec

func (s ecGroupNameSpecMap) get(name string) (FieldSpec, bool) {
	fs, ok := s[name]
	return fs, ok
}

// EcGroups collects details about the constants used to describe EcGroups
var EcGroups = FieldGroup{
	"EC", "Groups",
	ecGroupNames[:],
	ecGroupSpecByName,
}

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[URLEncoding-0]
	_ = x[StdEncoding-1]
	_ = x[invalidBase64Encoding-2]
}

const _Base64Encoding_name = "URLEncodingStdEncodinginvalidBase64Encoding"

var _Base64Encoding_index = [...]uint8{0, 11, 22, 43}

func (i Base64Encoding) String() string {
	if i < 0 || i >= Base64Encoding(len(_Base64Encoding_index)-1) {
		return "Base64Encoding(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Base64Encoding_name[_Base64Encoding_index[i]:_Base64Encoding_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[JSONString-0]
	_ = x[JSONUint64-1]
	_ = x[JSONObject-2]
	_ = x[invalidJSONRefType-3]
}

const _JSONRefType_name = "JSONStringJSONUint64JSONObjectinvalidJSONRefType"

var _JSONRefType_index = [...]uint8{0, 10, 20, 30, 48}

func (i JSONRefType) String() string {
	if i < 0 || i >= JSONRefType(len(_JSONRefType_index)-1) {
		return "JSONRefType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _JSONRefType_name[_JSONRefType_index[i]:_JSONRefType_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[VrfAlgorand-0]
	_ = x[invalidVrfStandard-1]
}

const _VrfStandard_name = "VrfAlgorandinvalidVrfStandard"

var _VrfStandard_index = [...]uint8{0, 11, 29}

func (i VrfStandard) String() string {
	if i < 0 || i >= VrfStandard(len(_VrfStandard_index)-1) {
		return "VrfStandard(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VrfStandard_name[_VrfStandard_index[i]:_VrfStandard_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BlkSeed-0]
	_ = x[BlkTimestamp-1]
	_ = x[invalidBlockField-2]
}

const _BlockField_name = "BlkSeedBlkTimestampinvalidBlockField"

var _BlockField_index = [...]uint8{0, 7, 19, 36}

func (i BlockField) String() string {
	if i < 0 || i >= BlockField(len(_BlockField_index)-1) {
		return "BlockField(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BlockField_name[_BlockField_index[i]:_BlockField_index[i+1]]
}

var ecdsaVerifyCosts = []int{
	Secp256k1: 1700,
	Secp256r1: 2500,
}

var ecdsaDecompressCosts = []int{
	Secp256k1: 650,
	Secp256r1: 2400,
}
