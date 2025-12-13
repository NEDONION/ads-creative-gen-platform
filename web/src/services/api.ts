import axios, { AxiosInstance } from 'axios';
import type {
  ApiResponse,
  GenerateRequest,
  TaskData,
  TaskDetailData,
  AssetsListData,
  ListAssetsParams,
  TasksListData,
  ListTasksParams,
} from '../types';

const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1';

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.response.use(
  response => response,
  error => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

export const creativeAPI = {
  generate: async (data: GenerateRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/creative/generate', data);
    return response.data;
  },

  getTask: async (taskId: string): Promise<ApiResponse<TaskDetailData>> => {
    const response = await apiClient.get<ApiResponse<TaskDetailData>>(`/creative/task/${taskId}`);
    return response.data;
  },

  listAssets: async (params: ListAssetsParams): Promise<ApiResponse<AssetsListData>> => {
    const response = await apiClient.get<ApiResponse<AssetsListData>>('/creative/assets', { params });
    return response.data;
  },

  listTasks: async (params: ListTasksParams): Promise<ApiResponse<TasksListData>> => {
    const response = await apiClient.get<ApiResponse<TasksListData>>('/creative/tasks', { params });
    return response.data;
  },
};

export const healthAPI = {
  check: async (): Promise<any> => {
    const response = await axios.get('/health');
    return response.data;
  },

  ping: async (): Promise<ApiResponse> => {
    const response = await apiClient.get<ApiResponse>('/ping');
    return response.data;
  },
};

export default apiClient;
