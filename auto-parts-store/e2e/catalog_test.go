//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

func TestBrowseCatalogAndViewPartDetail(t *testing.T) {
	page := newPage(t)

	if _, err := page.Goto(baseURL()); err != nil {
		t.Fatalf("could not goto catalog: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=part-list]"); err != nil {
		t.Fatalf("part list never rendered: %v", err)
	}

	cards, err := page.Locator("[data-testid^=part-card-]").Count()
	if err != nil {
		t.Fatalf("could not count part cards: %v", err)
	}
	if cards == 0 {
		t.Fatal("expected at least one part card from seed data")
	}

	links, err := page.Locator("[data-testid^=part-link-]").All()
	if err != nil || len(links) == 0 {
		t.Fatalf("could not find part links: %v", err)
	}
	if err := links[0].Click(); err != nil {
		t.Fatalf("could not click first part link: %v", err)
	}

	if _, err := page.WaitForSelector("[data-testid=part-detail]"); err != nil {
		t.Fatalf("part detail page never rendered: %v", err)
	}
	name, err := page.Locator("[data-testid=part-detail-name]").TextContent()
	if err != nil {
		t.Fatalf("could not read part detail name: %v", err)
	}
	if name == "" {
		t.Error("expected a non-empty part name on the detail page")
	}
}

func TestSearchFiltersCatalogResults(t *testing.T) {
	page := newPage(t)

	if _, err := page.Goto(baseURL()); err != nil {
		t.Fatalf("could not goto catalog: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=part-list]"); err != nil {
		t.Fatalf("part list never rendered: %v", err)
	}

	if err := page.Locator("[data-testid=search-input]").Fill("oil"); err != nil {
		t.Fatalf("could not fill search box: %v", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		t.Fatalf("search results did not finish loading: %v", err)
	}

	cardNames, err := page.Locator("[data-testid^=part-card-] h3").AllTextContents()
	if err != nil {
		t.Fatalf("could not read filtered part names: %v", err)
	}
	if len(cardNames) == 0 {
		t.Fatal("expected at least one result for 'oil'")
	}
	for _, name := range cardNames {
		if !strings.Contains(strings.ToLower(name), "oil") {
			t.Errorf("result %q does not match search term 'oil'", name)
		}
	}

	if err := page.Locator("[data-testid=search-input]").Fill("no-such-part-xyz"); err != nil {
		t.Fatalf("could not fill search box: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=no-results]"); err != nil {
		t.Fatalf("expected a no-results message for an unmatched search: %v", err)
	}
}

func TestFilterByCategory(t *testing.T) {
	page := newPage(t)

	if _, err := page.Goto(baseURL()); err != nil {
		t.Fatalf("could not goto catalog: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=part-list]"); err != nil {
		t.Fatalf("part list never rendered: %v", err)
	}

	if _, err := page.Locator("[data-testid=category-filter]").SelectOption(playwright.SelectOptionValues{
		Values: playwright.StringSlice("brakes"),
	}); err != nil {
		t.Fatalf("could not select category: %v", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		t.Fatalf("filtered results did not finish loading: %v", err)
	}

	cardNames, err := page.Locator("[data-testid^=part-card-] h3").AllTextContents()
	if err != nil {
		t.Fatalf("could not read filtered part names: %v", err)
	}
	if len(cardNames) == 0 {
		t.Fatal("expected at least one brakes part")
	}
	for _, name := range cardNames {
		if !strings.Contains(strings.ToLower(name), "brake") && !strings.Contains(strings.ToLower(name), "rotor") {
			t.Errorf("result %q does not look like a brakes part", name)
		}
	}
}
