<script setup lang="ts">
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { markdown } from '@codemirror/lang-markdown';
import { defaultHighlightStyle, syntaxHighlighting } from '@codemirror/language';
import {
  closeSearchPanel,
  highlightSelectionMatches,
  openSearchPanel,
  search,
  selectNextOccurrence,
  selectSelectionMatches,
} from '@codemirror/search';
import { Compartment, EditorState, Prec } from '@codemirror/state';
import { crosshairCursor, drawSelection, EditorView, keymap, rectangularSelection } from '@codemirror/view';

// compact: the padding of a box of its own (the snippet dialog), in place of the note's margins.
const props = defineProps<{
  modelValue: string;
  label: string;
  readonly?: boolean;
  autofocus?: boolean;
  compact?: boolean;
}>();
const emit = defineEmits<{ 'update:modelValue': [string] }>();

const { t } = useI18n();
const box = ref<HTMLElement>();
let view: EditorView | undefined;
const editable = new Compartment();

// CodeMirror is wired for the multi-cursor: Alt+click, Ctrl+G (next occurrence) and Ctrl+Shift+G (all of them),
// which ProseMirror cannot do (#112), plus the find panel on Ctrl+F and the markdown colors.
const theme = EditorView.theme({
  '&': { height: '100%', backgroundColor: 'transparent', color: 'var(--ui-text)' },
  '&.cm-focused': { outline: 'none' },
  '.cm-scroller': { fontFamily: 'var(--font-mono)', fontSize: '0.875rem', lineHeight: '1.6', overflow: 'auto' },
  '.cm-content': { padding: props.compact ? '0.5rem 0.75rem' : '0 2rem 2rem', caretColor: 'var(--ui-primary)' },
  '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--ui-primary)', borderLeftWidth: '2px' },
  // The multi-cursor needs drawSelection: CodeMirror draws the selection itself, so it takes the theme's color
  // in place of the browser's blue. The carets it adds are shown even unfocused, so their count stays readable.
  // The layer carries the transparency: an alpha on the color itself comes out opaque through CodeMirror's styles.
  // Its own rule for the focused selection is nested that deep, and only a selector of the same shape wins over it.
  '.cm-selectionLayer': { opacity: '0.3' },
  '.cm-selectionLayer .cm-selectionBackground, &.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground':
    { backgroundColor: 'var(--ui-primary)' },
  '.cm-selectionMatch': { backgroundColor: 'var(--ui-bg-accented)' },
  // The find panel is CodeMirror's own markup, dressed in the app's tokens: the inputs and buttons of Nuxt UI.
  '.cm-panels': { backgroundColor: 'var(--ui-bg-muted)', color: 'var(--ui-text)', border: 'none' },
  '.cm-panels.cm-panels-top': { borderBottom: '1px solid var(--ui-border)' },
  // CodeMirror separates the find row from the replace row with a <br>: a full-width flex item keeps the two rows.
  '.cm-search': {
    display: 'flex',
    flexWrap: 'wrap',
    alignItems: 'center',
    gap: '0.375rem',
    padding: '0.5rem 2rem 0.5rem 0.75rem',
  },
  '.cm-search br': { flexBasis: '100%', height: '0' },
  '.cm-search label': {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '0.25rem',
    margin: '0',
    fontSize: '0.75rem',
    color: 'var(--ui-text-muted)',
  },
  '.cm-search input[type="checkbox"]': { accentColor: 'var(--ui-primary)', margin: '0' },
  '.cm-textfield': {
    fontFamily: 'inherit',
    fontSize: '0.75rem',
    backgroundColor: 'var(--ui-bg)',
    color: 'var(--ui-text)',
    border: '1px solid var(--ui-border-accented)',
    borderRadius: '0.375rem',
    padding: '0.25rem 0.5rem',
    minWidth: '12rem',
    margin: '0',
  },
  '.cm-textfield:focus': { outline: 'none', borderColor: 'var(--ui-primary)' },
  // Its buttons come with a light gradient of their own: unreadable over the dark mode's white text.
  '.cm-button': {
    fontFamily: 'inherit',
    fontSize: '0.75rem',
    backgroundColor: 'var(--ui-bg)',
    backgroundImage: 'none',
    color: 'var(--ui-text)',
    border: '1px solid var(--ui-border-accented)',
    borderRadius: '0.375rem',
    padding: '0.25rem 0.5rem',
    margin: '0',
  },
  '.cm-button:hover': { backgroundColor: 'var(--ui-bg-elevated)' },
  '.cm-search [name="close"]': {
    backgroundColor: 'transparent',
    border: 'none',
    color: 'var(--ui-text-muted)',
    fontSize: '1rem',
    padding: '0 0.25rem',
    top: '0.375rem',
    right: '0.5rem',
  },
  '.cm-search [name="close"]:hover': { color: 'var(--ui-text)' },
  '.cm-searchMatch': { backgroundColor: 'var(--ui-bg-accented)' },
  '.cm-searchMatch-selected': { backgroundColor: 'var(--ui-primary)', color: 'var(--ui-bg)' },
});

// The search panel of CodeMirror writes its own labels: translated here, as every text of the app.
const phrases = () => ({
  Find: t('note.search.find'),
  Replace: t('note.search.replace'),
  next: t('note.search.next'),
  previous: t('note.search.previous'),
  all: t('note.search.all'),
  'match case': t('note.search.matchCase'),
  'by word': t('note.search.byWord'),
  regexp: t('note.search.regexp'),
  replace: t('note.search.replaceOne'),
  'replace all': t('note.search.replaceAll'),
  close: t('note.search.close'),
  'Go to line': t('note.search.goToLine'),
  go: t('note.search.go'),
});

onMounted(() => {
  if (!box.value) return;
  view = new EditorView({
    parent: box.value,
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        history(),
        drawSelection(),
        rectangularSelection(),
        crosshairCursor(),
        highlightSelectionMatches(),
        search({ top: true }),
        EditorState.allowMultipleSelections.of(true),
        EditorState.phrases.of(phrases()),
        EditorView.lineWrapping,
        EditorView.contentAttributes.of({ 'aria-label': props.label }),
        // macOS holds a Ctrl+Shift+letter as its unshifted self, so a keymap entry never sees the Shift: read here.
        Prec.highest(
          EditorView.domEventHandlers({
            keydown(event, view) {
              if (!event.ctrlKey || !event.shiftKey || event.key.toLowerCase() !== 'g') return false;
              return selectSelectionMatches(view);
            },
          }),
        ),
        // Safari keeps Cmd+D for its bookmarks and never hands it over, and macOS reads Ctrl+D as a forward delete:
        // Ctrl+G takes the next occurrence and Ctrl+Shift+G all of them.
        // The find and replace panel opens on Ctrl+F (Cmd+F) and nothing else: searchKeymap put it up on its own
        // keys too (Ctrl+G among them, where the multi-cursor lives), which is not wanted.
        keymap.of([
          { key: 'Mod-f', run: openSearchPanel, scope: 'editor search-panel', preventDefault: true },
          { key: 'Ctrl-f', run: openSearchPanel, scope: 'editor search-panel', preventDefault: true },
          { key: 'Escape', run: closeSearchPanel, scope: 'editor search-panel' },
          { key: 'Ctrl-g', run: selectNextOccurrence, preventDefault: true },
          { key: 'Mod-d', run: selectNextOccurrence, preventDefault: true },
          ...defaultKeymap,
          ...historyKeymap,
        ]),
        markdown(),
        syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
        theme,
        editable.of(EditorState.readOnly.of(!!props.readonly)),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) emit('update:modelValue', update.state.doc.toString());
        }),
      ],
    }),
  });
  if (props.autofocus) view.focus();
});
onBeforeUnmount(() => view?.destroy());

// The text replaced from the outside (a version restored, a note renamed): what was typed here comes back the same, and is left alone.
watch(
  () => props.modelValue,
  (next) => {
    if (!view || next === view.state.doc.toString()) return;
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: next } });
  },
);
watch(
  () => props.readonly,
  (next) => view?.dispatch({ effects: editable.reconfigure(EditorState.readOnly.of(!!next)) }),
);
</script>

<template>
  <div ref="box" class="flex min-h-0 flex-1 flex-col overflow-hidden" />
</template>
