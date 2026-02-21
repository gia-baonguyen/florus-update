import api from './client';
import { ApiResponse, Product } from '../types';

export interface WishlistItem {
  id: number;
  product_id: number;
  product: Product;
  created_at: string;
}

export interface WishlistResponse {
  items: WishlistItem[];
  total_items: number;
}

export const wishlistApi = {
  getWishlist: async (): Promise<WishlistResponse> => {
    const response = await api.get<ApiResponse<WishlistResponse>>('/wishlist');
    return response.data.data;
  },

  addToWishlist: async (productId: number): Promise<WishlistItem> => {
    const response = await api.post<ApiResponse<WishlistItem>>('/wishlist', {
      product_id: productId,
    });
    return response.data.data;
  },

  removeFromWishlist: async (productId: number): Promise<void> => {
    await api.delete(`/wishlist/${productId}`);
  },

  checkInWishlist: async (productId: number): Promise<boolean> => {
    const response = await api.get<ApiResponse<{ in_wishlist: boolean }>>(`/wishlist/check/${productId}`);
    return response.data.data.in_wishlist;
  },

  getWishlistIds: async (): Promise<number[]> => {
    const response = await api.get<ApiResponse<{ product_ids: number[] }>>('/wishlist/ids');
    return response.data.data.product_ids || [];
  },
};

export default wishlistApi;
