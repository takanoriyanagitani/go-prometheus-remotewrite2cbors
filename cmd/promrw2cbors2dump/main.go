package main

import (
	"bufio"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	rc "github.com/takanoriyanagitani/go-prometheus-remotewrite2cbors"
	sd "github.com/takanoriyanagitani/go-prometheus-remotewrite2cbors/compress/snappy/dec"
	wf "github.com/takanoriyanagitani/go-prometheus-remotewrite2cbors/ser/cbor/wreq2cbor/fa"
)

var smaxPostBodySize string = os.Getenv("ENV_MAX_POST_BODY_SIZE")

var sListen string = os.Getenv("ENV_LISTEN_ADDR_PORT")

func sub() error {
	slog.Info("setting up", "ENV_MAX_POST_BODY_SIZE", smaxPostBodySize)

	imaxPostBodySize, err := strconv.Atoi(smaxPostBodySize)
	if nil != err {
		return err
	}

	var bwtr = bufio.NewWriter(os.Stdout)

	hbldr := rc.HandlerBuilder{
		PostBodySizeMax: rc.PostBodySizeMax(int64(imaxPostBodySize)),
		RawToDecoded:    sd.RawToDecoded,
		WreqWriter:      wf.IoWtrToWreqWtr(bwtr),
	}

	var hndl http.HandlerFunc = hbldr.ToHandler()

	http.HandleFunc("/receive", hndl)

	slog.Info("setup done", "post body size max", imaxPostBodySize)
	slog.Info("setup done", "ENV_LISTEN_ADDR_PORT", sListen)

	slog.Info("listening...")

	lerr := http.ListenAndServe(sListen, nil)
	if nil != lerr {
		return lerr
	}

	return bwtr.Flush()
}

func main() {
	err := sub()
	if nil != err {
		slog.Error("error got", "err", err)
	}
}
