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

func (e parseError) Rule() string {
	return "PARSE"
}

type lintError struct {
	error

	l int // line index
	b int
	e int

	s DiagnosticSeverity

	r string
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

func (e lintError) Rule() string {
	return e.r
}

func readInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return int8(v), nil
}

func readBool(s string) (bool, error) {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, err
	}

	return v, nil
}

func readUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, err
	}

	return uint8(v), nil
}

func readAssetHoldingField(v uint64, s string) (AssetHoldingField, bool, error) {
	spec, ok := assetHoldingFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return AssetHoldingField(value), false, nil
}

func readVrfVerifyField(v uint64, s string) (VrfStandard, bool, error) {
	spec, ok := vrfStandardSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return VrfStandard(value), false, nil
}

func readBlockField(v uint64, s string) (BlockField, bool, error) {
	spec, ok := blockFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return BlockField(value), false, nil
}

func readVoterParams(v uint64, s string) (VoterParamsField, bool, error) {
	spec, ok := voterParamsFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return VoterParamsField(value), false, nil
}

func readMimcField(v uint64, s string) (MimcConfig, bool, error) {
	spec, ok := mimcConfigSpecByName[s]
	if ok {
		needed := spec.Version()
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse mimc field")
	}

	return MimcConfig(value), false, nil
}

func readVoterParamsField(v uint64, s string) (VoterParamsField, bool, error) {
	spec, ok := voterParamsFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return VoterParamsField(value), false, nil
}

func readAcctParams(v uint64, s string) (AcctParamsField, bool, error) {
	spec, ok := acctParamsFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return AcctParamsField(value), false, nil
}

func readAppParamsField(v uint64, s string) (AppParamsField, bool, error) {
	spec, ok := appParamsFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return AppParamsField(value), false, nil
}

func readAssetParamsField(v uint64, s string) (AssetParamsField, bool, error) {
	spec, ok := assetParamsFieldSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return AssetParamsField(value), false, nil
}

func readGlobalField(v uint64, s string, mode RunMode) (GlobalField, bool, error) {
	spec, ok := globalFieldSpecByName[s]
	if ok {
		if spec.mode != ModeAny {
			if spec.mode != mode {
				return 0, true, errors.Errorf("not available in this mode (need: %s, got: %s)", spec.mode, mode)
			}
		}

		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return GlobalField(value), false, nil
}

func readEcGroupField(v uint64, s string) (EcGroup, bool, error) {
	spec, ok := ecGroupSpecByName[s]
	if ok {
		needed := spec.Version()
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return EcGroup(value), false, nil
}

func readBase64EncodingField(v uint64, s string) (Base64Encoding, bool, error) {
	spec, ok := base64EncodingSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return Base64Encoding(value), false, nil
}

func readJsonRefField(v uint64, s string) (JSONRefType, bool, error) {
	spec, ok := jsonRefSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return JSONRefType(value), false, nil
}

func readTxnField(c fieldContext, v uint64, s string, m RunMode) (TxnField, bool, error) {
	spec, ok := txnFieldSpecByName[s]
	if ok {
		if spec.effects && m == ModeSig {
			return 0, true, errors.Errorf("not available in this mode (need: %s, got: %s)", ModeApp, m)
		}

		switch c {
		case txnaFieldContext:
			if !spec.array {
				return 0, true, errors.New("not available in array context")
			}
		}
		var needed uint64

		switch c {
		case txnFieldContext:
			needed = spec.version
		case itxnFieldContext:
			if spec.itxVersion == 0 {
				return 0, true, errors.New("not available for internal transactions")
			}
			needed = spec.itxVersion
		}

		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return TxnField(value), false, nil
}

func readInt(a *arguments) (uint64, error) {
	val, err := strconv.ParseUint(a.Text(), 0, 64)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func readConstInt(a *arguments) (uint64, error) {
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

func readEcdsaCurveIndex(v uint64, s string) (EcdsaCurve, bool, error) {
	spec, ok := ecdsaCurveSpecByName[s]
	if ok {
		needed := spec.version
		if needed > v {
			return 0, true, errors.Errorf("not available in this version (need >= %d, got: %d)", needed, v)
		}

		return spec.field, true, nil
	}

	value, err := readUint8(s)
	if err != nil {
		return 0, false, errors.Wrap(err, "failed to parse txn field")
	}

	return EcdsaCurve(value), false, nil
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
