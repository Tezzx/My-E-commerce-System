package types

type AddressReq struct {
	Receiver  string `json:"receiver" binding:"required,max=50"`
	Phone     string `json:"phone" binding:"required,max=20"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail" binding:"required,max=200"`
	IsDefault bool   `json:"is_default"`
}

type AddressResp struct {
	ID        uint   `json:"id"`
	Receiver  string `json:"receiver"`
	Phone     string `json:"phone"`
	Province  string `json:"province"`
	City      string `json:"city"`
	District  string `json:"district"`
	Detail    string `json:"detail"`
	IsDefault bool   `json:"is_default"`
}
