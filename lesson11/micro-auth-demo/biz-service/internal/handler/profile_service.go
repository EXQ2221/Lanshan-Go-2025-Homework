package handler

import (
	"context"

	"example.com/micro-auth-demo/biz-service/internal/biz"
)

type ProfileServiceImpl struct {
	Service *biz.ProfileService
}

func NewProfileServiceImpl(service *biz.ProfileService) *ProfileServiceImpl {
	return &ProfileServiceImpl{Service: service}
}

func (h *ProfileServiceImpl) GetProfile(ctx context.Context, userID int64) (*biz.Profile, error) {
	return h.Service.GetProfile(ctx, userID)
}
