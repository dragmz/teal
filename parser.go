package teal

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type ParserError struct {
	error
	l int
}

func (e ParserError) Line() int {
	return e.l
}

func (e ParserError) Error() string {
	return e.error.Error()
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

func readInt(s *bufio.Scanner) (uint64, error) {
	switch s.Text() {
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

	val, err := strconv.ParseUint(s.Text(), 0, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse uint64")
	}

	return val, nil
}

func readBytes(s *bufio.Scanner) ([]byte, error) {
	args := []string{s.Text()}
	for s.Scan() {
		args = append(args, s.Text())
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

func Parse(source string) (Program, []ParserError) {
	var res Program

	line := 0

	r := bufio.NewScanner(strings.NewReader(source))
	r.Split(bufio.ScanLines)

	var errs []ParserError

	for r.Scan() {
		line++
		err := func() error {
			var e Expr

			s := bufio.NewScanner(strings.NewReader(r.Text()))
			s.Split(bufio.ScanWords)

			s.Scan()
			if strings.HasPrefix(s.Text(), "//") {
				e = &CommentExpr{Text: r.Text()[2:]}
			} else if strings.HasSuffix(s.Text(), ":") {
				idx := strings.LastIndex(s.Text(), ":")
				name := s.Text()[:idx]
				e = &LabelExpr{Name: name}
			} else {
				switch s.Text() {
				case "":
					e = Empty
				case "#pragma":
					if !s.Scan() {
						return errors.New("missing pragma name")
					}

					switch s.Text() {
					case "version":
						if !s.Scan() {
							return errors.New("missing pragma version value")
						}

						version, err := strconv.Atoi(s.Text())
						if err != nil {
							return errors.Wrap(err, "failed to parse pragma version")
						}

						e = &PragmaExpr{Version: uint8(version)}
					default:
						return errors.Errorf("unexpected #pragma: %s", s.Text())
					}
				case "bnz":
					if !s.Scan() {
						return errors.New("missing bnz label name")
					}

					e = &BnzExpr{Label: &LabelExpr{Name: s.Text()}}
				case "bz":
					if !s.Scan() {
						return errors.New("missing bz label name")
					}

					e = &BzExpr{Label: &LabelExpr{Name: s.Text()}}
				case "b":
					if !s.Scan() {
						return errors.New("missing b label name")
					}

					e = &BExpr{Label: &LabelExpr{Name: s.Text()}}
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
					if !s.Scan() {
						return errors.New("missing ecdsa_verify curve index")
					}

					curve, err := readEcdsaCurveIndex(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read ecdsa_verify curve index")
					}
					e = &EcdsaVerifyExpr{Index: curve}
				case "ecdsa_pk_decompress":
					if !s.Scan() {
						return errors.New("missing ecdsa_pk_decompress curve index")
					}

					curve, err := readEcdsaCurveIndex(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read ecdsa_pk_decompress curve index")
					}
					e = &EcdsaPkDecompressExpr{Index: curve}
				case "ecdsa_pk_recover":
					if !s.Scan() {
						return errors.New("missing ecdsa_pk_recover curve index")
					}

					curve, err := readEcdsaCurveIndex(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read ecdsa_pk_recover curve index")
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

					for s.Scan() {
						if strings.HasPrefix(s.Text(), "//") {
							break
						}

						value, err := readInt(s)
						if err != nil {
							return errors.Wrap(err, "failed to read int")
						}

						values = append(values, value)
					}

					e = &IntcBlockExpr{Values: values}
				case "intc":
					if !s.Scan() {
						return errors.New("missing intc value")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse intc value")
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

					for s.Scan() {
						if strings.HasPrefix(s.Text(), "//") {
							break
						}

						b, err := readBytes(s)
						if err != nil {
							return errors.Wrap(err, "failed to read bytes")
						}

						values = append(values, b)
					}

					e = &BytecBlockExpr{Values: values}
				case "bytec":
					if !s.Scan() {
						return errors.New("missing bytec index")
					}
					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse bytec index")
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
					if !s.Scan() {
						return errors.New("missing arg index")
					}
					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse arg index")
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
					if !s.Scan() {
						return errors.New("missing gitxna t")
					}

					t, err := strconv.ParseUint(s.Text(), 10, 64)
					if err != nil {
						return errors.Wrapf(err, "failed to parse gitxna t: %s", s.Text())
					}

					if !s.Scan() {
						return errors.New("missing gitxna f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse gitxna f: %s", s.Text())
					}

					if !s.Scan() {
						return errors.New("missing gitxna i")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gitxna i")
					}

					e = &GitxnaExpr{Group: uint8(t), Field: f, Index: uint8(value)}
				case "gtxn":
					if !s.Scan() {
						return errors.New("missing gtxn t")
					}

					t, err := strconv.ParseUint(s.Text(), 10, 64)
					if err != nil {
						return errors.Wrapf(err, "failed to parse gtxn t: %s", s.Text())
					}

					if !s.Scan() {
						return errors.New("missing gtxn f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse gtxn f: %s", s.Text())
					}

					e = &GtxnExpr{Index: uint8(t), Field: f}
				case "txn":
					if !s.Scan() {
						return errors.New("missing txn field")
					}

					field, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse txn field: %s", s.Text())
					}

					// TODO: check the field value
					e = &TxnExpr{Field: field}
				case "global":
					if !s.Scan() {
						return errors.New("missing global field")
					}

					field, err := readGlobalField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse global field: %s", s.Text())
					}

					e = &GlobalExpr{Index: field}
				case "load":
					if !s.Scan() {
						return errors.New("missing load index")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse load i")
					}

					e = &LoadExpr{Index: uint8(value)}
				case "gload":
					if !s.Scan() {
						return errors.New("missing gload t")
					}

					t, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gload t")
					}

					if !s.Scan() {
						return errors.New("missing gload i")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gload i")
					}

					e = &GloadExpr{Group: uint8(t), Index: uint8(value)}
				case "gloads":
					if !s.Scan() {
						return errors.New("missing gloads i")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gloads i")
					}

					e = &GloadsExpr{Index: uint8(value)}
				case "store":
					if !s.Scan() {
						return errors.New("missing store i")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse store i")
					}

					e = &StoreExpr{Index: uint8(value)}
				case "txna":
					if !s.Scan() {
						return errors.New("missing txna f")
					}

					field, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse txna field: %s", s.Text())
					}

					e = &TxnaExpr{Field: field}
				case "gtxns":
					if !s.Scan() {
						return errors.New("missing gtxns f")
					}

					field, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse gtxns field: %s", s.Text())
					}

					e = &GtxnsExpr{Field: field}
				case "gaid":
					if !s.Scan() {
						return errors.New("missing gitxn t")
					}

					t, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gitxn t")
					}

					e = &GaidExpr{Group: uint8(t)}
				case "gtxna":
					if !s.Scan() {
						return errors.New("missing gitxn t")
					}

					t, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gitxn t")
					}

					if !s.Scan() {
						return errors.New("missing gtxnsa f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse gtxnsa f: %s", s.Text())
					}

					if !s.Scan() {
						return errors.New("missing gtxnsa i")
					}

					i, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gtxnsa i")
					}

					e = &GtxnaExpr{Group: uint8(t), Field: f, Index: uint8(i)}
				case "gtxnsa":
					if !s.Scan() {
						return errors.New("missing gtxnsa f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse gtxnsa f: %s", s.Text())
					}

					if !s.Scan() {
						return errors.New("missing gtxnsa i")
					}

					i, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gtxnsa i")
					}

					e = &GtxnsaExpr{Field: f, Index: uint8(i)}
				case "txnas":
					if !s.Scan() {
						return errors.New("missing txnas f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse txnas f: %s", s.Text())
					}

					e = &TxnasExpr{Field: f}
				case "extract":
					if !s.Scan() {
						return errors.New("missing extract s")
					}
					start, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read extract s")
					}

					if !s.Scan() {
						return errors.New("missing extract l")
					}

					length, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read extract l")
					}

					e = &ExtractExpr{Start: uint8(start), Length: uint8(length)}
				case "substring":
					if !s.Scan() {
						return errors.New("missing substract s")
					}
					start, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read substract s")
					}

					if !s.Scan() {
						return errors.New("missing substract e")
					}

					end, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read substract e")
					}

					e = &SubstringExpr{Start: uint8(start), End: uint8(end)}
				case "proto":
					if !s.Scan() {
						return errors.New("missing proto a")
					}
					a, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read proto a")
					}

					if !s.Scan() {
						return errors.New("missing proto r")
					}

					r, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read proto r")
					}

					e = &ProtoExpr{Args: uint8(a), Results: uint8(r)}
				case "byte":
					if !s.Scan() {
						return errors.New("missing byte value")
					}

					value, err := readBytes(s)
					if err != nil {
						return errors.Wrap(err, "failed to parse byte value")
					}
					e = &ByteExpr{Value: value}
				case "pushbytes":
					if !s.Scan() {
						return errors.New("missing pushbytes value")
					}

					value, err := readBytes(s)
					if err != nil {
						return errors.Wrap(err, "failed to parse pushbytes value")
					}
					e = &PushBytesExpr{Value: value}
				case "pushint":
					if !s.Scan() {
						return errors.New("missing pushint value")
					}

					value, err := readInt(s)
					if err != nil {
						return errors.Wrap(err, "failed to parse pushint value")
					}
					e = &PushIntExpr{Value: value}
				case "asset_params_get":
					if !s.Scan() {
						return errors.New("missing asset_params_get f")
					}

					field, err := readAssetField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse asset_params_get f: %s", s.Text())
					}

					e = &AssetParamsGetExpr{Field: field}
				case "int":
					if !s.Scan() {
						return errors.New("missing int value")
					}
					value, err := readInt(s)
					if err != nil {
						return errors.Wrap(err, "failed to read int value")
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
					if !s.Scan() {
						return errors.New("missing dig n")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read dig n")
					}

					e = &DigExpr{Index: uint8(value)}
				case "gtxnsas":
					if !s.Scan() {
						return errors.New("missing gtxnsas f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to read gtxnsas f: %s", s.Text())
					}

					e = &GtxnsasExpr{Field: f}
				case "gitxn":
					if !s.Scan() {
						return errors.New("missing gitxn t")
					}

					t, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse gitxn t")
					}

					if !s.Scan() {
						return errors.New("missing gtxn f")
					}

					f, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read gitxn f")
					}

					e = &GitxnExpr{Index: uint8(t), Field: f}
				case "asset_holding_get":
					if !s.Scan() {
						return errors.New("missing asset_holding_get f")
					}

					f, err := readAssetHoldingField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to read asset_holding_get f: %s", s.Text())
					}

					e = &AssetHoldingGetExpr{Field: f}
				case "acct_params_get":
					if !s.Scan() {
						return errors.New("missing acct_params_get f")
					}

					f, err := readAcctParams(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse acct_params_get f: %s", s.Text())
					}

					e = &AcctParamsGetExpr{Field: f}

				case "app_params_get":
					if !s.Scan() {
						return errors.New("missing app_params_get f")
					}

					f, err := readAppField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse app_params_get f: %s", s.Text())
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
					if !s.Scan() {
						return errors.New("missing vrf_verify f")
					}
					f, err := readVrfVerifyField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse vrf_verify f: %s", s.Text())
					}

					e = &VrfVerifyExpr{Field: f}
				case "block":
					if !s.Scan() {
						return errors.New("missing block f")
					}

					f, err := readBlockField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse block f: %s", s.Text())
					}

					e = &BlockExpr{Field: f}
				case "switch":
					var labels []*LabelExpr
					for s.Scan() {
						if strings.HasPrefix(s.Text(), "//") {
							break
						}
						labels = append(labels, &LabelExpr{Name: s.Text()})
					}
					e = &SwitchExpr{Labels: labels}
				case "match":
					var labels []*LabelExpr
					for s.Scan() {
						if strings.HasPrefix(s.Text(), "//") {
							break
						}

						labels = append(labels, &LabelExpr{Name: s.Text()})
					}
					e = &MatchExpr{Labels: labels}
				case "callsub":
					if !s.Scan() {
						return errors.New("missing callsub label name")
					}
					e = &CallSubExpr{Label: &LabelExpr{Name: s.Text()}}
				case "assert":
					e = Assert
				case "dup":
					e = Dup
				case "frame_bury":
					if !s.Scan() {
						return errors.New("missing frame_bury index")
					}
					value, err := readInt8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse frame_bury i")
					}
					e = &FrameBuryExpr{Index: value}
				case "frame_dig":
					if !s.Scan() {
						return errors.New("missing frame_dig index")
					}
					value, err := readInt8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse frame_dig i")
					}

					e = &FrameDigExpr{Index: value}
				case "setbyte":
					e = SetByte
				case "uncover":
					if !s.Scan() {
						return errors.New("missing uncover index")
					}
					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to parse uncover i")
					}
					e = &UncoverExpr{Index: uint8(value)}
				case "cover":
					if !s.Scan() {
						return errors.New("missing cover n")
					}

					value, err := readUint8(s.Text())
					if err != nil {
						return errors.Wrap(err, "failed to read cover n")
					}

					e = &CoverExpr{Index: uint8(value)}
				case "concat":
					e = Concat
				case "itxn_begin":
					e = ItxnBegin
				case "itxn_submit":
					e = ItxnSubmit
				case "itxn":
					if !s.Scan() {
						return errors.New("missing itxn f")
					}

					field, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to parse itxn f: %s", s.Text())
					}

					e = &ItxnExpr{Field: field}
				case "itxn_field":
					if !s.Scan() {
						return errors.New("missing itxn_field f")
					}

					field, err := readTxnField(s.Text())
					if err != nil {
						return errors.Wrapf(err, "failed to read itxn_field f: %s", s.Text())
					}

					e = &ItxnFieldExpr{Field: field}
				default:
					return errors.Errorf("unexpected opcode: %s", s.Text())
				}
			}
			res = append(res, e)

			return nil
		}()

		if err != nil {
			errs = append(errs, ParserError{error: err, l: line})
		}
	}

	return res, errs
}
