package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos/wasm"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestContract(t *testing.T) {
	t.Parallel()

	// Create chain factory with Juno
	numVals := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:      "juno",
			Version:   "latest",
			ChainName: "juno1",
			ChainConfig: ibc.ChainConfig{
				GasPrices:      "0ujuno",
				GasAdjustment:  2.0,
				EncodingConfig: wasm.WasmEncoding(),
			},
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	juno := chains[0].(*cosmos.CosmosChain)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().AddChain(juno)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
		// This can be used to write to the block database which will index all block data e.g. txs, msgs, events, etc.
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
	})
	require.NoError(t, err)

	// User Setup
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	user := users[0]
	keyname := user.KeyName
	uaddr := user.Bech32Address("juno")

	user2 := users[1]
	uaddr2 := user2.Bech32Address("juno")

	// Contract Testing
	codeId, err := juno.StoreContract(ctx, keyname, "../artifacts/journaling.wasm")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "1", codeId)

	contract, err := juno.InstantiateContract(ctx, keyname, codeId, fmt.Sprintf(`{"manager":"%s","allowed_submitters":["%s"]}`, uaddr, uaddr), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(contract)

	// Execute on the chain and add an entry for a user
	msg := fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Apr-26-2023", "My title here", "https://reece.sh", "note")
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	if err != nil {
		t.Fatal(err)
	}

	var jer JournalEntriesResponse
	if err := juno.QueryContract(ctx, contract, QueryMsg{GetEntries: &GetEntries{Address: uaddr}}, &jer); err != nil {
		t.Fatal(err)
	}
	for k, v := range *jer.Data {
		t.Log(k, v)
	}

	// Add whitelist for a new user uaddr2
	msg = fmt.Sprintf(`{"whitelist":{"address":"%s"}}`, uaddr2)
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	if err != nil {
		t.Fatal(err)
	}

	var resp WhitelistResponse
	if err := juno.QueryContract(ctx, contract, QueryMsg{GetWhitelist: &struct{}{}}, &resp); err != nil {
		t.Fatal(err)
	}
	t.Log("\nWhitelistResponse-> " + strings.Join(resp.Data, ","))

	// == submit another entry ==
	msg = fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Apr-26-2023", "2nd title", "github.com/2", "")
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	if err != nil {
		t.Fatal(err)
	}
	if err := juno.QueryContract(ctx, contract, QueryMsg{GetEntries: &GetEntries{Address: uaddr}}, &jer); err != nil {
		t.Fatal(err)
	}
	for k, v := range *jer.Data {
		t.Log(k, v)
	}

	// Final Cleanup
	t.Cleanup(func() {
		_ = ic.Close()
	})
}
