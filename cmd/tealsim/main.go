package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/dragmz/teal/sim"
	"github.com/pkg/errors"
)

type args struct {
	Algod      string
	AlgodToken string
	Sender     string

	Approval string
	Clear    string
}

func run(a args) error {
	ac, err := algod.MakeClient(a.Algod, a.AlgodToken)
	if err != nil {
		return errors.Wrap(err, "failed to make algod client")
	}

	abs, err := os.ReadFile(a.Approval)
	if err != nil {
		return errors.Wrap(err, "failed to read approval program")
	}

	cbs, err := os.ReadFile(a.Clear)
	if err != nil {
		return errors.Wrap(err, "failed to read clear program")
	}

	res, err := sim.Run(abs, cbs, sim.RunConfig{
		Sender: a.Sender,
		Ac:     ac,
		Create: sim.AppRunConfig{},
		Call:   sim.AppRunConfig{},
	})
	if err != nil {
		return errors.Wrap(err, "failed to run simulation")
	}

	fmt.Println("Create")

	for _, a := range res.Create.Approval {
		fmt.Printf("%d: %s\n", a.Line, a.Text)
	}

	fmt.Println("Call")

	for _, a := range res.Call.Approval {
		fmt.Printf("%d: %s\n", a.Line, a.Text)
	}

	return nil
}

func main() {
	var a args

	flag.StringVar(&a.Algod, "algod", "https://testnet-api.algonode.cloud", "algod endpoint")
	flag.StringVar(&a.AlgodToken, "algod-token", "", "algod token")
	flag.StringVar(&a.Sender, "sender", "F77YBQEP4EAJYCQPS4GYEW2WWJXU6DQ4OJHRYSV74UXHOTRWXYRN7HNP3U", "sender address")
	flag.StringVar(&a.Approval, "approval", ".data/approve.teal", "approval program")
	flag.StringVar(&a.Clear, "clear", ".data/clear.teal", "clear program")
	flag.Parse()

	err := run(a)
	if err != nil {
		panic(err)
	}
}
