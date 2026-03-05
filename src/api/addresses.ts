import { api } from './client';
import type { ApiResponse, UserAddress } from '../types';

export const addressesApi = {
  getMyAddresses: async (): Promise<UserAddress[]> => {
    const response = await api.get<ApiResponse<UserAddress[]>>('/addresses');
    // Backend may return null when there are no addresses; normalize to []
    return response.data.data || [];
  },

  createAddress: async (payload: Omit<UserAddress, 'id' | 'is_default'> & { is_default?: boolean }): Promise<UserAddress> => {
    const response = await api.post<ApiResponse<UserAddress>>('/addresses', {
      full_name: payload.full_name,
      phone: payload.phone,
      street: payload.street,
      city: payload.city,
      state: payload.state,
      postal_code: payload.postal_code,
      country_code: payload.country_code,
      is_default: payload.is_default ?? false,
    });
    return response.data.data;
  },

  updateAddress: async (
    id: number,
    payload: Partial<Omit<UserAddress, 'id' | 'is_default'>> & { is_default?: boolean }
  ): Promise<UserAddress> => {
    const response = await api.put<ApiResponse<UserAddress>>(`/addresses/${id}`, payload);
    return response.data.data;
  },

  deleteAddress: async (id: number): Promise<void> => {
    await api.delete(`/addresses/${id}`);
  },

  setDefault: async (id: number): Promise<void> => {
    await api.put(`/addresses/${id}/default`);
  },
};

