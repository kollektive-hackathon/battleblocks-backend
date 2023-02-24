package utils

type PageResponse[T any] struct {
	Items         []T   `json:"items"`
	NextPageToken int64 `json:"nextPageToken,omitempty"`
	ItemCount     int64 `json:"itemCount"`
}

func NewPageResponse[T any]() *PageResponse[T] {
	return &PageResponse[T]{}
}

func (pr *PageResponse[T]) WithItems(items []T) *PageResponse[T] {
	pr.Items = items
	return pr
}

func (pr *PageResponse[T]) WithNextPageToken(pageToken int64) *PageResponse[T] {
	pr.NextPageToken = pageToken
	return pr
}

func (pr *PageResponse[T]) WithItemCount(itemCount int64) *PageResponse[T] {
	pr.ItemCount = itemCount
	return pr
}

func (pr *PageResponse[T]) Build() *PageResponse[T] {
	return pr
}
