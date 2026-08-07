package wreq2cbor

import (
	"context"
	"io"

	fa "github.com/fxamacker/cbor/v2"
	p2 "github.com/prometheus/prometheus/prompb"
	rc "github.com/takanoriyanagitani/go-prometheus-remotewrite2cbors"
)

type Encoder struct{ *fa.Encoder }

func (e Encoder) WriteReq(_ context.Context, req *rc.RwRequest) error {
	var wreq *p2.WriteRequest = &req.WriteRequest
	return e.Encoder.Encode(wreq)
}

func (e Encoder) AsWreqWriter() rc.WreqWriter { return e.WriteReq }

func IoWtrToWreqWtr(wtr io.Writer) rc.WreqWriter {
	return Encoder{Encoder: fa.NewEncoder(wtr)}.AsWreqWriter()
}
