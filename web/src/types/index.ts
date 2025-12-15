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
  title?: string;
  product_name?: string;
  cta_text?: string;
  selling_points?: string[];
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
  product_name?: string;
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

// 文案生成与确认
export type LanguageOption = 'auto' | 'zh' | 'en';

export interface GenerateCopywritingRequest {
  product_name: string;
  language?: LanguageOption;
}

export interface VariantConfig {
  style?: string;
  prompt?: string;
}

export interface CopywritingCandidates {
  task_id: string;
  cta_candidates: string[];
  selling_point_candidates: string[];
}

export interface ConfirmCopywritingRequest {
  task_id: string;
  selected_cta_index: number;
  selected_sp_indexes: number[];
  edited_cta?: string;
  edited_sps?: string[];
  product_image_url?: string;
  style?: string;
  num_variants?: number;
  formats?: string[];
  variant_configs?: VariantConfig[];
}

export interface StartCreativeRequest {
  task_id: string;
  product_image_url?: string;
  style?: string;
  num_variants?: number;
  formats?: string[];
  variant_configs?: VariantConfig[];
}

export interface AssetData {
  id: string;
  numeric_id?: number;
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
  title?: string;
  product_name?: string;
  cta_text?: string;
  selling_points?: string[];
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
  product_name?: string;
  cta_text?: string;
  selling_points?: string[];
  first_image?: string;
}

export interface DeleteTaskResponse {
  task_id: string;
  status: string;
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

// Experiments
export interface ExperimentVariantInput {
  creative_id: number | string;
  weight: number;
  bucket_start?: number;
  bucket_end?: number;
  title?: string;
  product_name?: string;
  image_url?: string;
  cta_text?: string;
  selling_points?: string[];
}

export interface Experiment {
  experiment_id: string;
  name: string;
  product_name?: string;
  status: string;
  created_at?: string;
  start_at?: string;
  end_at?: string;
  variants?: ExperimentVariantInput[];
}

export interface ExperimentMetrics {
  experiment_id: string;
  variants: {
    creative_id: number;
    impressions: number;
    clicks: number;
    ctr: number;
  }[];
}

export interface ExperimentAssignData {
  creative_id: number;
  asset_uuid?: string;
  task_id?: number;
  title?: string;
  product_name?: string;
  cta_text?: string;
  selling_points?: string[];
  image_url?: string;
}

export interface ExperimentsListData {
  experiments: Experiment[];
  total: number;
  page: number;
  page_size: number;
}

// Trace 页面
export interface TraceStep {
  step_name: string;
  component: string;
  status: string;
  duration_ms: number;
  start_at: string;
  end_at: string;
  input_preview?: string;
  output_preview?: string;
  error_message?: string;
}

export interface TraceItem {
  trace_id: string;
  model_name: string;
  model_version: string;
  status: string;
  duration_ms: number;
  start_at: string;
  end_at: string;
  source?: string;
  input_preview?: string;
  output_preview?: string;
  error_message?: string;
  steps?: TraceStep[];
}

export interface TraceListData {
  traces: TraceItem[];
  total: number;
  page: number;
  page_size: number;
}
