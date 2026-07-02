package lsp

type ConnectEvent struct {
	Started bool
	Command string
	Ext     string
	Error   string
}

type EventHandler func(ConnectEvent)
