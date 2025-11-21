package analytics

type StatItemDTO struct {
	UserID      string `json:"user_id"`
	ReviewCount int    `json:"review_count"`
}

type StatsResponseDTO struct {
	Stats []StatItemDTO `json:"stats"`
}
