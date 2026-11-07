/** Logs this device out, then opens the login page. */
export function useLogout() {
  const api = useApi();

  return async function logout() {
    await api('/api/logout', { method: 'POST' });
    clearNuxtData(); // The next account must not see what this one read.
    await navigateTo('/login');
  };
}
