package domain

import "errors"

// Sentinel errors for the product domain.
var (
	ErrDiscountInvalidPercentage = errors.New("discount percentage must be between 0 and 100")
	ErrDiscountInvalidPeriod     = errors.New("discount end date must be after start date")

	// Product errors
	ErrProductNotActive         = errors.New("product is not active")
	ErrProductNotFound          = errors.New("product not found")
	ErrProductIDRequired        = errors.New("product id is required")
	ErrProductNameRequired      = errors.New("product name is required")
	ErrProductBasePriceRequired = errors.New("product base price is required")

	// Discount errors
	ErrInvalidDiscountPeriod = errors.New("invalid discount period")
	ErrNoActiveDiscount      = errors.New("product has no active discount")

	// General validation errors
	ErrInvalidStatus = errors.New("invalid product status")

	// Money errors
	ErrNegativeAmount        = errors.New("money amount cannot be negative")
	ErrCurrencyMismatch      = errors.New("currency mismatch")
	ErrInvalidCurrency       = errors.New("invalid currency code")
	ErrDivisionByZero        = errors.New("division by zero")
	ErrInvalidDiscountAmount = errors.New("discount amount must be between 0 and 100")
)
