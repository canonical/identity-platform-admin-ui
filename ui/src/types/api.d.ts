export interface ApiResponse<T> {
  data: T;
  message: string;
  status: number;
}

export type PaginatedResponse<T> = {
  _meta: {
    next?: string;
    prev?: string;
  };
} & ApiResponse<T>;

export interface ErrorResponse {
  error?: string;
  message?: string;
}
