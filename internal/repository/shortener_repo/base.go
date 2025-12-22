package shortener_repo

type ShortenerRepo interface {
	SetValue(string, string) error
	GetValue(string) (string, error)
}
