use cosmwasm_schema::{cw_serde, QueryResponses};

// use crate::state::State;

#[cw_serde]
pub struct InstantiateMsg {
    pub manager: Option<String>,
    pub allowed_submitters: Vec<String>,
}

#[cw_serde]
pub struct JournalEntry {
    pub date: String,
    pub title: String,
    pub repo_pr: String,
    pub notes: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    // admin only
    /// Whitelist an address which can submit entries into their journal.
    Whitelist { address: String },
    /// Remove an address from the whitelist.
    /// This will not remove any existing entries from the journal.
    Remove { address: String },

    // Users
    /// Submit a new entry into the journal.    
    Submit { entries: Vec<JournalEntry> },
    // remove entry by unique id
    // modify entry by ID
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(Vec<JournalEntry>)]
    GetEntries {
        address: String,
        // range, use pagination here
    },
    #[returns(JournalEntry)]
    GetSpecificEntry { address: String, id: u128 },
    #[returns(Vec<String>)]
    GetWhitelist {},
}
