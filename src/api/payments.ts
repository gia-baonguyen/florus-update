import api from './client';
import { ApiResponse } from '../types';

export type PaymentMethod = 'cod' | 'zalopay' | 'momo' | 'vnpay';

export interface CreatePaymentRequest {
  order_code: string;
  payment_method: PaymentMethod;
}

export interface CreatePaymentResponse {
  order_code: string;
  payment_url: string;
  method: string;
}

export interface PaymentStatus {
  id: number;
  order_id: number;
  transaction_id: string;
  method: PaymentMethod;
  amount: number;
  status: 'pending' | 'success' | 'failed' | 'cancelled' | 'refunded';
  provider_trans_id?: string;
  payment_url?: string;
  error_message?: string;
  paid_at?: string;
  created_at: string;
}

export const paymentsApi = {
  createPayment: async (data: CreatePaymentRequest): Promise<CreatePaymentResponse> => {
    const response = await api.post<ApiResponse<CreatePaymentResponse>>('/payments/create', data);
    return response.data.data;
  },

  getPaymentStatus: async (orderCode: string): Promise<PaymentStatus> => {
    const response = await api.get<ApiResponse<PaymentStatus>>(`/payments/status/${orderCode}`);
    return response.data.data;
  },
};

export default paymentsApi;
