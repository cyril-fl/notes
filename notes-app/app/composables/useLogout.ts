/** Logs this device out, then opens the login page. */
export function useLogout() {
  const api = useApi();

  return async function logout() {
    await api('/api/logout', { method: 'POST' });
    clearNuxtData(); // The next account must not see what this one read.
    // The graph's layout, kept in the browser before the server laid it out (#153): it names this account's notes and tags.
    indexedDB.deleteDatabase('notes-graph');
    await navigateTo('/login');
  };
}
