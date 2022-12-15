package teal

import (
	"strconv"

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

	l int // line index
	b int
	e int

	s DiagnosticSeverity
}

func (e lintError) Line() int {
	return e.l
}

func (e lintError) Begin() int {
	return e.b
}

func (e lintError) End() int {
	return e.e
}

func (e lintError) String() string {
	return e.error.Error()
}

func (e lintError) Severity() DiagnosticSeverity {
	return e.s
}

func readInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return int8(v), nil
}

func readUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return uint8(v), nil
}

func readAssetHoldingField(s string) (AssetHoldingField, error) {
	spec, ok := assetHoldingFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return AssetHoldingField(value), nil
}

func readVrfVerifyField(s string) (VrfStandard, error) {
	spec, ok := vrfStandardSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return VrfStandard(value), nil
}

func readBlockField(s string) (BlockField, error) {
	spec, ok := blockFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return BlockField(value), nil
}

func readAcctParams(s string) (AcctParamsField, error) {
	spec, ok := acctParamsFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return AcctParamsField(value), nil
}

func readAppParamsField(s string) (AppParamsField, error) {
	spec, ok := appParamsFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return AppParamsField(value), nil
}

func readAssetParamsField(s string) (AssetParamsField, error) {
	spec, ok := assetParamsFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return AssetParamsField(value), nil
}

func readGlobalField(s string) (GlobalField, error) {
	spec, ok := globalFieldSpecByName[s]
	if ok {
		return spec.field, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse txn field")
	}

	return GlobalField(value), nil
}

func readJsonRefField(s string) (JSONRefType, bool, error) {
	spec, ok := jsonRefSpecByName[s]
	if ok {
		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return JSONRefType(value), false, nil
}

func readTxnField(s string) (TxnField, bool, error) {
	spec, ok := txnFieldSpecByName[s]
	if ok {
		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return TxnField(value), false, nil
}

func readInt(a *arguments) (uint64, error) {
	i, ok := txnTypeMap[a.Text()]
	if ok {
		return i, nil
	}

	oc, ok := onCompletionMap[a.Text()]
	if ok {
		return oc, nil
	}

	val, err := strconv.ParseUint(a.Text(), 0, 64)
	if err != nil {
		return 0, err
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
	case Secp256k1:
	case Secp256r1:
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
