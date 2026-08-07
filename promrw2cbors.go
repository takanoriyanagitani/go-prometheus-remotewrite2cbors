package promrw2cbors

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	p2 "github.com/prometheus/prometheus/prompb"
)

var (
	ErrNotPost error = errors.New("not a post request")
)

type PostBodyRaw []byte

type PostBodyDecoded struct{ Raw []byte }

type RawToDecoded func(context.Context, PostBodyRaw) (PostBodyDecoded, error)

type RwRequest struct{ p2.WriteRequest }

func (r *RwRequest) Unmarshal(decoded PostBodyDecoded) error {
	return r.WriteRequest.Unmarshal(decoded.Raw)
}

func (r *RwRequest) WriteReq() *p2.WriteRequest { return &r.WriteRequest }

func (r *RwRequest) Timeseries() []p2.TimeSeries {
	return r.WriteRequest.Timeseries
}

func (r *RwRequest) Metadata() []p2.MetricMetadata {
	return r.WriteRequest.Metadata
}

type WreqWriter func(context.Context, *RwRequest) error

func (r *RwRequest) ToWriter(ctx context.Context, wtr WreqWriter) error {
	return wtr(ctx, r)
}

type PostBodySizeMax int64

func (m PostBodySizeMax) ToLimited(rdr io.Reader) io.Reader {
	return &io.LimitedReader{R: rdr, N: int64(m)}
}

type HandlerBuilder struct {
	PostBodySizeMax
	RawToDecoded
	WreqWriter
}

type HTTPReq struct{ *http.Request }

func (q HTTPReq) PostBodyReader() (io.Reader, error) {
	if http.MethodPost != q.Request.Method {
		slog.Warn("invalid request got", "method", q.Request.Method)
		return nil, ErrNotPost
	}

	return q.Request.Body, nil
}

func (q HTTPReq) PostBodyLimited(pmax PostBodySizeMax) (io.Reader, error) {
	rdr, err := q.PostBodyReader()
	lmtd := pmax.ToLimited(rdr)
	return lmtd, err
}

func (q HTTPReq) LimitedToBuf(pmax PostBodySizeMax, buf *bytes.Buffer) error {
	buf.Reset()

	lmtd, lerr := q.PostBodyLimited(pmax)
	if nil != lerr {
		return lerr
	}

	_, berr := io.Copy(buf, lmtd)
	return berr
}

func (b HandlerBuilder) ToHandler() http.HandlerFunc {
	return func(rwtr http.ResponseWriter, req *http.Request) {
		var ctx context.Context = req.Context()

		var buf bytes.Buffer
		err := HTTPReq{req}.LimitedToBuf(b.PostBodySizeMax, &buf)
		if nil != err {
			http.Error(rwtr, "unable to get snappy bytes", http.StatusBadRequest)
			slog.Warn("invalid request", "body", "unable to get the body")
			return
		}

		decoded, err := b.RawToDecoded(ctx, PostBodyRaw(buf.Bytes()))
		if nil != err {
			http.Error(rwtr, "invalid snappy bytes", http.StatusBadRequest)
			slog.Warn(
				"invalid request",
				"body",
				"invalid snappy bytes",
				"detail",
				err,
			)
			return
		}

		var parsed RwRequest
		perr := (&parsed).Unmarshal(decoded)
		if nil != perr {
			http.Error(rwtr, "invalid remote write request", http.StatusBadRequest)
			slog.Warn("invalid request", "body", "invalid remote write bytes")
			return
		}

		werr := b.WreqWriter(ctx, &parsed)
		if nil != werr {
			http.Error(
				rwtr,
				"unable to process the request",
				http.StatusInternalServerError,
			)
			return
		}

		rwtr.WriteHeader(http.StatusNoContent)
	}
}
