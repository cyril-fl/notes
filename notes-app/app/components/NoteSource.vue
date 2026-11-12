<script setup lang="ts">
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { markdown } from '@codemirror/lang-markdown';
import { defaultHighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { highlightSelectionMatches, selectNextOccurrence, selectSelectionMatches } from '@codemirror/search';
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

const box = ref<HTMLElement>();
let view: EditorView | undefined;
const editable = new Compartment();

// CodeMirror is wired for the multi-cursor: Alt+click, Ctrl+G (next occurrence) and Ctrl+Shift+G (all of them),
// which ProseMirror cannot do (#112), plus the markdown colors. The find and replace panel of CodeMirror is left
// out: its keys fought the browser's and its look was not the app's. The work is on the `source-find-panel` branch.
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
        EditorState.allowMultipleSelections.of(true),
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
        keymap.of([
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
