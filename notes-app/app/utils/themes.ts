// The themes an account picks from: the Nuxt UI colors they set, palettes of Tailwind or of main.css.
// Their ids are repeated in domain.Themes, the API refuses the others.
export const themes = [
  { id: 'notes', primary: 'blue', neutral: 'plain' },
  { id: 'dracula', primary: 'dracula', neutral: 'dracula-gray' },
  { id: 'catppuccin', primary: 'catppuccin', neutral: 'catppuccin-gray' },
  { id: 'nord', primary: 'nord', neutral: 'nord-gray' },
  { id: 'gruvbox', primary: 'gruvbox', neutral: 'gruvbox-gray' },
];

export type Theme = (typeof themes)[number];
