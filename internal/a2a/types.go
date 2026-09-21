package a2a

import "encoding/json"

// AgentCard is the A2A agent discovery document (protocol 0.3.0).
type AgentCard struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	URL                string   `json:"url"`
	Version            string   `json:"version"`
	ProtocolVersion    string   `json:"protocolVersion"`
	PreferredTransport string   `json:"preferredTransport"`
	DefaultInputModes  []string `json:"defaultInputModes"`
	DefaultOutputModes []string `json:"defaultOutputModes"`
}

// jsonRPCRequest is a JSON-RPC 2.0 request envelope.
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response envelope.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type messageSendParams struct {
	Message       a2aMessage                `json:"message"`
	Configuration *messageSendConfiguration `json:"configuration,omitempty"`
}

type messageSendConfiguration struct {
	Blocking bool `json:"blocking,omitempty"`
}

type a2aMessage struct {
	Kind      string    `json:"kind"`
	MessageID string    `json:"messageId"`
	Role      string    `json:"role"`
	Parts     []a2aPart `json:"parts"`
	ContextID string    `json:"contextId,omitempty"`
	TaskID    string    `json:"taskId,omitempty"`
}

type a2aPart struct {
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
}

// taskOrMessage covers message/send results (Task or Message).
type taskOrMessage struct {
	Kind      string          `json:"kind"`
	ID        string          `json:"id"`
	Status    *taskStatus     `json:"status,omitempty"`
	Artifacts []artifact      `json:"artifacts,omitempty"`
	History   []a2aMessage    `json:"history,omitempty"`
	MessageID string          `json:"messageId,omitempty"`
	Role      string          `json:"role,omitempty"`
	Parts     []a2aPart       `json:"parts,omitempty"`
	Raw       json.RawMessage `json:"-"`
}

type taskStatus struct {
	State     string      `json:"state"`
	Timestamp string      `json:"timestamp,omitempty"`
	Message   *a2aMessage `json:"message,omitempty"`
}

type artifact struct {
	ArtifactID string    `json:"artifactId"`
	Name       string    `json:"name,omitempty"`
	Parts      []a2aPart `json:"parts"`
}
