export default defineAppConfig({
  ui: {
    // The theme: each key takes a Tailwind palette or one declared in main.css.
    colors: { primary: 'notes', secondary: 'slate', neutral: 'plain' },
    // Empty states sit in the page, not in a bordered box, in the middle of the space under the page header.
    // TODO: 12rem is the layout padding plus the page header, adjust it if the header grows.
    empty: { slots: { root: 'min-h-[calc(100dvh-12rem)]' }, defaultVariants: { variant: 'naked' } },
  },
});
