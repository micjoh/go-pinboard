package pinboard

type PinboardError string

func (e PinboardError) Error() string {
	return string(e)
}

type PinboardAPIError string

func (e PinboardAPIError) Error() string {
	return string(e)
}

var (
	ErrTooManyRequests = PinboardError("429: Too Many Requests")
	ErrForbidden       = PinboardError("403: Forbidden")
)
