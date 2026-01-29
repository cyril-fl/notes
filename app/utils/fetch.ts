export type BodyRequest = Record<string, unknown> | undefined;

export interface FetchOptions<T extends BodyRequest = undefined> {
  url: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  body?: T;
}

export type FetchResult<R> =
  | { ok: true; response: ServerResponse<R> }
  | { ok: false; error: Error };

/* TODO
  - transforme fetch en builder avec callback on success et on error dazns un utils externe
  - rendre le code optimisete
  - gerer les todo depuis le server
*/

interface Props {
  isLoading: Ref<boolean>;
  error: Ref<Error | null>;
}

interface FetchBuilder<R> extends PromiseLike<FetchResult<R>> {
  onSuccess: (
    callback: (response: ServerResponse<R>) => void
  ) => FetchBuilder<R>;
  onError: (callback: (error: Error) => void) => FetchBuilder<R>;
}

export function useFetchApi({ isLoading, error }: Props) {
  return function fetch<R, T extends BodyRequest = undefined>(
    options: FetchOptions<T>,
    ctx?: string
  ): FetchBuilder<R> {
    const prefix = ctx ? `[DATA API] - ${ctx}` : `[DATA API]`;

    let onSuccessCallback: ((response: ServerResponse<R>) => void) | null =
      null;
    let onErrorCallback: ((error: Error) => void) | null = null;
    let pendingPromise: Promise<FetchResult<R>> | null = null;

    const execute = async (): Promise<FetchResult<R>> => {
      try {
        if (isLoading.value) throw new Error('Already loading...');

        isLoading.value = true;
        error.value = null;

        const response = await $fetch<ServerResponse<R>>(
          `/api/${options.url}`,
          {
            method: options.method,
            body: options.body,
            headers: {},
          }
        );

        console.info(prefix + ' response', response);

        await onSuccessCallback?.(response);
        return { ok: true, response };
      } catch (e) {
        const err = e instanceof Error ? e : new Error(String(e));
        error.value = err;

        console.error(prefix + ' error', err);

        onErrorCallback?.(err);
        return { ok: false, error: err };
      } finally {
        isLoading.value = false;
      }
    };

    const getPromise = (): Promise<FetchResult<R>> => {
      if (!pendingPromise) {
        pendingPromise = execute();
      }
      return pendingPromise;
    };

    const builder: FetchBuilder<R> = {
      onSuccess(callback) {
        onSuccessCallback = callback;
        return builder;
      },
      onError(callback) {
        onErrorCallback = callback;
        return builder;
      },
      then(onfulfilled, onrejected) {
        return getPromise().then(onfulfilled, onrejected);
      },
    };

    return builder;
  };
}
