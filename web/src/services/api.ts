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
  GenerateCopywritingRequest,
  CopywritingCandidates,
  ConfirmCopywritingRequest,
  StartCreativeRequest,
  DeleteTaskResponse,
  ExperimentVariantInput,
  ExperimentMetrics,
  ExperimentsListData,
  ExperimentAssignData,
  TraceListData,
  TraceItem,
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

  generateCopywriting: async (data: GenerateCopywritingRequest): Promise<ApiResponse<CopywritingCandidates>> => {
    const response = await apiClient.post<ApiResponse<CopywritingCandidates>>('/copywriting/generate', data);
    return response.data;
  },

  confirmCopywriting: async (data: ConfirmCopywritingRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/copywriting/confirm', data);
    return response.data;
  },

  startCreative: async (data: StartCreativeRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/creative/start', data);
    return response.data;
  },

  deleteTask: async (taskId: string): Promise<ApiResponse<DeleteTaskResponse>> => {
    const response = await apiClient.delete<ApiResponse<DeleteTaskResponse>>(`/creative/task/${taskId}`);
    return response.data;
  },
};

export const experimentAPI = {
  create: async (payload: { name: string; product_name?: string; variants: ExperimentVariantInput[] }): Promise<ApiResponse<{ experiment_id: string; status: string }>> => {
    const res = await apiClient.post<ApiResponse<{ experiment_id: string; status: string }>>('/experiments', payload);
    return res.data;
  },
  list: async (params: { page?: number; page_size?: number; status?: string } = {}): Promise<ApiResponse<ExperimentsListData>> => {
    const res = await apiClient.get<ApiResponse<ExperimentsListData>>('/experiments', { params });
    return res.data;
  },
  updateStatus: async (id: string, status: string): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/status`, { status });
    return res.data;
  },
  assign: async (id: string, userKey?: string): Promise<ApiResponse<ExperimentAssignData>> => {
    const res = await apiClient.get<ApiResponse<ExperimentAssignData>>(`/experiments/${id}/assign`, { params: { user_key: userKey } });
    return res.data;
  },
  hit: async (id: string, creativeId: number): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/hit`, { creative_id: creativeId });
    return res.data;
  },
  click: async (id: string, creativeId: number): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/click`, { creative_id: creativeId });
    return res.data;
  },
  metrics: async (id: string): Promise<ApiResponse<ExperimentMetrics>> => {
    const res = await apiClient.get<ApiResponse<ExperimentMetrics>>(`/experiments/${id}/metrics`);
    return res.data;
  },
};

export const traceAPI = {
  list: async (params: { page?: number; page_size?: number; status?: string; model_name?: string; trace_id?: string } = {}): Promise<ApiResponse<TraceListData>> => {
    const res = await apiClient.get<ApiResponse<TraceListData>>('/model_traces', { params });
    return res.data;
  },
  detail: async (traceId: string): Promise<ApiResponse<TraceItem>> => {
    const res = await apiClient.get<ApiResponse<TraceItem>>(`/model_traces/${traceId}`);
    return res.data;
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
