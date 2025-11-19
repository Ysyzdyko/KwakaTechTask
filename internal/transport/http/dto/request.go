package dto

type ParseRequest struct {
	SpreadsheetID  string `json:"spreadsheet_id" binding:"required"`
	RestaurantName string `json:"restaurant_name" binding:"required"`
}

type ProductStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=available not_available deleted"`
	Reason string `json:"reason"`
}
