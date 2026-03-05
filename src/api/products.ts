import api from './client';
import { ApiResponse, Product, Category, Cart } from '../types';

export interface ProductFilterParams {
  page?: number;
  limit?: number;
  categoryId?: number;
  search?: string;
  minPrice?: number;
  maxPrice?: number;
  brand?: string;
  sortBy?: 'newest' | 'price_asc' | 'price_desc' | 'popular' | 'rating';
}

export const productsApi = {
  getAll: async (page = 1, limit = 20, sortBy?: string): Promise<{ products: Product[]; meta: any }> => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('limit', limit.toString());
    if (sortBy) params.append('sort_by', sortBy);

    const response = await api.get<ApiResponse<Product[]>>(`/products?${params}`);
    return { products: response.data.data || [], meta: response.data.meta };
  },

  getAllWithFilter: async (filters: ProductFilterParams): Promise<{ products: Product[]; meta: any }> => {
    const params = new URLSearchParams();
    params.append('page', (filters.page || 1).toString());
    params.append('limit', (filters.limit || 20).toString());
    if (filters.categoryId) params.append('category_id', filters.categoryId.toString());
    if (filters.search) params.append('search', filters.search);
    if (filters.minPrice !== undefined) params.append('min_price', filters.minPrice.toString());
    if (filters.maxPrice !== undefined) params.append('max_price', filters.maxPrice.toString());
    if (filters.brand) params.append('brand', filters.brand);
    if (filters.sortBy) params.append('sort_by', filters.sortBy);

    const response = await api.get<ApiResponse<Product[]>>(`/products?${params}`);
    return { products: response.data.data || [], meta: response.data.meta };
  },

  getByCategory: async (categoryId: number, page = 1, limit = 20): Promise<{ products: Product[]; meta: any }> => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('limit', limit.toString());
    params.append('category_id', categoryId.toString());

    const response = await api.get<ApiResponse<Product[]>>(`/products?${params}`);
    return { products: response.data.data || [], meta: response.data.meta };
  },

  getBrands: async (): Promise<string[]> => {
    const response = await api.get<ApiResponse<string[]>>('/products/brands');
    return response.data.data || [];
  },

  getById: async (id: number): Promise<Product> => {
    const response = await api.get<ApiResponse<Product>>(`/products/${id}`);
    return response.data.data;
  },

  getBySlug: async (slug: string): Promise<Product> => {
    const response = await api.get<ApiResponse<Product>>(`/products/slug/${slug}`);
    return response.data.data;
  },
};

export const categoriesApi = {
  getAll: async (): Promise<Category[]> => {
    const response = await api.get<ApiResponse<Category[]>>('/categories');
    return response.data.data;
  },

  getById: async (id: number): Promise<Category> => {
    const response = await api.get<ApiResponse<Category>>(`/categories/${id}`);
    return response.data.data;
  },
};

export const cartApi = {
  get: async (): Promise<Cart> => {
    const response = await api.get<ApiResponse<Cart>>('/cart');
    return response.data.data;
  },

  addItem: async (productId: number, quantity: number): Promise<Cart> => {
    const response = await api.post<ApiResponse<Cart>>('/cart', {
      product_id: productId,
      quantity,
    });
    return response.data.data;
  },

  // Backend expects cart item ID in the path (not product ID)
  updateItem: async (itemId: number, quantity: number): Promise<Cart> => {
    const response = await api.put<ApiResponse<Cart>>(`/cart/${itemId}`, {
      quantity,
    });
    return response.data.data;
  },

  // Backend expects cart item ID here as well
  removeItem: async (itemId: number): Promise<void> => {
    await api.delete(`/cart/${itemId}`);
  },

  clear: async (): Promise<void> => {
    await api.delete('/cart');
  },
};

export const recommendationsApi = {
  getColdStart: async (): Promise<Product[]> => {
    const response = await api.get<ApiResponse<Product[]>>('/recommendations/cold-start');
    return response.data.data || [];
  },

  getWarmStart: async (): Promise<Product[]> => {
    const response = await api.get<ApiResponse<Product[]>>('/recommendations/warm-start');
    return response.data.data || [];
  },

  getSimilar: async (productId: number): Promise<Product[]> => {
    const response = await api.get<ApiResponse<Product[]>>(`/recommendations/similar/${productId}`);
    return response.data.data || [];
  },

  getCoViewed: async (productId: number): Promise<Product[]> => {
    const response = await api.get<ApiResponse<Product[]>>(`/recommendations/co-viewed/${productId}`);
    return response.data.data || [];
  },

  getCrossSell: async (productId: number): Promise<Product[]> => {
    const response = await api.get<ApiResponse<Product[]>>(`/recommendations/cross-sell/${productId}`);
    return response.data.data || [];
  },
};

export const ordersApi = {
  getAll: async (page = 1, limit = 10): Promise<{ orders: import('../types').Order[]; meta: any }> => {
    const response = await api.get<import('../types').ApiResponse<import('../types').Order[]>>(`/orders?page=${page}&limit=${limit}`);
    return { orders: response.data.data, meta: response.data.meta };
  },

  getById: async (id: number): Promise<import('../types').Order> => {
    const response = await api.get<import('../types').ApiResponse<import('../types').Order>>(`/orders/${id}`);
    return response.data.data;
  },

  create: async (
    shippingAddress: string,
    note?: string,
    couponCode?: string,
    paymentMethod?: string,
    shippingAddressId?: number,
    shippingMethodCode?: string,
    loyaltyPointsToUse?: number
  ): Promise<import('../types').Order> => {
    const response = await api.post<import('../types').ApiResponse<import('../types').Order>>('/orders', {
      shipping_address: shippingAddress,
      note: note || '',
      coupon_code: couponCode || '',
      payment_method: paymentMethod || 'cod',
      shipping_address_id: shippingAddressId,
      shipping_method_code: shippingMethodCode,
      loyalty_points_to_use: loyaltyPointsToUse,
    });
    return response.data.data;
  },

  cancel: async (id: number): Promise<void> => {
    await api.put(`/orders/${id}/cancel`);
  },

  // Buy Now - create order directly from a single product without using cart
  buyNow: async (
    productId: number,
    quantity: number,
    shippingAddress: string,
    note?: string,
    couponCode?: string,
    paymentMethod?: string,
    shippingAddressId?: number,
    shippingMethodCode?: string,
    loyaltyPointsToUse?: number
  ): Promise<import('../types').Order> => {
    const response = await api.post<import('../types').ApiResponse<import('../types').Order>>('/orders/buy-now', {
      product_id: productId,
      quantity: quantity,
      shipping_address: shippingAddress,
      note: note || '',
      coupon_code: couponCode || '',
      payment_method: paymentMethod || 'cod',
      shipping_address_id: shippingAddressId,
      shipping_method_code: shippingMethodCode,
      loyalty_points_to_use: loyaltyPointsToUse,
    });
    return response.data.data;
  },
};

export default { productsApi, categoriesApi, cartApi, recommendationsApi, ordersApi };
