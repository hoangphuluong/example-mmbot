package model

type AssetType string

const (
	Coin AssetType = "coin"
	Fiat AssetType = "fiat"
)

type AssetSymbol struct {
	ID     string    `json:"id" gorm:"primary_key"`
	Symbol string    `json:"symbol"`
	Name   string    `json:"name"`
	Type   AssetType `json:"assetType"`
}

func (AssetSymbol) TableName() string {
	return "asset_symbol"
}
