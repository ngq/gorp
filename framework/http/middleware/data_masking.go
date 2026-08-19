// Package middleware provides automatic sensitive data masking middleware for gorp framework.
package middleware

import (
	"bytes"
	"regexp"

	"github.com/gin-gonic/gin"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

var (
	mobileRegex   = regexp.MustCompile(`(1[3-9]\d)\d{4}(\d{4})`)
	idCardRegex   = regexp.MustCompile(`([1-9]\d{5})\d{8}(\d{3}[\dX])`)
	bankCardRegex = regexp.MustCompile(`(62\d{4})\d{6,9}(\d{4})`)
	emailRegex    = regexp.MustCompile(`([a-zA-Z0-9_\-\.]{1,2})[a-zA-Z0-9_\-\.]+(@[a-zA-Z0-9_\-\.]+\.[a-zA-Z]{2,})`)
)

// DataMaskingOptions configures data masking rules.
type DataMaskingOptions struct {
	MaskMobile   bool
	MaskIDCard   bool
	MaskBankCard bool
	MaskEmail    bool
}

// DefaultDataMaskingOptions returns default production settings.
func DefaultDataMaskingOptions() DataMaskingOptions {
	return DataMaskingOptions{
		MaskMobile:   true,
		MaskIDCard:   true,
		MaskBankCard: true,
		MaskEmail:    true,
	}
}

// DataMaskingOption configures DataMaskingOptions.
type DataMaskingOption func(*DataMaskingOptions)

// DataMaskingMiddleware intercepts HTTP JSON responses and automatically masks sensitive fields (mobile, id card, bank card, email).
func DataMaskingMiddleware(opts ...DataMaskingOption) transportcontract.Middleware {
	cfg := DefaultDataMaskingOptions()
	for _, o := range opts {
		o(&cfg)
	}

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if gc, ok := unwrapGinContext(c); ok && gc.Writer != nil {
				w := &maskingResponseWriter{
					ResponseWriter: gc.Writer,
					cfg:            cfg,
				}
				gc.Writer = w
			}
			next(c)
		}
	}
}

type maskingResponseWriter struct {
	gin.ResponseWriter
	cfg DataMaskingOptions
}

func (w *maskingResponseWriter) Write(b []byte) (int, error) {
	masked := MaskSensitiveData(b, w.cfg)
	return w.ResponseWriter.Write(masked)
}

func (w *maskingResponseWriter) WriteString(s string) (int, error) {
	masked := MaskSensitiveData([]byte(s), w.cfg)
	return w.ResponseWriter.WriteString(string(masked))
}

// MaskSensitiveData performs regex-based masking on sensitive strings (mobile, id card, bank card, email).
func MaskSensitiveData(b []byte, cfg DataMaskingOptions) []byte {
	if len(b) == 0 {
		return b
	}

	res := bytes.Clone(b)
	// Apply Bank Card (16-19 digits) FIRST to avoid 19-digit ICBC 622202 cards matching 18-digit ID card regex
	if cfg.MaskBankCard {
		res = bankCardRegex.ReplaceAll(res, []byte("${1}******${2}"))
	}
	if cfg.MaskIDCard {
		res = idCardRegex.ReplaceAll(res, []byte("${1}********${2}"))
	}
	if cfg.MaskMobile {
		res = mobileRegex.ReplaceAll(res, []byte("${1}****${2}"))
	}
	if cfg.MaskEmail {
		res = emailRegex.ReplaceAll(res, []byte("${1}***${2}"))
	}
	return res
}
