/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package logger

import (
	"io"
	"log/slog"
	"net"

	"github.com/gorilla/handlers"
)

func HttpLogFormater(writer io.Writer, params handlers.LogFormatterParams) {
	req := params.Request

	username := "-"
	if params.URL.User != nil {
		if name := params.URL.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = params.URL.RequestURI()
	}

	slog.Info("http request",
		"host", host,
		"username", username,
		"timestamp", params.TimeStamp.UTC().Format("02/Jan/2006:15:04:05 -0700"),
		"method", req.Method,
		"uri", uri,
		"proto", req.Proto,
		"status", params.StatusCode,
		"size", params.Size,
		"referer", req.Referer(),
		"user_agent", req.UserAgent(),
	)
}
