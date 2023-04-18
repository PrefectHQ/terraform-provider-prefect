package prefect_api

type Workspace struct {
	Id                     string `json:"id,omitempty"`
	Created                string `json:"created,omitempty"`
	Updated                string `json:"updated,omitempty"`
	AccountId              string `json:"account_id,omitempty"`
	Name                   string `json:"name,omitempty"`
	Description            string `json:"description,omitempty"`
	Handle                 string `json:"handle,omitempty"`
	DefaultWorkspaceRoleId string `json:"default_workspace_role_id,omitempty"`
}

type WorkQueue struct {
	Id               string `json:"id,omitempty"`
	Created          string `json:"created,omitempty"`
	Updated          string `json:"updated,omitempty"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	IsPaused         bool   `json:"is_paused,omitempty"`
	ConcurrencyLimit int    `json:"concurrency_limit,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	WorkPoolId       string `json:"work_pool_id,omitempty"`
	Filter           string `json:"filter,omitempty"`
	LastPolled       string `json:"last_polled,omitempty"`
}

type BlockType struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type BlockSchema struct {
	Id          string `json:"id,omitempty"`
	BlockTypeId string `json:"block_type_id,omitempty"`
}

type BlockDocument struct {
	Id                      string      `json:"id,omitempty"`
	Created                 string      `json:"created,omitempty"`
	Updated                 string      `json:"updated,omitempty"`
	Name                    string      `json:"name,omitempty"`
	BlockSchemaId           string      `json:"block_schema_id,omitempty"`
	BlockTypeId             string      `json:"block_type_id,omitempty"`
	IsAnonymous             bool        `json:"is_anonymous,omitempty"`
	Data                    interface{} `json:"data,omitempty"`
	BlockDocumentReferences interface{} `json:"block_document_references,omitempty"`
}
