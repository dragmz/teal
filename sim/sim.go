package sim

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/common/models"
	"github.com/algorand/go-algorand-sdk/v2/crypto"
	"github.com/algorand/go-algorand-sdk/v2/encoding/json"
	"github.com/algorand/go-algorand-sdk/v2/logic"
	"github.com/algorand/go-algorand-sdk/v2/transaction"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/pkg/errors"
)

type ProgramExecutionTrace struct {
	PC   int
	Line int
	Text string
}

type ProgramExecution struct {
	Approval []ProgramExecutionTrace
	Clear    []ProgramExecutionTrace

	Inner []ProgramExecution
}

type args struct {
	Algod      string
	AlgodToken string
	Sender     string

	Approval string
	Clear    string
}

type program struct {
	lines []string
	sm    logic.SourceMap
}

func (p program) translateUnit(unit []models.SimulationOpcodeTraceUnit) ([]ProgramExecutionTrace, error) {
	r := []ProgramExecutionTrace{}

	for _, u := range unit {
		line := p.sm.PcToLine[int(u.Pc)]

		r = append(r, ProgramExecutionTrace{
			PC:   int(u.Pc),
			Line: line,
			Text: p.lines[line],
		})
	}

	return r, nil
}

func (p program) translateTrace(trace models.SimulationTransactionExecTrace) (ProgramExecution, error) {
	var e ProgramExecution

	approval, err := p.translateUnit(trace.ApprovalProgramTrace)
	if err != nil {
		return e, errors.Wrap(err, "failed to translate approval program trace")
	}

	clear, err := p.translateUnit(trace.ClearStateProgramTrace)
	if err != nil {
		return e, errors.Wrap(err, "failed to translate clear program trace")
	}

	inner := make([]ProgramExecution, len(trace.InnerTrace))

	for i, inn := range trace.InnerTrace {
		inner[i], err = p.translateTrace(inn)
		if err != nil {
			return e, errors.Wrap(err, "failed to translate inner trace")
		}
	}

	e = ProgramExecution{
		Approval: approval,
		Clear:    clear,
		Inner:    inner,
	}

	return e, nil
}

type Result struct {
	Create ProgramExecution
	Call   ProgramExecution
}

func Run(ac *algod.Client, sender string, approval []byte, clear []byte) (Result, error) {
	var r Result

	ar, err := ac.TealCompile([]byte(approval)).Sourcemap(true).Do(context.Background())
	if err != nil {
		return r, errors.Wrap(err, "failed to compile approval program")
	}

	cr, err := ac.TealCompile([]byte(clear)).Sourcemap(true).Do(context.Background())
	if err != nil {
		return r, errors.Wrap(err, "failed to compile clear program")
	}

	sp, err := ac.SuggestedParams().Do(context.Background())
	if err != nil {
		return r, errors.Wrap(err, "failed to get suggested params")
	}

	addr, err := types.DecodeAddress(sender)
	if err != nil {
		return r, errors.Wrap(err, "failed to decode sender address")
	}

	acbs, err := base64.StdEncoding.DecodeString(ar.Result)
	if err != nil {
		return r, errors.Wrap(err, "failed to decode approval program")
	}

	ccbs, err := base64.StdEncoding.DecodeString(cr.Result)
	if err != nil {
		return r, errors.Wrap(err, "failed to decode clear program")
	}

	appId := uint64(340618153)
	appAddr := crypto.GetApplicationAddress(appId)

	paytx, err := transaction.MakePaymentTxn(addr.String(), appAddr.String(), 1000000, nil, "", sp)
	if err != nil {
		return r, errors.Wrap(err, "failed to make payment transaction")
	}

	calltx, err := transaction.MakeApplicationCallTx(appId,
		nil, nil, nil, nil, types.NoOpOC, nil, nil,
		types.StateSchema{}, types.StateSchema{},
		sp, addr, nil, types.Digest{}, [32]byte{}, types.ZeroAddress)
	if err != nil {
		return r, errors.Wrap(err, "failed to make application create transaction")
	}

	metatx, err := transaction.MakeApplicationCreateTx(false, acbs, ccbs,
		types.StateSchema{}, types.StateSchema{},
		nil, nil, nil, nil, sp, addr, nil, types.Digest{}, [32]byte{}, types.ZeroAddress)
	if err != nil {
		return r, errors.Wrap(err, "failed to make meta transaction")
	}

	group, err := crypto.ComputeGroupID([]types.Transaction{paytx, calltx, metatx})
	if err != nil {
		return r, errors.Wrap(err, "failed to compute group id")
	}

	paytx.Group = group
	calltx.Group = group
	metatx.Group = group

	paystx := types.SignedTxn{Txn: paytx}
	callstx := types.SignedTxn{Txn: calltx}
	metastx := types.SignedTxn{Txn: metatx}

	sr, err := ac.SimulateTransaction(models.SimulateRequest{
		AllowEmptySignatures: true,
		ExecTraceConfig: models.SimulateTraceConfig{
			Enable: true,
		},
		TxnGroups: []models.SimulateRequestTransactionGroup{
			{
				Txns: []types.SignedTxn{paystx, callstx, metastx},
			},
		},
	}).Do(context.Background())
	if err != nil {
		return r, errors.Wrap(err, "failed to simulate transaction")
	}

	fmt.Println(string(json.Encode(sr)))

	lines := strings.Split(string(approval), "\n")

	sm, err := logic.DecodeSourceMap(*ar.Sourcemap)
	if err != nil {
		return r, errors.Wrap(err, "failed to decode source map")
	}

	p := program{
		lines: lines,
		sm:    sm,
	}

	res := sr.TxnGroups[0].TxnResults[1]

	t, err := p.translateTrace(res.ExecTrace)
	if err != nil {
		return r, errors.Wrap(err, "failed to translate trace")
	}

	if len(t.Inner) > 0 {
		r.Create = t.Inner[0]
	}

	if len(t.Inner) > 1 {
		r.Call = t.Inner[1]
	}

	return r, nil
}
