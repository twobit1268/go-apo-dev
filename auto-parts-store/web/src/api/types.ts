export interface Category {
  id: string;
  name: string;
  slug: string;
}

export interface Part {
  id: string;
  sku: string;
  name: string;
  description: string;
  categoryId: string;
  priceCents: number;
  stockQty: number;
}

export interface CartItem {
  partId: string;
  quantity: number;
}

export interface Cart {
  id: string;
  customerId: string;
  items: CartItem[];
  createdAt: string;
}

export interface OrderItem {
  partId: string;
  quantity: number;
  unitPriceCents: number;
}

export interface Order {
  id: string;
  customerId: string;
  status: string;
  totalCents: number;
  items: OrderItem[];
  createdAt: string;
}
