package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/dock-tech/isis-golang-lib/domainerrors"
	log "github.com/dock-tech/isis-golang-lib/observability/logger/zap"
	"github.com/dock-tech/isis-golang-lib/strutil"
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

func ParseRequest(ctx *fiber.Ctx, receiver interface{}) (map[string]interface{}, error) {
	fields, err := readRequestToMap(ctx.UserContext(), ctx.Request())
	if err != nil || fields == nil {
		return fields, &domainerrors.UnsupportedMediaTypeError{}
	}

	headers := ctx.GetReqHeaders()
	for kHeaders := range headers {
		if fHeaders := headers[kHeaders]; len(fHeaders) > 0 {
			ksHeaders := strutil.ToSnake(kHeaders)
			fields[ksHeaders] = strings.Join(headers[kHeaders], "")
		}
	}

	b, err := jsoniter.Marshal(fields)
	if err != nil {
		return fields, &domainerrors.UnsupportedMediaTypeError{}
	}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	err = d.Decode(receiver)
	if err != nil {
		return fields, err
	}

	return fields, nil
}

func readRequestToMap(ctx context.Context, req *fasthttp.Request) (map[string]interface{}, error) {
	var resp map[string]interface{}

	dcd := json.NewDecoder(bytes.NewBuffer(req.Body()))
	defer func(ctx context.Context) {
		err := req.CloseBodyStream()
		if err != nil {
			log.Errorf(ctx, "Error executing function CloseBodyStream - "+err.Error())
		}
	}(ctx)
	dcd.UseNumber()
	err := dcd.Decode(&resp)
	if err != nil {
		return nil, err
	}

	parseAdditionalDataFields(resp)

	return resp, nil
}

func parseAdditionalDataFields(dataMap map[string]interface{}) {
	if additionalData, ok := dataMap["additional_data"].(map[string]interface{}); ok {
		delete(dataMap, "additional_data")
		for k, v := range additionalData {
			dataMap[k] = v
		}
	}
}
