/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"bytes"
	"encoding/json"
)

// nolint:unused // this function is intentionally unused for now, after using please remove this comment
func transcode(in, out interface{}) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(in)

	if err != nil {
		return err
	}

	return json.NewDecoder(buf).Decode(out)
}
