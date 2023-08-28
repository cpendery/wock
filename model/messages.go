package model

type Message struct {
	MsgType  MessageType `json:"msgType"`
	ClientId string      `json:"clientId,omitempty"`
	Data     []byte      `json:"data,omitempty"`
}

type MessageType int32

const (
	TerminateMessage MessageType = 0
	StatusMessage    MessageType = 1
	MockMessage      MessageType = 2
	UnmockMessage    MessageType = 3
	ClearMessage     MessageType = 4
	SuccessMessage   MessageType = 5
	ErrorMessage     MessageType = 6
	TimeoutMessage   MessageType = 7
)

type MockedHost struct {
	Host      string
	Directory string
}

type MockMessageData struct {
	Host      string `json:"host"`
	Directory string `json:"dir"`
}
