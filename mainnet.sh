# uploads the contract to mainnet (from reece's validator)
# TODO: Migrate to juno for this new contract

export JOED_NODE="https://joe-rpc.polkachu.com:443"
JUNOD_COMMAND_ARGS="--gas 5000000 --gas-prices=0.0025ujoe --from reece --broadcast-mode block --chain-id joe-1 --node $JOED_NODE --from validator --yes"
KEY_ADDR="joe1hj5fveer5cjtn4wd6wstzugjfdxzl0xp0cyvu4" # validator key from validator

joed status

alias BINARY="joed"

UPLOAD=$(joed tx wasm store artifacts/reviews.wasm $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $UPLOAD
BASE_CODE_ID=$(joed q tx $UPLOAD --output json | jq -r '.logs[0].events[] | select(.type == "store_code").attributes[] | select(.key == "code_id").value') && echo "Code Id: $BASE_CODE_ID"
# BASE_CODE_ID=34


TX_HASH=$(joed tx wasm instantiate "$BASE_CODE_ID" '{"max_saved":50,"cooldown_blocks":25}' --label "joe-reviews" $JUNOD_COMMAND_ARGS --admin "$KEY_ADDR" | jq -r '.txhash') && echo $TX_HASH

export CONTRACT=$(joed query tx $TX_HASH --output json | jq -r '.logs[0].events[0].attributes[0].value') && echo "CONTRACT: $CONTRACT"

# joe1nxd9d6pt7e9l2k6znclr0hkn95f60fganywkza3nqf9cvsnhzgds4yd2md

joed q wasm contract-state smart $CONTRACT '{"get_reviews":{}}'