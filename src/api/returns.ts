import { api } from './client';
import type { ApiResponse, ReturnRequest } from '../types';

export const returnsApi = {
  create: async (payload: {
    order_id: number;
    items: { order_item_id: number; quantity: number }[];
    reason?: string;
    note?: string;
  }): Promise<ReturnRequest> => {
    const response = await api.post<ApiResponse<ReturnRequest>>('/returns', payload);
    return response.data.data;
  },

  getMyReturns: async (): Promise<ReturnRequest[]> => {
    const response = await api.get<ApiResponse<ReturnRequest[]>>('/returns');
    return response.data.data;
  },

  getById: async (id: number): Promise<ReturnRequest> => {
    const response = await api.get<ApiResponse<ReturnRequest>>(`/returns/${id}`);
    return response.data.data;
  },
};

