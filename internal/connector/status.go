package connector

// ConnectionStatus is the high-level WebSocket connection state.
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusReconnecting ConnectionStatus = "reconnecting"
	StatusFailed       ConnectionStatus = "failed"
)

// MapConnectionStatus maps a pusher-go connection state string to ConnectionStatus.
func MapConnectionStatus(state string) ConnectionStatus {
	switch state {
	case "connected":
		return StatusConnected
	case "disconnected":
		return StatusDisconnected
	case "connecting":
		return StatusConnecting
	case "unavailable":
		return StatusReconnecting
	case "failed":
		return StatusFailed
	default:
		return StatusDisconnected
	}
}
