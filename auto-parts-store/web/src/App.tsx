import { Route, Routes } from "react-router-dom";
import { Header } from "./components/Header";
import { CatalogPage } from "./pages/CatalogPage";
import { PartDetailPage } from "./pages/PartDetailPage";
import { CartPage } from "./pages/CartPage";
import { CheckoutPage } from "./pages/CheckoutPage";
import { OrderConfirmationPage } from "./pages/OrderConfirmationPage";

export function App() {
  return (
    <>
      <Header />
      <main className="app-main">
        <Routes>
          <Route path="/" element={<CatalogPage />} />
          <Route path="/parts/:id" element={<PartDetailPage />} />
          <Route path="/cart" element={<CartPage />} />
          <Route path="/checkout" element={<CheckoutPage />} />
          <Route path="/orders/:id" element={<OrderConfirmationPage />} />
        </Routes>
      </main>
    </>
  );
}
