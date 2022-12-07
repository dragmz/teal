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
