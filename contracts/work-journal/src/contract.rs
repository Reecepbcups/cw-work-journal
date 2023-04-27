use std::collections::BTreeMap;

#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{State, DataEntries, DATA_STATE, STATE};

// version info for migration info
const CONTRACT_NAME: &str = "crates.io:journal";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    // save initial state
    let state = State {
        allowed_submitters: msg.allowed_submitters,
        manager: msg.manager.unwrap_or_else(|| _info.sender.to_string()),
    };
    STATE.save(deps.storage, &state)?;

    // Init date entries
    let data_state = DataEntries {
        entries: BTreeMap::new(),
    };
    DATA_STATE.save(deps.storage, &data_state)?;

    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        // Admin Only
        ExecuteMsg::Whitelist { address } => {            
            let mut state: State = STATE.load(deps.storage)?;        
            if !state.allowed_submitters.contains(&address) {
                state.allowed_submitters.push(address);
            }

            STATE.save(deps.storage, &state)?;
            Ok(Response::new().add_attribute("method", "whitelist"))
        }
        ExecuteMsg::Remove { address } => {
            let mut state: State = STATE.load(deps.storage)?;
            if state.allowed_submitters.contains(&address) {
                state.allowed_submitters.retain(|x| x != &address);
            }

            STATE.save(deps.storage, &state)?;            
            Ok(Response::new().add_attribute("method", "remove"))
        }

        // Submission
        ExecuteMsg::Submit { entries } => {
            let sender = info.sender.to_string();
            let state: State = STATE.load(deps.storage)?;

            // users can only submit if they are whitelisted
            if !state.allowed_submitters.contains(&sender) {
                return Err(ContractError::Unauthorized {});
            }

            let mut data_state: DataEntries = DATA_STATE.load(deps.storage)?;        
            if !data_state.entries.contains_key(&sender) {
                data_state.entries.insert(sender.clone(), BTreeMap::new());
            }

            // get latest id in the btree map. Change this in the future to be better
            let mut latest_id: u128 = 0;
            for (key, _) in data_state.entries.get(&sender).unwrap() {
                if *key > latest_id {
                    latest_id = *key;
                }
            }

            // add entries to the btree map
            for entry in entries {
                latest_id += 1;
                data_state.entries.get_mut(&sender).unwrap().insert(latest_id, entry);
            }
            
            DATA_STATE.save(deps.storage, &data_state)?;

            Ok(Response::new().add_attribute("method", "submit"))
        }
    }
}
#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {        
        QueryMsg::GetEntries { address } => {
            let data_state: DataEntries = DATA_STATE.load(deps.storage)?;
            let entries = data_state.entries.get(&address).unwrap();

            to_binary(&entries)
        }

        QueryMsg::GetSpecificEntry { address, id } => {
            let data_state: DataEntries = DATA_STATE.load(deps.storage)?;
            let entry = data_state.entries.get(&address).unwrap().get(&id).unwrap();

            to_binary(&entry)
        }

        QueryMsg::GetWhitelist {  } => {
            let state: State = STATE.load(deps.storage)?;
            let whitelist = state.allowed_submitters;

            to_binary(&whitelist)
        }
    }
}
