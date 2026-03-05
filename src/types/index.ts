// User types
export interface User {
  id: number;
  email: string;
  name: string;
  phone?: string;
  address?: string;
  role: 'admin' | 'user';
  user_status: 'cold' | 'warm';
  loyalty_tier?: string;
  loyalty_points?: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

// Product types
export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  is_active: boolean;
}

export interface ProductImage {
  id: number;
  url: string;
  alt_text?: string;
  sort_order: number;
  is_primary: boolean;
}

export interface Product {
  id: number;
  name: string;
  slug: string;
  brand: string;
  price: number;
  original_price?: number;
  discount?: number;
  category_id: number;
  category?: Category;
  description: string;
  image_url: string;
  images?: ProductImage[];
  stock: number;
  in_stock: boolean;
  rating: number;
  review_count: number;
  ai_score?: number;
  ai_recommend_type?: string;
  tags?: string[];
  ingredients?: string[];
}

// Cart types
export interface CartItem {
  id: number;
  product_id: number;
  product: Product;
  quantity: number;
  price: number;
  total: number;
}

export interface Cart {
  items: CartItem[] | null;
  subtotal: number;
  shipping_fee: number;
  total: number;
  item_count: number;
}

// Order types
export interface OrderItem {
  id: number;
  product_id: number;
  product: Product;
  quantity: number;
  unit_price: number;
  total_price: number;
}

export interface Order {
  id: number;
  order_code: string;
  subtotal: number;
  shipping_fee: number;
  discount: number;
  total: number;
  status: string;
  payment_method: 'cod' | 'zalopay' | 'momo' | 'vnpay';
  payment_status: 'pending' | 'success' | 'failed' | 'cancelled' | 'refunded';
  shipping_address: string;
  note: string;
  items: OrderItem[];
  created_at: string;
}

// API Response types
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

// Admin types
export interface AdminUser {
  id: number;
  email: string;
  name: string;
  phone?: string;
  address?: string;
  role: 'admin' | 'user';
  user_status: 'cold' | 'warm';
  order_count: number;
  created_at: string;
}

export interface AdminOrder {
  id: number;
  order_code: string;
  user: User;
  subtotal: number;
  shipping_fee: number;
  discount: number;
  total: number;
  status: string;
  shipping_address: string;
  note: string;
  items: OrderItem[];
  created_at: string;
  updated_at: string;
}

export interface OrderStatusCount {
  status: string;
  count: number;
}

export interface TopProduct {
  id: number;
  name: string;
  image_url: string;
  total_sold: number;
}

export interface DashboardStats {
  total_orders: number;
  total_revenue: number;
  total_users: number;
  pending_orders: number;
  orders_by_status: OrderStatusCount[];
  new_users_this_month: number;
  recent_orders: AdminOrder[];
  top_products: TopProduct[];
}

// Address / shipping / loyalty / returns
export interface UserAddress {
  id: number;
  full_name: string;
  phone: string;
  street: string;
  city: string;
  state: string;
  postal_code: string;
  country_code: string;
  is_default: boolean;
}

export interface ShippingMethod {
  code: string;
  name: string;
  description?: string;
}

export interface LoyaltyInfo {
  tier: string;
  points: number;
}

export interface ReturnItem {
  id: number;
  order_item_id: number;
  product?: Product;
  quantity: number;
  refund_amount: number;
  status?: string;
}

export interface ReturnRequest {
  id: number;
  order_id: number;
  status: string;
  reason?: string;
  note?: string;
  items: ReturnItem[];
  created_at: string;
  updated_at: string;
}
