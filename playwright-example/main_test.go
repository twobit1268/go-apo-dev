package main

import (
	"os"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

var (
	pw      *playwright.Playwright
	browser playwright.Browser
)

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

// login navigates to /dashboard, signs in with the demo account, and
// returns a page that has landed on the authenticated dashboard.
func login(t *testing.T) playwright.Page {
	t.Helper()

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto("http://localhost/dashboard"); err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	if err := page.Locator("[data-testid=email-input]").Fill("demo@fintech.com"); err != nil {
		t.Fatalf("could not fill email: %v", err)
	}
	if err := page.Locator("[data-testid=password-input]").Fill("password123"); err != nil {
		t.Fatalf("could not fill password: %v", err)
	}
	if err := page.Locator("[data-testid=login-btn]").Click(); err != nil {
		t.Fatalf("could not click login: %v", err)
	}
	if err := page.WaitForURL("**/dashboard"); err != nil {
		t.Fatalf("did not land on dashboard: %v", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		t.Fatalf("dashboard data did not finish loading: %v", err)
	}

	return page
}

func TestLoginRedirectsToDashboard(t *testing.T) {
	page := login(t)
	defer page.Close()

	heading, err := page.Locator("[data-testid=dashboard-heading]").TextContent()
	if err != nil {
		t.Fatalf("could not get heading: %v", err)
	}

	want := "Welcome back, Demo User"
	if heading != want {
		t.Errorf("got heading %q, want %q", heading, want)
	}
}

func TestDashboardShowsBalanceCard(t *testing.T) {
	page := login(t)
	defer page.Close()

	amount, err := page.Locator("[data-testid=balance-amount]").TextContent()
	if err != nil {
		t.Fatalf("could not get balance amount: %v", err)
	}
	if amount == "" {
		t.Error("expected a non-empty balance amount")
	}

	account, err := page.Locator("[data-testid=account-number]").TextContent()
	if err != nil {
		t.Fatalf("could not get account number: %v", err)
	}
	if account == "" {
		t.Error("expected a non-empty account number")
	}
	t.Logf("balance: %s, account: %s", amount, account)
}

func TestDashboardNavigation(t *testing.T) {
	page := login(t)
	defer page.Close()

	links := map[string]string{
		"nav-dashboard":    "/dashboard",
		"nav-transactions": "/transactions",
		"nav-send":         "/send",
		"nav-budgets":      "/budgets",
	}

	for testID, wantHref := range links {
		href, err := page.Locator("[data-testid=" + testID + "]").GetAttribute("href")
		if err != nil {
			t.Errorf("%s: could not get href: %v", testID, err)
			continue
		}
		if href != wantHref {
			t.Errorf("%s: got href %q, want %q", testID, href, wantHref)
		}
	}
}

func TestDashboardRecentTransactions(t *testing.T) {
	page := login(t)
	defer page.Close()

	rows, err := page.Locator("[data-testid^=tx-row-]").Count()
	if err != nil {
		t.Fatalf("could not count transaction rows: %v", err)
	}
	if rows == 0 {
		t.Error("expected at least one recent transaction row")
	}
	t.Logf("found %d transaction rows", rows)
}
