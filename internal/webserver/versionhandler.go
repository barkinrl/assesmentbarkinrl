/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kazimsarikaya/assesmentbarkinrl/internal/config"
)

func getVersion(w http.ResponseWriter, r *http.Request, data map[string]interface{}) apiActionResult {
	return func() {

		json, err := json.Marshal(map[string]interface{}{"version": config.GetConfig().GetVersion(), "build_time": config.GetConfig().GetBuildTime(), "go_version": config.GetConfig().GetGoVersion()})

		if err != nil {
			sendError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(json)

		if err != nil {
			slog.Error("Error writing response", "error", err)

		}
	}
}
