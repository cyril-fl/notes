/** base: for everyone; pro: with the Pro mode of the account; off: for no one (unfinished, withdrawn). */
export type Level = 'base' | 'pro' | 'off';

/** What each feature of the app needs: the one place to move a feature in or out of the Pro mode. */
export const features = {
  graph: 'base',
  focusMode: 'pro',
  contextPanel: 'pro',
  settingsNav: 'pro',
  slashCommands: 'pro',
  tables: 'pro',
  codeHighlight: 'pro',
  callouts: 'pro',
  wordCount: 'pro',
  contextMenu: 'pro',
} as const satisfies Record<string, Level>;

export type Feature = keyof typeof features;

/** The features the Pro mode unlocks, listed by the settings. */
export const proFeatures = (Object.keys(features) as Feature[]).filter((feature) => features[feature] === 'pro');

export function unlocked(level: Level, pro: boolean) {
  return level === 'base' || (level === 'pro' && pro);
}
