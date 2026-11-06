export default defineAppConfig({
  ui: {
    // Empty states sit in the page, not in a bordered box, in the middle of the space under the page header.
    // shortcut: 12rem is the layout padding plus the page header, adjust it if the header grows.
    empty: { slots: { root: 'min-h-[calc(100dvh-12rem)]' }, defaultVariants: { variant: 'naked' } },
  },
});
