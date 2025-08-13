package templates

// PlayerViewModel is a presentation-friendly shape for player rows.
type PlayerViewModel struct {
    UUID        string            `json:"uuid"`
    Name        string            `json:"name"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    Status      string            `json:"status"`
}

// ServerViewModel is a presentation-friendly shape for server rows.
type ServerViewModel struct {
    Name        string            `json:"name"`
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations"`
    Status      string            `json:"status"`
    PlayerCount int               `json:"player_count"`
}
