package service

import "testing"

func TestCategorizeChannelTest(t *testing.T) {
	cases := map[int]string{
		200: "ok",
		204: "ok",
		400: "bad_request",
		401: "auth",
		403: "auth",
		404: "not_found",
		429: "rate_limit",
		500: "server_error",
		502: "server_error",
		503: "server_error",
		522: "upstream_blocked",
		524: "upstream_blocked",
		418: "unknown",
	}
	for status, want := range cases {
		if got := categorizeChannelTest(status); got != want {
			t.Fatalf("categorizeChannelTest(%d) = %s, want %s", status, got, want)
		}
	}
}
