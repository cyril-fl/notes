import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { themes } from '../app/utils/themes.ts';

const css = readFileSync(new URL('../app/assets/css/main.css', import.meta.url), 'utf8');
const palette: Record<string, string> = {};
for (const [, name, hex] of css.matchAll(/--color-([a-z-]+-\d+):\s*(#[0-9a-f]{6})/g)) palette[name!] = hex!;

const rgb = (hex: string) => [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16));
const luminance = (hex: string) => {
  const [r, g, b] = rgb(hex).map((v) => (v / 255 <= 0.03928 ? v / 255 / 12.92 : ((v / 255 + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r! + 0.7152 * g! + 0.0722 * b!;
};
const contrast = (a: string, b: string) => {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi! + 0.05) / (lo! + 0.05);
};
// a at the given opacity over b, as bg-primary/10 draws it.
const over = (a: string, b: string, alpha: number) =>
  '#' + rgb(a).map((v, i) => Math.round(v * alpha + rgb(b)[i]! * (1 - alpha)).toString(16).padStart(2, '0')).join('');

// The shades Nuxt UI takes for each token, with the light accent on the 600 of main.css.
for (const theme of themes) {
  for (const dark of [false, true]) {
    test(`${theme.id} ${dark ? 'dark' : 'light'} reaches WCAG AA`, () => {
      const grey = (shade: number) => palette[`${theme.neutral}-${shade}`]!;
      const primary = palette[`${theme.primary}-${dark ? 400 : 600}`]!;
      const bg = dark ? grey(900) : '#ffffff';
      const frame = dark ? grey(800) : grey(50);
      const elevated = dark ? grey(800) : grey(100);
      const text = dark ? grey(200) : grey(700);
      const muted = dark ? grey(400) : grey(500);
      const pairs = {
        'text on the page': [text, bg],
        'text on an accented row': [text, dark ? grey(700) : grey(200)],
        'muted text on the page': [muted, bg],
        'muted text on a menu': [muted, elevated],
        'muted text on the frame': [muted, frame],
        'a link': [primary, bg],
        'the active row': [primary, over(primary, bg, 0.1)],
        'a button': [dark ? grey(900) : '#ffffff', primary],
        'a selected row': [text, over(primary, bg, 0.15)],
      };
      for (const [name, [fg, back]] of Object.entries(pairs)) {
        assert.ok(contrast(fg!, back!) >= 4.5, `${name}: ${fg} on ${back} is ${contrast(fg!, back!).toFixed(2)}:1`);
      }
    });
  }
}
