package repository

type ShortenerRepo interface {
	SetValue(string, string) error
	GetValue(string) (string, error)
}
