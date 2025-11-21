package analytics

func ToDTO(stats []ReviewerStat) StatsResponseDTO {
	items := make([]StatItemDTO, len(stats))

	for i, s := range stats {
		items[i] = StatItemDTO{
			UserID:      s.UserID,
			ReviewCount: s.Count,
		}
	}

	return StatsResponseDTO{
		Stats: items,
	}
}
