# CosmWasm Work Journal contact

**I don't have time for this, archiving.**

This is a contract that saves a work log of values for different approved users. The idea is similar to <https://work.reece.sh>.

```
Map<UserAddress -> Map<UniqueID, JournalEntry>>

UserAddress = juno1...

UniqueId - get last Id function, auto increment. Would be nice if there was an SQL tables type in cosmwasm

JournalEntry:
- date
- task
- repo_pr
- notes


whitelist
- list of allowed users. Only the contract admin / manager can add / remove.
```
