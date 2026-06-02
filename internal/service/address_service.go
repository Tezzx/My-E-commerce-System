package service

import (
	"order-payment-system/internal/model"
	"order-payment-system/internal/repository"
	"order-payment-system/internal/types"
)

type AddressService struct {
	repo *repository.AddressRepo
}

func NewAddressService(repo *repository.AddressRepo) *AddressService {
	return &AddressService{repo: repo}
}

func (s *AddressService) Create(userID uint, req *types.AddressReq) error {
	if req.IsDefault {
		_ = s.repo.ClearDefault(userID)
	}
	return s.repo.Create(&model.Address{
		UserID:    userID,
		Receiver:  req.Receiver,
		Phone:     req.Phone,
		Province:  req.Province,
		City:      req.City,
		District:  req.District,
		Detail:    req.Detail,
		IsDefault: req.IsDefault,
	})
}

func (s *AddressService) Update(userID, addrID uint, req *types.AddressReq) error {
	addr, err := s.repo.GetByID(userID, addrID)
	if err != nil {
		return err
	}
	if req.IsDefault {
		_ = s.repo.ClearDefault(userID)
	}
	addr.Receiver = req.Receiver
	addr.Phone = req.Phone
	addr.Province = req.Province
	addr.City = req.City
	addr.District = req.District
	addr.Detail = req.Detail
	addr.IsDefault = req.IsDefault
	return s.repo.Update(addr)
}

func (s *AddressService) Delete(userID, addrID uint) error {
	return s.repo.Delete(userID, addrID)
}

func (s *AddressService) List(userID uint) ([]types.AddressResp, error) {
	addrs, err := s.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	var resp []types.AddressResp
	for _, a := range addrs {
		resp = append(resp, types.AddressResp{
			ID:        a.ID,
			Receiver:  a.Receiver,
			Phone:     a.Phone,
			Province:  a.Province,
			City:      a.City,
			District:  a.District,
			Detail:    a.Detail,
			IsDefault: a.IsDefault,
		})
	}
	return resp, nil
}

func (s *AddressService) GetByID(userID, addrID uint) (*model.Address, error) {
	return s.repo.GetByID(userID, addrID)
}
