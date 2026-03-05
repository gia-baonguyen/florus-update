import api from './client';
import { ApiResponse, AuthResponse, User } from '../types';

export interface UpdateProfileRequest {
  name?: string;
  phone?: string;
  address?: string;
}

export interface ChangePasswordRequest {
  current_password: string;
  new_password: string;
}

export const authApi = {
  login: async (email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/login', {
      email,
      password,
    });
    return response.data.data;
  },

  loginWithGoogle: async (idToken: string): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/google', {
      id_token: idToken,
    });
    return response.data.data;
  },

  register: async (name: string, email: string, password: string): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/auth/register', {
      name,
      email,
      password,
    });
    return response.data.data;
  },

  getMe: async (): Promise<User> => {
    const response = await api.get<User>('/auth/me');
    return response.data;
  },

  updateProfile: async (data: UpdateProfileRequest): Promise<User> => {
    const response = await api.put<ApiResponse<User>>('/auth/profile', data);
    return response.data.data;
  },

  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await api.put<ApiResponse<null>>('/auth/password', data);
  },

  // Forgot / reset password (demo)
  forgotPassword: async (email: string): Promise<void> => {
    await api.post<ApiResponse<null>>('/auth/forgot-password', { email });
  },

  resetPassword: async (email: string, newPassword: string): Promise<void> => {
    await api.post<ApiResponse<null>>('/auth/reset-password', {
      email,
      new_password: newPassword,
    });
  },
};

export default authApi;
