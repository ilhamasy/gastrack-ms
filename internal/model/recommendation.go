package model

type Recommendation struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ActionLabel string `json:"action_label,omitempty"`
	ActionURL   string `json:"action_url,omitempty"`
	Priority    int    `json:"priority"`
}
