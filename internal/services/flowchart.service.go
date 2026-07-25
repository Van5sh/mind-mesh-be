package services

import "example/hello/internal/repository"

type FlowchartService struct {
	repo *repository.FlowchartRepository
}

func NewFlowchartService(repo *repository.FlowchartRepository) *FlowchartService {
	return &FlowchartService{repo: repo}
}
