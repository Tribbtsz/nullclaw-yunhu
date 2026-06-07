package channel

// ProtocolVersion is the external channel protocol version supported by this plugin.
const ProtocolVersion = 2

// Manifest describes the plugin's capabilities to nullclaw.
type Manifest struct {
	ProtocolVersion int             `json:"protocol_version"`
	Capabilities    ManifestCapabilities `json:"capabilities"`
}

type ManifestCapabilities struct {
	Health       bool `json:"health"`
	Streaming    bool `json:"streaming"`
	SendRich     bool `json:"send_rich"`
	Typing       bool `json:"typing"`
	Edit         bool `json:"edit"`
	Delete       bool `json:"delete"`
	Reactions    bool `json:"reactions"`
	ReadReceipts bool `json:"read_receipts"`
}

// BuildManifest returns a Manifest declaring the plugin's supported capabilities.
func BuildManifest() Manifest {
	return Manifest{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ManifestCapabilities{
			Health:       true,
			Streaming:    false,
			SendRich:     true,
			Typing:       false,
			Edit:         true,
			Delete:       true,
			Reactions:    false,
			ReadReceipts: false,
		},
	}
}

type StartResult struct {
	Started bool `json:"started"`
}

type StopResult struct {
	Accepted bool `json:"accepted"`
}

type SendResult struct {
	Accepted  bool   `json:"accepted"`
	MessageID string `json:"message_id,omitempty"`
}

type EditResult struct {
	Accepted bool `json:"accepted"`
}

type DeleteResult struct {
	Accepted bool `json:"accepted"`
}

type HealthResult struct {
	Healthy bool `json:"healthy"`
}
