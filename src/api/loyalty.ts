import { api } from './client';
import type { LoyaltyInfo } from '../types';

interface LoyaltyResponse {
  tier: string;
  points: number;
}

export const loyaltyApi = {
  getMyLoyalty: async (): Promise<LoyaltyInfo> => {
    const response = await api.get<{ success: boolean; message: string; data: LoyaltyResponse }>('/loyalty');
    const data = response.data.data;
    return {
      tier: data.tier,
      points: data.points,
    };
  },
};

