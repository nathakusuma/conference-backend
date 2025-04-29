package dto

type LazyLoadResponse struct {
	HasMore bool        `json:"has_more"`
	FirstID interface{} `json:"first_id"`
	LastID  interface{} `json:"last_id"`
}
