package test

import (
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestContract(t *testing.T) {
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

	// user2 := users[1]
	// uaddr2 := user2.Bech32Address("juno")

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

	// var jer JournalEntriesResponse
	// if err := juno.QueryContract(ctx, contract, QueryMsg{GetEntries: &GetEntries{Address: uaddr}}, &jer); err != nil {
	// 	t.Fatal(err)
	// }
	// for k, v := range *jer.Data {
	// 	t.Log(k, v)
	// }
	res := GetAddressesEntries(juno, ctx, contract, uaddr, t)
	for k, v := range *res.Data {
		t.Log(k, v)
	}

	// Add whitelist for a new user uaddr2
	// msg = fmt.Sprintf(`{"whitelist":{"address":"%s"}}`, uaddr2)
	// _, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// var resp WhitelistResponse
	// if err := juno.QueryContract(ctx, contract, QueryMsg{GetWhitelist: &struct{}{}}, &resp); err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log("\nWhitelistResponse-> " + strings.Join(resp.Data, ","))

	// == submit another entry ==
	msg = fmt.Sprintf(`{"submit":{"entries":[{"date":"%s","title":"%s","repo_pr":"%s","notes":"%s"}]}}`, "Apr-26-2023", "2nd title", "github.com/2", "")
	_, err = juno.ExecuteContract(ctx, keyname, contract, msg)
	if err != nil {
		t.Fatal(err)
	}
	res = GetAddressesEntries(juno, ctx, contract, uaddr, t)
	for k, v := range *res.Data {
		t.Log(k, v)
	}

	// Final Cleanup
	t.Cleanup(func() {
		_ = ic.Close()
	})
}
