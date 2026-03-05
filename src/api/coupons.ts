import api from './client';
import { ApiResponse } from '../types';

export interface Coupon {
  id: number;
  code: string;
  discount_type: 'percent' | 'fixed';
  discount_value: number;
  min_order_amount: number;
  max_discount_amount: number;
  usage_limit: number;
  used_count: number;
  start_date: string;
  end_date: string;
  is_active: boolean;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface CouponListResponse {
  coupons: Coupon[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

export interface ValidateCouponResponse {
  coupon: Coupon;
  discount: number;
  final_total: number;
}

export interface CreateCouponRequest {
  code: string;
  discount_type: 'percent' | 'fixed';
  discount_value: number;
  min_order_amount?: number;
  max_discount_amount?: number;
  usage_limit?: number;
  start_date: string;
  end_date: string;
  description?: string;
  is_active: boolean;
}

export const couponsApi = {
  // Public - validate coupon
  validateCoupon: async (code: string, orderTotal: number): Promise<ValidateCouponResponse> => {
    const response = await api.post<ApiResponse<ValidateCouponResponse>>('/coupons/validate', {
      code,
      order_total: orderTotal,
    });
    return response.data.data;
  },

  // Public - get available coupons for a given order total (optional)
  getAvailable: async (orderTotal?: number): Promise<Coupon[]> => {
    const response = await api.get<ApiResponse<{ coupons: Coupon[] | null }>>('/coupons/available', {
      params: orderTotal ? { order_total: orderTotal } : undefined,
    });
    return response.data.data.coupons || [];
  },

  // Admin - get all coupons
  getAll: async (page = 1, limit = 20): Promise<CouponListResponse> => {
    const response = await api.get<ApiResponse<CouponListResponse>>(`/admin/coupons`, {
      params: { page, limit },
    });
    return response.data.data;
  },

  // Admin - get coupon by ID
  getById: async (id: number): Promise<Coupon> => {
    const response = await api.get<ApiResponse<Coupon>>(`/admin/coupons/${id}`);
    return response.data.data;
  },

  // Admin - create coupon
  create: async (data: CreateCouponRequest): Promise<Coupon> => {
    const response = await api.post<ApiResponse<Coupon>>('/admin/coupons', data);
    return response.data.data;
  },

  // Admin - update coupon
  update: async (id: number, data: CreateCouponRequest): Promise<Coupon> => {
    const response = await api.put<ApiResponse<Coupon>>(`/admin/coupons/${id}`, data);
    return response.data.data;
  },

  // Admin - delete coupon
  delete: async (id: number): Promise<void> => {
    await api.delete(`/admin/coupons/${id}`);
  },
};

export default couponsApi;
