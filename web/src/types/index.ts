export interface ApiResponse<T = any> {
  code: number;
  message?: string;
  data?: T;
}

export interface TaskData {
  task_id: string;
  status: TaskStatus;
}

export type TaskStatus = 'pending' | 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled';

export interface CreativeData {
  id: string;
  format: string;
  image_url: string;
  width: number;
  height: number;
}

export interface TaskDetailData {
  task_id: string;
  status: TaskStatus;
  title: string;
  progress: number;
  error?: string;
  creatives?: CreativeData[];
  created_at?: string;
  completed_at?: string;
  selling_points?: string[];
  product_image_url?: string;
  requested_formats?: string[];
  style?: string;
  cta_text?: string;
  num_variants?: number;
}

export interface GenerateRequest {
  title: string;
  selling_points: string[];
  product_image_url?: string;
  requested_formats: string[];
  style?: string;
  cta_text?: string;
  num_variants: number;
}

export interface AssetData {
  id: string;
  task_id: number;
  format: string;
  width: number;
  height: number;
  file_size?: number;
  storage_type: string;
  public_url: string;
  image_url?: string;
  style?: string;
  created_at: string;
  updated_at: string;
}

export interface AssetsListData {
  assets: AssetData[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ListAssetsParams {
  page?: number;
  page_size?: number;
  format?: string;
  task_id?: string;
}

export interface TaskListItem {
  id: string;
  title: string;
  status: TaskStatus;
  progress: number;
  created_at: string;
  completed_at?: string;
  error_message?: string;
}

export interface TasksListData {
  tasks: TaskListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ListTasksParams {
  page?: number;
  page_size?: number;
  status?: string;
}
