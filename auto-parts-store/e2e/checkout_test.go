//go:build e2e

package e2e

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

func TestAddToCartAndCheckout(t *testing.T) {
	page := newPage(t)

	if _, err := page.Goto(baseURL()); err != nil {
		t.Fatalf("could not goto catalog: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=part-list]"); err != nil {
		t.Fatalf("part list never rendered: %v", err)
	}

	addButtons, err := page.Locator("[data-testid^=add-to-cart-]").All()
	if err != nil || len(addButtons) == 0 {
		t.Fatalf("could not find add-to-cart buttons: %v", err)
	}
	if err := addButtons[0].Click(); err != nil {
		t.Fatalf("could not add first part to cart: %v", err)
	}

	if err := page.Locator("[data-testid=nav-cart]").Click(); err != nil {
		t.Fatalf("could not navigate to cart: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=cart-items]"); err != nil {
		t.Fatalf("cart items never rendered: %v", err)
	}

	if err := page.Locator("[data-testid=checkout-btn]").Click(); err != nil {
		t.Fatalf("could not click checkout: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=place-order-btn]"); err != nil {
		t.Fatalf("checkout page never rendered: %v", err)
	}
	if err := page.Locator("[data-testid=place-order-btn]").Click(); err != nil {
		t.Fatalf("could not place order: %v", err)
	}

	if err := page.WaitForURL("**/orders/*", playwright.PageWaitForURLOptions{}); err != nil {
		t.Fatalf("did not land on order confirmation: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=order-confirmation]"); err != nil {
		t.Fatalf("order confirmation never rendered: %v", err)
	}

	orderID, err := page.Locator("[data-testid=order-id]").TextContent()
	if err != nil {
		t.Fatalf("could not read order id: %v", err)
	}
	if !strings.HasPrefix(orderID, "Order #") {
		t.Errorf("unexpected order id text: %q", orderID)
	}

	status, err := page.Locator("[data-testid=order-status]").TextContent()
	if err != nil {
		t.Fatalf("could not read order status: %v", err)
	}
	if !strings.Contains(status, "placed") {
		t.Errorf("expected order status to be 'placed', got %q", status)
	}
}

func TestRemoveCartItemUpdatesTotal(t *testing.T) {
	page := newPage(t)

	if _, err := page.Goto(baseURL()); err != nil {
		t.Fatalf("could not goto catalog: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=part-list]"); err != nil {
		t.Fatalf("part list never rendered: %v", err)
	}

	addButtons, err := page.Locator("[data-testid^=add-to-cart-]").All()
	if err != nil || len(addButtons) < 2 {
		t.Fatalf("expected at least two parts to add to cart: %v", err)
	}
	if err := addButtons[0].Click(); err != nil {
		t.Fatalf("could not add first part: %v", err)
	}
	if err := addButtons[1].Click(); err != nil {
		t.Fatalf("could not add second part: %v", err)
	}

	if err := page.Locator("[data-testid=nav-cart]").Click(); err != nil {
		t.Fatalf("could not navigate to cart: %v", err)
	}
	if _, err := page.WaitForSelector("[data-testid=cart-items]"); err != nil {
		t.Fatalf("cart items never rendered: %v", err)
	}

	totalBefore, err := page.Locator("[data-testid=cart-total]").TextContent()
	if err != nil {
		t.Fatalf("could not read cart total: %v", err)
	}

	removeButtons, err := page.Locator("[data-testid^=remove-item-]").All()
	if err != nil || len(removeButtons) == 0 {
		t.Fatalf("could not find remove buttons: %v", err)
	}
	if err := removeButtons[0].Click(); err != nil {
		t.Fatalf("could not remove a cart item: %v", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		t.Fatalf("cart did not finish updating: %v", err)
	}

	totalAfter, err := page.Locator("[data-testid=cart-total]").TextContent()
	if err != nil {
		t.Fatalf("could not read updated cart total: %v", err)
	}
	if totalAfter == totalBefore {
		t.Errorf("expected cart total to change after removing an item, stayed at %q", totalBefore)
	}
}
