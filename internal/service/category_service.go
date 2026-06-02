package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/types"
)

type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(req *types.CategoryReq) error {
	return s.repo.Create(&model.Category{
		Name:     req.Name,
		ParentID: req.ParentID,
		Sort:     req.Sort,
	})
}

func (s *CategoryService) Update(id uint, req *types.CategoryReq) error {
	cat, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	cat.Name = req.Name
	cat.ParentID = req.ParentID
	cat.Sort = req.Sort
	return s.repo.Update(cat)
}

func (s *CategoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *CategoryService) GetTree() ([]types.CategoryResp, error) {
	cats, err := s.repo.GetTree()
	if err != nil {
		return nil, err
	}
	return toTreeResp(cats), nil
}

func toTreeResp(cats []model.Category) []types.CategoryResp {
	var resp []types.CategoryResp
	for _, c := range cats {
		item := types.CategoryResp{
			ID:       c.ID,
			Name:     c.Name,
			ParentID: c.ParentID,
			Sort:     c.Sort,
			Children: toTreeResp(c.Children),
		}
		resp = append(resp, item)
	}
	return resp
}
