function resolvePrefix(prefix: string, ctx?: string): string {
  return ctx ? `${prefix} - ${ctx}` : prefix;
}

const log = {
  info: (prefix: string, context: string, ...data: unknown[]) => {
    console.info(resolvePrefix(prefix, context), ...data);
  },
  error: (prefix: string, context: string, ...data: unknown[]) => {
    console.error(resolvePrefix(prefix, context), ...data);
  },
  warn: (prefix: string, context: string, ...data: unknown[]) => {
    console.warn(resolvePrefix(prefix, context), ...data);
  },
};

// Data api
const DATA_API = '[DATA API]';

export const logApi = {
  info: (context: string, ...data: unknown[]) => {
    log.info(DATA_API, context, ...data);
  },
  error: (context: string, ...data: unknown[]) => {
    log.error(DATA_API, context, ...data);
  },
  warn: (context: string, ...data: unknown[]) => {
    log.warn(DATA_API, context, ...data);
  },
};

// Submit
const SUBMIT = '[HANDLE SUBMIT]';

export const logSubmit = {
  info: (context: string, ...data: unknown[]) => {
    log.info(SUBMIT, context, ...data);
  },
};
