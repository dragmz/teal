package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/dragmz/abs"
	"github.com/dragmz/teal"
	"github.com/pkg/errors"
)

type args struct {
	Algod      string
	AlgodToken string

	DevAlgod      string
	DevAlgodToken string

	Round uint64
}

func run(a args) error {
	ac, err := algod.MakeClient(a.Algod, a.AlgodToken)
	if err != nil {
		return errors.Wrap(err, "failed to make algod client")
	}

	dac, err := algod.MakeClient(a.DevAlgod, a.DevAlgodToken)
	if err != nil {
		return errors.Wrap(err, "failed to make dev algod client")
	}

	if a.Round == 0 {
		status, err := ac.Status().Do(context.Background())
		if err != nil {
			return errors.Wrap(err, "failed to get status")
		}

		a.Round = status.LastRound
	}

	blocks, err := abs.MakeBlocks(ac, abs.WithRetry(time.Second))
	if err != nil {
		return errors.Wrap(err, "failed to make blocks stream")
	}

	ch := make(chan types.Block)

	go func() {
		for b := range ch {
			fmt.Printf("Block: %d at %s\n", b.Round, time.Now())
			for txidx, tx := range b.Payset {
				if tx.Txn.Type != "appl" {
					continue
				}
				if tx.Txn.ApprovalProgram == nil {
					continue
				}
				err := func() error {
					fmt.Println("Program length:", len(tx.Txn.ApprovalProgram))

					resp, err := dac.TealDisassemble(tx.Txn.ApprovalProgram).Do(context.Background())
					if err != nil {
						return errors.Wrap(err, "failed to disassemble")
					}

					res := teal.Process(resp.Result)
					if len(res.Diagnostics) > 0 {
						for _, err := range res.Diagnostics {
							fmt.Printf("%d:%d:%d: %s\n", b.Round, txidx, err.Line(), err)
						}
					}

					return nil
				}()

				if err != nil {
					fmt.Printf("Failed to process app - err: %s\n", err)
				}
			}
		}
	}()

	err = blocks.Stream(context.Background(), a.Round, ch)
	if err != nil {
		return errors.Wrap(err, "failed to stream blocks")
	}

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.Algod, "algod", "https://mainnet-api.algonode.network", "algod address")
	flag.StringVar(&a.AlgodToken, "algod-token", "", "algod token")

	flag.StringVar(&a.DevAlgod, "dev-algod", "", "dev algod token")
	flag.StringVar(&a.DevAlgodToken, "dev-algod-token", "", "dev algod token")

	flag.Uint64Var(&a.Round, "round", 0, "first round to process")

	flag.Parse()

	err := run(a)

	if err != nil {
		panic(err)
	}
}
