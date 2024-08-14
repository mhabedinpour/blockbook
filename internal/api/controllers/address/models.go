package address

type AddressModel struct {
	Address string `json:"address" uri:"address" binding:"required,eth_addr"`
}
