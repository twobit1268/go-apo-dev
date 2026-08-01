//go:build e2e

// Package e2e drives the running React SPA (built + served by
// docker-compose --profile full, per `make test-e2e`) with a real Chromium
// via playwright-go, mirroring the pattern already established in this
// repo's playwright-example/.
package e2e

import (
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

var (
	pw      *playwright.Playwright
	browser playwright.Browser
)

func baseURL() string {
	if url := os.Getenv("WEB_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:5173"
}

func TestMain(m *testing.M) {
	if err := playwright.Install(); err != nil {
		panic(err)
	}

	var err error
	pw, err = playwright.Run()
	if err != nil {
		panic(err)
	}
	browser, err = pw.Chromium.Launch()
	if err != nil {
		panic(err)
	}

	code := m.Run()

	browser.Close()
	pw.Stop()
	os.Exit(code)
}

func newPage(t *testing.T) playwright.Page {
	t.Helper()
	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}
	t.Cleanup(func() { page.Close() })
	return page
}
