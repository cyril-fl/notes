type LogLevel = 'info' | 'error' | 'warn' | 'debug';
type LogContext = string;
interface Logger {
  info: (context: LogContext, ...data: unknown[]) => void;
  error: (context: LogContext, ...data: unknown[]) => void;
  warn: (context: LogContext, ...data: unknown[]) => void;
  debug?: (context: LogContext, ...data: unknown[]) => void;
}

const LOG_PREFIXES = {
  DATA_API: '[DATA API]',
  SUBMIT: '[HANDLE SUBMIT]',
  PLUGIN: '[PLUGIN]',
  COMPONENT: '[COMPONENT]',
} as const;

function resolvePrefix(prefix: string, ctx?: string): string {
  return ctx ? `${prefix} - ${ctx}` : prefix;
}

function createLogger(prefix: string): Logger {
  const logLevels: LogLevel[] = ['info', 'error', 'warn', 'debug'];

  return logLevels.reduce((logger, level) => {
    logger[level] = (context: string, ...data: unknown[]) => {
      const consoleMethod = console[level] || console.log;
      consoleMethod(resolvePrefix(prefix, context), ...data);
    };
    return logger;
  }, {} as Logger);
}

export const logApi = createLogger(LOG_PREFIXES.DATA_API);
export const logSubmit = createLogger(LOG_PREFIXES.SUBMIT);
export const logPlugin = createLogger(LOG_PREFIXES.PLUGIN);
export const logComponent = createLogger(LOG_PREFIXES.COMPONENT);
