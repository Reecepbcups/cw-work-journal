use std::collections::BTreeMap;

use cosmwasm_schema::cw_serde;
use cw_storage_plus::Item;

use crate::msg::JournalEntry;

#[cw_serde]
pub struct State {
    pub manager: String,    
    pub allowed_submitters: Vec<String>,    
}

// move this


#[cw_serde]
pub struct DataEntries {    
    pub entries: BTreeMap<String, BTreeMap<u128, JournalEntry>>, 
}



pub const STATE: Item<State> = Item::new("state");

pub const DATA_STATE: Item<DataEntries> = Item::new("data_entries");
