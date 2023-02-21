package auth

type GoogleIdentityPlatformErrorResponse struct {
	Error GoogleIdentityPlatformError `json:"error"`
}

type GoogleIdentityPlatformError struct {
	Code    uint                                `json:"code"`
	Errors  []GoogleIdentityPlatformErrorDetail `json:"errors,omitempty"`
	Message string                              `json:"message"`
	Status  string                              `json:"status,omitempty"`
}

type GoogleIdentityPlatformErrorDetail struct {
	Domain  string `json:"domain"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}
