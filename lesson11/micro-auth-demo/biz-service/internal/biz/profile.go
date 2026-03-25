package biz

import "context"

type Profile struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

type ProfileService struct{}

func NewProfileService() *ProfileService {
	return &ProfileService{}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*Profile, error) {
	_ = ctx

	return &Profile{
		UserID:   userID,
		Nickname: "demo-user",
		Bio:      "profile service scaffold",
	}, nil
}
