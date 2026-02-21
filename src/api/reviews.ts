import api from './client';
import { ApiResponse, User } from '../types';

export interface Review {
  id: number;
  user_id: number;
  product_id: number;
  rating: number;
  comment: string;
  user: User;
  created_at: string;
  updated_at: string;
}

export interface ReviewListResponse {
  reviews: Review[];
  total_reviews: number;
  average_rating: number;
}

export interface CreateReviewRequest {
  rating: number;
  comment: string;
}

export interface UpdateReviewRequest {
  rating?: number;
  comment?: string;
}

export const reviewsApi = {
  getProductReviews: async (productId: number): Promise<ReviewListResponse> => {
    const response = await api.get<ApiResponse<ReviewListResponse>>(`/products/${productId}/reviews`);
    return response.data.data;
  },

  createReview: async (productId: number, data: CreateReviewRequest): Promise<Review> => {
    const response = await api.post<ApiResponse<Review>>(`/products/${productId}/reviews`, data);
    return response.data.data;
  },

  updateReview: async (reviewId: number, data: UpdateReviewRequest): Promise<Review> => {
    const response = await api.put<ApiResponse<Review>>(`/reviews/${reviewId}`, data);
    return response.data.data;
  },

  deleteReview: async (reviewId: number): Promise<void> => {
    await api.delete(`/reviews/${reviewId}`);
  },

  getUserReview: async (productId: number): Promise<Review | null> => {
    try {
      const response = await api.get<ApiResponse<Review>>(`/products/${productId}/my-review`);
      return response.data.data;
    } catch {
      return null;
    }
  },
};

export default reviewsApi;
