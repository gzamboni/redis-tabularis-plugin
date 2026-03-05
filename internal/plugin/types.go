package plugin

import "encoding/json"

type ConnectionParams struct {
	Driver   string  `json:"driver"`
	Host     *string `json:"host"`
	Port     *int    `json:"port"`
	Database string  `json:"database"`
	Username *string `json:"username"`
	Password *string `json:"password"`
	SSLMode  *string `json:"ssl_mode"`
}

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type QueryCondition struct {
	Column   string
	Operator string
	Value    string
} // QueryCondition represents a WHERE clause condition
// =, !=, >, <, >=, <=, LIKE

type OrderBy struct {
	Column    string
	Direction string
} // OrderBy represents an ORDER BY clause
// ASC or DESC

type QueryParser struct {
	Table      string
	Conditions []QueryCondition // QueryParser holds parsed query information

	OrderBy []OrderBy
	Limit   int
	Offset  int
}

type InsertParser struct {
	Table   string
	Columns []string // InsertParser holds parsed INSERT statement information

	Values [][]string
}

type UpdateParser struct {
	Table      string
	SetClauses map[ // UpdateParser holds parsed UPDATE statement information
	string]string
	Conditions []QueryCondition
}

type DeleteParser struct {
	Table      string
	Conditions []QueryCondition // DeleteParser holds parsed DELETE statement information

}
