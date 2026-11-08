/** The sections of the settings, in order: one page, or one page each with the Pro mode's settings nav. */
export const settingsSections = [
  { id: 'general', icon: 'settings' },
  { id: 'appearance', icon: 'appearance' },
  { id: 'editor', icon: 'snippet' },
  { id: 'trash', icon: 'trash' },
  { id: 'data', icon: 'data' },
  { id: 'pro', icon: 'pro' },
  { id: 'danger', icon: 'danger' },
] as const;

export type SettingsSection = (typeof settingsSections)[number]['id'];
