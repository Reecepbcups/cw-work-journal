package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/stretchr/testify/require"
)

// re-arrange these variables to make more sense
func GetAddressesEntries(chain *cosmos.CosmosChain, ctx context.Context, contract string, uaddr string, t *testing.T) *JournalEntriesResponse {
	var jer JournalEntriesResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetEntries: &GetEntries{Address: uaddr}}, &jer)
	require.NoError(t, err)
	// for k, v := range *jer.Data {
	// 	t.Log(k, v)
	// }
	return &jer
}

func GetWhitelistAddresses(chain *cosmos.CosmosChain, ctx context.Context, contract string, t *testing.T) *WhitelistResponse {
	var resp WhitelistResponse
	err := chain.QueryContract(ctx, contract, QueryMsg{GetWhitelist: &struct{}{}}, &resp)
	require.NoError(t, err)
	t.Log("\n\nWhitelistResponse-> " + strings.Join(resp.Data, ","))
	return &resp
}

func TestWhitelist(t *testing.T) {
	t.Parallel()

	// Create chain factory with Juno
	chains := CreateBaseChain(t)
	juno := chains[0].(*cosmos.CosmosChain)

	// Builds the chain for testing
	ic, ctx, _, _ := BuildInitialChain(t, chains)

	// User Setup
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", int64(10_000_000), juno, juno)
	user := users[0]
	keyname := user.KeyName
	uaddr := user.Bech32Address("juno")

	user2 := users[1]
	keyname2 := user2.KeyName
	uaddr2 := user2.Bech32Address("juno")

	// Contract Testing
	codeId, err := juno.StoreContract(ctx, keyname, "../artifacts/journaling.wasm")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "1", codeId)

	contract, err := juno.InstantiateContract(ctx, keyname, codeId, fmt.Sprintf(`{"manager":"%s","allowed_submitters":[]}`, uaddr), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(contract)

	// Try to submit by not being on the whtelist
	msg := fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Jan-1-2000", "T1", "LINK1", "NOTE1")
	_, err = juno.ExecuteContract(ctx, keyname2, contract, msg)
	require.NoError(t, err)

	// TODO: Remove contracts/work-journal/src/contract.rs:113:27 unwrap()
	// ensure there are no keys in GetAddressesEntries
	t.Log(*GetAddressesEntries(juno, ctx, contract, uaddr2, t).Data)
	require.Equal(t, 0, len(*GetAddressesEntries(juno, ctx, contract, uaddr2, t).Data))

	// Add whitelist for a new user uaddr2
	msg = fmt.Sprintf(`{"whitelist":{"address":"%s"}}`, uaddr2)
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	require.NoError(t, err)

	// Get allowed submitters and ensure the query returns the correct amount
	wle := GetWhitelistAddresses(juno, ctx, contract, t)
	require.Equal(t, 1, len(wle.Data))

	// Try to submit again
	msg = fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Jan-1-2000", "T1", "LINK1", "NOTE1")
	_, err = juno.ExecuteContract(ctx, keyname2, contract, msg)
	require.NoError(t, err)
	// ensure they have length of 1 key
	require.Equal(t, 1, len(*GetAddressesEntries(juno, ctx, contract, uaddr2, t).Data))

	// remove whitelist for a new user uaddr2
	msg = fmt.Sprintf(`{"remove":{"address":"%s"}}`, uaddr2)
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	require.NoError(t, err)

	// Get allowed submitters and ensure the query returns the correct amount
	wle = GetWhitelistAddresses(juno, ctx, contract, t)
	require.Equal(t, 0, len(wle.Data))

	// try to submit again, will ot work for user2
	msg = fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Jan-1-2000", "T1", "LINK1", "NOTE1")
	_, err = juno.ExecuteContract(ctx, keyname2, contract, msg)
	require.NoError(t, err)
	// still 1
	require.Equal(t, 1, len(*GetAddressesEntries(juno, ctx, contract, uaddr2, t).Data))

	// var resp WhitelistResponse
	// err = juno.QueryContract(ctx, contract, QueryMsg{GetWhitelist: &struct{}{}}, &resp)
	// t.Log("\nWhitelistResponse-> " + strings.Join(resp.Data, ","))
	// require.Contains(t, resp.Data, uaddr2)

	// Final Cleanup
	t.Cleanup(func() {
		_ = ic.Close()
	})
}
