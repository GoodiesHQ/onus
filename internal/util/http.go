package util

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/rs/zerolog/log"
)

var queryDecoder = func() *schema.Decoder {
	d := schema.NewDecoder()
	d.IgnoreUnknownKeys(false)
	d.SetAliasTag("schema")
	return d
}()

func DecodeQueryParams(r *http.Request, dst any) error {
	err := queryDecoder.Decode(dst, r.URL.Query())
	if err != nil {
		return fmt.Errorf("failed to decode query params: %w", err)
	}
	return nil
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return fmt.Errorf("extra data after JSON object")
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data any, pretty bool) {
	var encoded []byte
	var err error
	if pretty {
		encoded, err = json.MarshalIndent(data, "", "    ")
	} else {
		encoded, err = json.Marshal(data)
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal JSON response")
		http.Error(w, "Unprocessesable response", http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(encoded)
}
