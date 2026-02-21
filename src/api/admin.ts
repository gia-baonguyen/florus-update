import api from './client';
import { ApiResponse, AdminOrder, AdminUser, DashboardStats, Product, Category } from '../types';

export const adminApi = {
  // Dashboard
  getStats: async (): Promise<DashboardStats> => {
    const response = await api.get<ApiResponse<DashboardStats>>('/admin/stats');
    return response.data.data;
  },

  // Orders
  getOrders: async (page = 1, limit = 10, status?: string): Promise<{ orders: AdminOrder[]; meta: any }> => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('limit', limit.toString());
    if (status) params.append('status', status);

    const response = await api.get<ApiResponse<AdminOrder[]>>(`/admin/orders?${params}`);
    return { orders: response.data.data, meta: response.data.meta };
  },

  getOrderDetail: async (id: number): Promise<AdminOrder> => {
    const response = await api.get<ApiResponse<AdminOrder>>(`/admin/orders/${id}`);
    return response.data.data;
  },

  updateOrderStatus: async (id: number, status: string): Promise<AdminOrder> => {
    const response = await api.put<ApiResponse<AdminOrder>>(`/admin/orders/${id}/status`, { status });
    return response.data.data;
  },

  // Users
  getUsers: async (page = 1, limit = 10): Promise<{ users: AdminUser[]; meta: any }> => {
    const params = new URLSearchParams();
    params.append('page', page.toString());
    params.append('limit', limit.toString());

    const response = await api.get<ApiResponse<AdminUser[]>>(`/admin/users?${params}`);
    return { users: response.data.data, meta: response.data.meta };
  },

  updateUserRole: async (id: number, role: 'admin' | 'user'): Promise<AdminUser> => {
    const response = await api.put<ApiResponse<AdminUser>>(`/admin/users/${id}/role`, { role });
    return response.data.data;
  },

  // Products (using existing endpoints with admin auth)
  createProduct: async (product: Partial<Product>): Promise<Product> => {
    const response = await api.post<ApiResponse<Product>>('/products', product);
    return response.data.data;
  },

  updateProduct: async (id: number, product: Partial<Product>): Promise<Product> => {
    const response = await api.put<ApiResponse<Product>>(`/products/${id}`, product);
    return response.data.data;
  },

  deleteProduct: async (id: number): Promise<void> => {
    await api.delete(`/products/${id}`);
  },

  // Categories (using existing endpoints with admin auth)
  createCategory: async (category: Partial<Category>): Promise<Category> => {
    const response = await api.post<ApiResponse<Category>>('/categories', category);
    return response.data.data;
  },

  updateCategory: async (id: number, category: Partial<Category>): Promise<Category> => {
    const response = await api.put<ApiResponse<Category>>(`/categories/${id}`, category);
    return response.data.data;
  },

  deleteCategory: async (id: number): Promise<void> => {
    await api.delete(`/categories/${id}`);
  },
};

export default adminApi;
