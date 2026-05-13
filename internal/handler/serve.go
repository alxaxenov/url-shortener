package handler

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/middleware"
	"github.com/go-chi/chi/v5"

	_ "github.com/alxaxenov/url-shortener/tree/v2/docs"
	"github.com/swaggo/http-swagger"
)

// IHandler интерфейс http обработчиков.
type IHandler interface {
	AddValue(w http.ResponseWriter, r *http.Request)
	GetValue(w http.ResponseWriter, r *http.Request)
	AddValueJSON(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	SaveBatch(w http.ResponseWriter, r *http.Request)
	UserURLs(w http.ResponseWriter, r *http.Request)
	DeleteURLs(w http.ResponseWriter, r *http.Request)
}

// IComplexMiddleware интерфейс сложной middleware, для запуска которой необходимо вызвать метод.
type IComplexMiddleware interface {
	Use(http.Handler) http.Handler
}

// Настройки времени таймаута хендлеров.
const (
	timeoutDefault = 3 * time.Second
	timeoutBatch   = 5 * time.Second
)

// Serve подключение хендлеров и запуск роутера.
func Serve(addr string, h IHandler, userMiddleware IComplexMiddleware, enableHTTPS bool) error {
	r := chi.NewRouter()

	r.Use(middleware.GzipMiddleware)
	r.Use(middleware.WithLogging)
	if userMiddleware != nil {
		r.Use(userMiddleware.Use)
	}

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Post("/", timeoutHandler(h.AddValue, timeoutDefault, ""))
	r.Get("/{id}", timeoutHandler(h.GetValue, timeoutDefault, ""))
	r.Get("/ping", timeoutHandler(h.Ping, timeoutDefault, ""))

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", timeoutHandler(h.AddValueJSON, timeoutDefault, ""))
		r.Post("/shorten/batch", timeoutHandler(h.SaveBatch, timeoutBatch, ""))

		r.Get("/user/urls", timeoutHandler(h.UserURLs, timeoutDefault, ""))
		r.Delete("/user/urls", timeoutHandler(h.DeleteURLs, timeoutDefault, ""))
	})

	logger.Logger.Info("Running server on", addr)
	if !enableHTTPS {
		return http.ListenAndServe(addr, r)
	}
	listener, err := getHTTPSListener(addr)
	if err != nil {
		return err
	}
	return http.Serve(listener, r)

}

func getHTTPSListener(addr string) (net.Listener, error) {
	cert := &x509.Certificate{
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("getHTTPSListener GenerateKey error %w", err)
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("getHTTPSListener CreateCertificate error %w", err)
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("getHTTPSListener pem.Encode cert error %w", err)
	}
	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, fmt.Errorf("getHTTPSListener pem.Encode private key error %w", err)
	}

	TLSCert := tls.Certificate{
		Certificate: [][]byte{certBytes},
		PrivateKey:  privateKey,
	}
	config := &tls.Config{Certificates: []tls.Certificate{TLSCert}}
	listener, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("getHTTPSListener tls.Listen error %w", err)
	}
	return listener, nil
}
