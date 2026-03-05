import api from './client';
import { ApiResponse } from '../types';

export interface UploadResponse {
  url: string;
}

export interface MultipleUploadResponse {
  urls: string[];
}

export const uploadApi = {
  // Upload a single image file
  uploadImage: async (file: File, folder: string = 'reviews'): Promise<string> => {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('folder', folder);

    const response = await api.post<ApiResponse<UploadResponse>>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data.data.url;
  },

  // Upload multiple image files
  uploadMultipleImages: async (files: File[], folder: string = 'reviews'): Promise<string[]> => {
    const formData = new FormData();
    files.forEach(file => {
      formData.append('files', file);
    });
    formData.append('folder', folder);

    const response = await api.post<ApiResponse<MultipleUploadResponse>>('/upload/multiple', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return response.data.data.urls;
  },

  // Upload a single image from base64 data URL
  uploadBase64: async (base64Data: string, folder: string = 'reviews'): Promise<string> => {
    const response = await api.post<ApiResponse<UploadResponse>>('/upload/base64', {
      image: base64Data,
      folder,
    });
    return response.data.data.url;
  },

  // Upload multiple images from base64 data URLs
  uploadMultipleBase64: async (base64DataArray: string[], folder: string = 'reviews'): Promise<string[]> => {
    const response = await api.post<ApiResponse<MultipleUploadResponse>>('/upload/base64/multiple', {
      images: base64DataArray,
      folder,
    });
    return response.data.data.urls;
  },
};

export default uploadApi;
