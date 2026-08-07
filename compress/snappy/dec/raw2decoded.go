package raw2decoded

import (
	"context"
	"fmt"

	"github.com/golang/snappy"
	rc "github.com/takanoriyanagitani/go-prometheus-remotewrite2cbors"
)

func Dec(original []byte) ([]byte, error) {
	decoded, err := snappy.Decode(nil, original)
	if nil != err {
		return nil, fmt.Errorf(
			"%w: rejected bytes(up to 16 bytes): %v",
			err,
			original[:min(16, len(original))],
		)
	}
	return decoded, nil
}

func RawToDec(_ context.Context, dat rc.PostBodyRaw) (rc.PostBodyDecoded, error) {
	decoded, err := Dec(dat)
	return rc.PostBodyDecoded{Raw: decoded}, err
}

var RawToDecoded rc.RawToDecoded = RawToDec
