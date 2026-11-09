/** The plans, each one including those before it: a new plan is added at the end. Repeated by domain.Plans of notes-go. */
export const plans = ['base', 'pro'] as const;
export type Plan = (typeof plans)[number];

/** A plan, from which the feature is available; off: for no one (unfinished, withdrawn). */
export type Level = Plan | 'off';

/** The plan of the account: one the app does not know (an older app after a new plan) gets the base features. */
export const planOf = (user: { plan: string }): Plan =>
  plans.includes(user.plan as Plan) ? (user.plan as Plan) : 'base';

/** What each feature of the app needs: the one place to move a feature in or out of the Pro mode. */
export const features = {
  graph: 'base',
  focusMode: 'pro',
  contextPanel: 'pro',
  settingsNav: 'pro',
  slashCommands: 'pro',
  tables: 'pro',
  codeHighlight: 'base',
  callouts: 'pro',
  wordCount: 'pro',
  contextMenu: 'pro',
  paletteActions: 'pro',
  sourceMode: 'base',
  noteLock: 'base',
  noteHistory: 'pro',
} as const satisfies Record<string, Level>;

export type Feature = keyof typeof features;

/** The subsection of each feature in the settings' comparison: a Record, so a new feature cannot be left out. */
const groups: Record<Feature, 'display' | 'editing' | 'navigation'> = {
  graph: 'navigation',
  focusMode: 'display',
  contextPanel: 'display',
  settingsNav: 'navigation',
  slashCommands: 'editing',
  tables: 'editing',
  codeHighlight: 'display',
  callouts: 'editing',
  wordCount: 'display',
  contextMenu: 'editing',
  paletteActions: 'navigation',
  sourceMode: 'editing',
  noteLock: 'editing',
  noteHistory: 'editing',
};

/** The features the settings compare between the standard and the Pro mode, by subsection. */
export const listedFeatures = (['display', 'editing', 'navigation'] as const).map((group) => ({
  group,
  features: (Object.keys(features) as Feature[]).filter(
    (feature) => groups[feature] === group && (features[feature] as Level) !== 'off',
  ),
}));

export function unlocked(level: Level, plan: Plan) {
  return level !== 'off' && plans.indexOf(level) <= plans.indexOf(plan);
}
