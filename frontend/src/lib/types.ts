export interface Tenant {
  id: number;
  name: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Project {
  id: number;
  tenant_id: number;
  slug: string;
  name: string;
  base_path: string;
  provider: 'gcs' | 's3' | 'rustfs';
  bucket: string;
  region: string;
  access_key_id?: string;
  secret_access_key?: string;
  credentials_json?: string;
  endpoint?: string;
  default_quality: number;
  default_format: string;
  max_width: number;
  max_height: number;
  rps: number;
  burst: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: number;
  username: string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface Invitation {
  id: number;
  code: string;
  is_used: boolean;
  created_by: number;
  used_by?: number | null;
  created_at: string;
  used_at?: string | null;
}

export interface ProjectMetricsDaily {
  date: string;
  requests: number;
  cache_hits: number;
  cache_misses: number;
  origin_transforms: number;
  bandwidth_bytes: number;
  errors: number;
  avg_response_time_ms: number;
}

export interface ProjectSummary {
  id: number;
  tenant_id: number;
  slug: string;
  name: string;
  base_path: string;
  provider: string;
  bucket: string;
  region: string;
  access_key_id?: string;
  secret_access_key?: string;
  credentials_json?: string;
  endpoint?: string;
  default_quality: number;
  default_format: string;
  max_width: number;
  max_height: number;
  rps: number;
  burst: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  requests_today: number;
  cache_hits_today: number;
  cache_misses_today: number;
  bandwidth_bytes_today: number;
  errors_today: number;
  requests_total: number;
  bandwidth_bytes_total: number;
  errors_total: number;
  cache_hit_ratio: number;
}

export interface ProjectBriefMetrics {
  project_id: number;
  slug: string;
  name: string;
  requests: number;
  cache_hits: number;
  bandwidth_bytes: number;
  errors: number;
}

export interface TenantMetricsSummary {
  total_requests: number;
  total_cache_hits: number;
  total_errors: number;
  total_bandwidth_bytes: number;
  cache_hit_ratio: number;
  per_project: ProjectBriefMetrics[];
}

export interface CreateProjectReq {
  slug: string;
  name?: string;
  base_path?: string;
  provider: 'gcs' | 's3' | 'rustfs';
  bucket: string;
  region?: string;
  access_key_id?: string;
  secret_access_key?: string;
  credentials_json?: string;
  endpoint?: string;
  default_quality?: number;
  default_format?: string;
  max_width?: number;
  max_height?: number;
  rps?: number;
  burst?: number;
}

export interface AdminCreateProjectReq extends CreateProjectReq {
  tenant_id: number;
}

export interface RegisterReq {
  name: string;
  email: string;
  password: string;
  invitation_code: string;
}

export interface LoginReq {
  email: string;
  password: string;
}

export interface TokenResp {
  token: string;
  tenant_id?: number;
  role?: string;
}

export interface UserLoginReq {
  username: string;
  password: string;
}

export interface AccountResp {
  id: number;
  name: string;
  email: string;
}

export interface AdminStats {
  tenants: number;
  projects: number;
  users: number;
}

export interface ProjectMetricsResponse {
  daily: ProjectMetricsDaily[];
}

export interface AdminGlobalMetrics {
  total_requests: number;
  total_bandwidth_bytes: number;
  cache_hit_ratio: number;
  per_project: ProjectBriefMetrics[];
}

export interface TenantDetail {
  tenant: Tenant;
  projects: Project[];
}
