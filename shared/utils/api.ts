// Guard
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const statusCodes = [400, 401, 403, 404, 500] as const;

type StatusCode = (typeof statusCodes)[number];

export function guard<T>(
  value: T,
  message: string,
  statusCode: StatusCode = 500
): asserts value is NonNullable<T> {
  if (!value) {
    throw createError({
      statusCode,
      statusMessage: message,
    });
  }
}
