// TODO Refacto
export interface ServerResponse<T = null> {
  message: string;
  data: T;
}

export interface ServerErrorResponse {
  error: Error;
}

export interface RequestBody<T = null> {
  data: T;
}
