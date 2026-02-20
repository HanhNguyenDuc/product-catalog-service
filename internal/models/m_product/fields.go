package m_product

// Column names cho bảng products trong Spanner.
const (
	ProductID            string = "product_id"
	Name                 string = "name"
	Description          string = "description"
	Category             string = "category"
	BasePriceNumerator   string = "base_price_numerator"
	BasePriceDenominator string = "base_price_denominator"
	DiscountPercent      string = "discount_percent"
	DiscountStartDate    string = "discount_start_date"
	DiscountEndDate      string = "discount_end_date"
	Status               string = "status"
	CreatedAt            string = "created_at"
	UpdatedAt            string = "updated_at"
	ArchivedAt           string = "archived_at"
)
