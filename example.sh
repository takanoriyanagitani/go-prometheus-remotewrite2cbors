#!/bin/bash

bin="./cmd/promrw2cbors2dump/promrw2cbors2dump"

export ENV_MAX_POST_BODY_SIZE=16777216
export ENV_LISTEN_ADDR_PORT="127.0.0.1:8804"

"${bin}" |
	python3 -m cbor2.tool --sequence --pretty
