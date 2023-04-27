package test

type JournalEntry struct {
	Date   string `json:"date"`
	Title  string `json:"title"`
	RepoPr string `json:"repo_pr"`
	Notes  string `json:"notes"`
}

type JournalEntries map[string]JournalEntry

type JournalEntriesResponse struct {
	Data *JournalEntries `json:"data"`
}

type WhitelistResponse struct {
	Data []string `json:"data"`
}

type GetEntries struct {
	Address string `json:"address"`
}

type QueryMsg struct {
	GetEntries   *GetEntries `json:"get_entries,omitempty"`
	GetWhitelist *struct{}   `json:"get_whitelist,omitempty"`
}
