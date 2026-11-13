/** The sections of the settings, in order: one page, or one page each with the Pro mode's settings nav. A feature hides its section without it, and an admin section shows for an admin account only. */
export const settingsSections = [
  { id: 'general', icon: 'settings' },
  { id: 'appearance', icon: 'appearance' },
  { id: 'editor', icon: 'snippet' },
  { id: 'data', icon: 'data' },
  { id: 'links', icon: 'linkBroken', feature: 'brokenLinks' },
  { id: 'pro', icon: 'pro' },
  { id: 'accounts', icon: 'accounts', admin: true },
] as const;

export type SettingsSection = (typeof settingsSections)[number]['id'];
