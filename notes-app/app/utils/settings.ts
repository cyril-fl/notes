/** The sections of the settings, in order: one page, or one page each with the Pro mode's settings nav. A feature hides its section without it. */
export const settingsSections = [
  { id: 'general', icon: 'settings' },
  { id: 'appearance', icon: 'appearance' },
  { id: 'editor', icon: 'snippet' },
  { id: 'data', icon: 'data' },
  { id: 'links', icon: 'linkBroken', feature: 'brokenLinks' },
  { id: 'pro', icon: 'pro' },
] as const;

export type SettingsSection = (typeof settingsSections)[number]['id'];
