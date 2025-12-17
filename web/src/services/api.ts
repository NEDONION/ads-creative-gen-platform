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
  WarmupStats,
} from '../types';

const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1';
const FALLBACK_API_BASE = import.meta.env.VITE_API_BASE_FALLBACK || 'http://localhost:4000/api/v1';

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

type CacheEntry<T> = {
  data: T;
  expiresAt: number;
};

const CACHE_TTL_MS = 60_000;
const responseCache = new Map<string, CacheEntry<any>>();

const normalizeParams = (params?: Record<string, any>): string => {
  if (!params) return '';
  const sortedKeys = Object.keys(params).sort();
  const normalized: Record<string, any> = {};
  sortedKeys.forEach(key => {
    const value = params[key];
    if (value !== undefined) normalized[key] = value;
  });
  return JSON.stringify(normalized);
};

const makeCacheKey = (url: string, params?: Record<string, any>): string => `${url}?${normalizeParams(params)}`;

const getFromCache = <T>(key: string): T | undefined => {
  const entry = responseCache.get(key);
  if (!entry) return undefined;
  if (Date.now() > entry.expiresAt) {
    responseCache.delete(key);
    return undefined;
  }
  return entry.data as T;
};

const setCache = <T>(key: string, data: T) => {
  responseCache.set(key, { data, expiresAt: Date.now() + CACHE_TTL_MS });
};

const clearCache = () => {
  responseCache.clear();
};

const cachedGet = async <T>(url: string, params?: Record<string, any>): Promise<T> => {
  const cacheKey = makeCacheKey(url, params);
  const cached = getFromCache<T>(cacheKey);
  if (cached) return cached;
  const response = await apiClient.get<T>(url, { params });
  setCache(cacheKey, response.data);
  return response.data;
};

const normalizeWarmupResponse = (payload: any): ApiResponse<WarmupStats> => {
  // 尝试把字符串解析成 JSON
  if (typeof payload === 'string') {
    try {
      const parsed = JSON.parse(payload);
      return normalizeWarmupResponse(parsed);
    } catch {
      return { code: -1, message: payload };
    }
  }

  const extractStats = (val: any): WarmupStats | undefined => {
    if (!val || typeof val !== 'object') return undefined;
    if ('runs' in val || 'recent' in val) return val as WarmupStats;
    return undefined;
  };

  // 先尝试直接取 WarmupStats
  const direct = extractStats(payload);
  if (direct) return { code: 0, data: direct };

  // 尝试 { data: WarmupStats }
  if (payload && typeof payload === 'object' && 'data' in payload) {
    const stats = extractStats((payload as any).data);
    const codeRaw = (payload as any).code;
    const code = typeof codeRaw === 'number' ? codeRaw : Number(codeRaw ?? 0);
    if (stats) return { code: code ?? 0, data: stats, message: (payload as any).message };
  }

  // 兜底：如果有 message 则返回
  if (payload && typeof payload === 'object' && 'message' in payload) {
    return { code: -1, message: (payload as { message?: string }).message };
  }

  return { code: -1, message: 'Invalid warmup response' };
};

export const creativeAPI = {
  generate: async (data: GenerateRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/creative/generate', data);
    clearCache();
    return response.data;
  },

  getTask: async (taskId: string): Promise<ApiResponse<TaskDetailData>> => {
    return cachedGet<ApiResponse<TaskDetailData>>(`/creative/task/${taskId}`);
  },

  listAssets: async (params: ListAssetsParams): Promise<ApiResponse<AssetsListData>> => {
    return cachedGet<ApiResponse<AssetsListData>>('/creative/assets', params);
  },

  listTasks: async (params: ListTasksParams): Promise<ApiResponse<TasksListData>> => {
    return cachedGet<ApiResponse<TasksListData>>('/creative/tasks', params);
  },

  generateCopywriting: async (data: GenerateCopywritingRequest): Promise<ApiResponse<CopywritingCandidates>> => {
    const response = await apiClient.post<ApiResponse<CopywritingCandidates>>('/copywriting/generate', data);
    clearCache();
    return response.data;
  },

  confirmCopywriting: async (data: ConfirmCopywritingRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/copywriting/confirm', data);
    clearCache();
    return response.data;
  },

  startCreative: async (data: StartCreativeRequest): Promise<ApiResponse<TaskData>> => {
    const response = await apiClient.post<ApiResponse<TaskData>>('/creative/start', data);
    clearCache();
    return response.data;
  },

  deleteTask: async (taskId: string): Promise<ApiResponse<DeleteTaskResponse>> => {
    const response = await apiClient.delete<ApiResponse<DeleteTaskResponse>>(`/creative/task/${taskId}`);
    clearCache();
    return response.data;
  },
};

export const experimentAPI = {
  create: async (payload: { name: string; product_name?: string; variants: ExperimentVariantInput[] }): Promise<ApiResponse<{ experiment_id: string; status: string }>> => {
    const res = await apiClient.post<ApiResponse<{ experiment_id: string; status: string }>>('/experiments', payload);
    clearCache();
    return res.data;
  },
  list: async (params: { page?: number; page_size?: number; status?: string } = {}): Promise<ApiResponse<ExperimentsListData>> => {
    return cachedGet<ApiResponse<ExperimentsListData>>('/experiments', params);
  },
  updateStatus: async (id: string, status: string): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/status`, { status });
    clearCache();
    return res.data;
  },
  assign: async (id: string, userKey?: string): Promise<ApiResponse<ExperimentAssignData>> => {
    return cachedGet<ApiResponse<ExperimentAssignData>>(`/experiments/${id}/assign`, { user_key: userKey });
  },
  hit: async (id: string, creativeId: number): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/hit`, { creative_id: creativeId });
    clearCache();
    return res.data;
  },
  click: async (id: string, creativeId: number): Promise<ApiResponse<any>> => {
    const res = await apiClient.post<ApiResponse<any>>(`/experiments/${id}/click`, { creative_id: creativeId });
    clearCache();
    return res.data;
  },
  metrics: async (id: string): Promise<ApiResponse<ExperimentMetrics>> => {
    return cachedGet<ApiResponse<ExperimentMetrics>>(`/experiments/${id}/metrics`);
  },
};

export const traceAPI = {
  list: async (params: { page?: number; page_size?: number; status?: string; model_name?: string; trace_id?: string; product_name?: string } = {}): Promise<ApiResponse<TraceListData>> => {
    return cachedGet<ApiResponse<TraceListData>>('/model_traces', params);
  },
  detail: async (traceId: string): Promise<ApiResponse<TraceItem>> => {
    return cachedGet<ApiResponse<TraceItem>>(`/model_traces/${traceId}`);
  },
};

export const warmupAPI = {
  status: async (): Promise<ApiResponse<WarmupStats>> => {
    const call = async (base: string): Promise<any> => {
      const res = await axios.get<ApiResponse<WarmupStats> | WarmupStats | { message?: string }>(`${base}/warmup/status`, {
        headers: { Accept: 'application/json' },
      });
      return res.data;
    };

    let payload = await call(API_BASE);
    console.log('[warmup/status] raw payload:', payload);

    // 若返回 HTML（可能是前端入口），尝试 fallback base
    if (typeof payload === 'string' && payload.toLowerCase().includes('<!doctype')) {
      console.warn('[warmup/status] got HTML, retry with fallback base:', FALLBACK_API_BASE);
      payload = await call(FALLBACK_API_BASE);
      console.log('[warmup/status] fallback payload:', payload);
    }

    return normalizeWarmupResponse(payload);
  },
  run: async (): Promise<ApiResponse<WarmupStats>> => {
    const call = async (base: string): Promise<any> => {
      const res = await axios.post<ApiResponse<WarmupStats> | WarmupStats | { message?: string }>(`${base}/warmup/run`);
      return res.data;
    };

    let payload = await call(API_BASE);
    console.log('[warmup/run] raw payload:', payload);

    if (typeof payload === 'string' && payload.toLowerCase().includes('<!doctype')) {
      console.warn('[warmup/run] got HTML, retry with fallback base:', FALLBACK_API_BASE);
      payload = await call(FALLBACK_API_BASE);
      console.log('[warmup/run] fallback payload:', payload);
    }

    clearCache();
    return normalizeWarmupResponse(payload);
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
