package feed

import "gorm.io/gorm"

// Service owns feed ranking orchestration. Ranking strategies can evolve
// independently from HTTP handlers and can later be replaced by ML ranking.
type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}
