import { api } from './client';
import type { ApiResponse, ShippingMethod } from '../types';

export const shippingApi = {
  getMethods: async (): Promise<ShippingMethod[]> => {
    const response = await api.get<ApiResponse<ShippingMethod[]>>('/shipping/methods');
    // backend currently wraps methods directly in data; may be null if no seed
    return response.data.data || [];
  },
};

