package feed

// RankingEngine contains feed scoring strategies.
// The formulas are intentionally isolated so they can later be replaced
// by recommendation models without changing feed handlers.
type RankingEngine struct{}

func NewRankingEngine() *RankingEngine {
	return &RankingEngine{}
}

// EngagementScore provides a shared scoring entry point.
// Future versions can incorporate watch time, completion rate,
// creator affinity, and ML-generated scores.
func (r *RankingEngine) EngagementScore(views int64, likes int64, comments int64) float64 {
	return float64(views)*0.2 + float64(likes)*0.5 + float64(comments)*0.3
}
