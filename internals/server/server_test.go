package server

import "testing"

func TestIsRedirectStatus(t *testing.T) {
	for _, status := range []int{
		301, 302, 303, 307, 308,
	} {
		if !isRedirectStatus(status) {
			t.Errorf("isRedirectStatus(%d) = false, want true", status)
		}
	}

	if isRedirectStatus(200) {
		t.Error("isRedirectStatus(200) = true, want false")
	}
}

func TestHandleRedirectionRewritesAbsoluteLocationAndPreservesQuery(t *testing.T) {
	headers := map[string][]string{
		"Location": {"http://localhost:8080/login?next=%2Faccount"},
	}

	handleRedirection(&headers, "demo", "example.test")

	if got, want := headers["Location"][0], "http://demo.example.test/login?next=%2Faccount"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
}

func TestHandleRedirectionLeavesRelativeLocationUntouched(t *testing.T) {
	headers := map[string][]string{"Location": {"/login?next=%2Faccount"}}

	handleRedirection(&headers, "demo", "example.test")

	if got, want := headers["Location"][0], "/login?next=%2Faccount"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
}

func TestHandleRedirectionAllowsMissingLocation(t *testing.T) {
	headers := map[string][]string{}

	handleRedirection(&headers, "demo", "example.test")
}
