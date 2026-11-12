<script setup lang="ts">
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { markdown } from '@codemirror/lang-markdown';
import { defaultHighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { highlightSelectionMatches, search, searchKeymap } from '@codemirror/search';
import { Compartment, EditorState } from '@codemirror/state';
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

// CodeMirror is wired for the multi-cursor: Alt+click, Cmd+D (next occurrence) and Cmd+Shift+L (all of them),
// which ProseMirror cannot do (#112), plus the search panel and the markdown colors.
const theme = EditorView.theme({
  '&': { height: '100%', backgroundColor: 'transparent', color: 'var(--ui-text)' },
  '&.cm-focused': { outline: 'none' },
  '.cm-scroller': { fontFamily: 'var(--font-mono)', fontSize: '0.875rem', lineHeight: '1.6', overflow: 'auto' },
  '.cm-content': { padding: props.compact ? '0.5rem 0.75rem' : '0 2rem 2rem', caretColor: 'var(--ui-primary)' },
  '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--ui-primary)', borderLeftWidth: '2px' },
  // The carets added by Cmd+D are drawn by CodeMirror, and shown even unfocused so the count stays readable.
  '.cm-selectionBackground, &.cm-focused .cm-selectionBackground, .cm-content ::selection': {
    backgroundColor: 'var(--ui-bg-accented)',
  },
  '.cm-selectionMatch': { backgroundColor: 'var(--ui-bg-elevated)' },
  '.cm-panels': { backgroundColor: 'var(--ui-bg-elevated)', color: 'var(--ui-text)', border: 'none' },
  '.cm-panels.cm-panels-top': { borderBottom: '1px solid var(--ui-border)' },
  '.cm-panel input, .cm-panel button': { fontFamily: 'inherit', fontSize: '0.75rem' },
  '.cm-panel input': {
    backgroundColor: 'var(--ui-bg)',
    border: '1px solid var(--ui-border-accented)',
    borderRadius: '0.25rem',
    padding: '0.125rem 0.25rem',
  },
  // Its buttons come with a light gradient of their own: unreadable over the dark mode's white text.
  '.cm-panel button': {
    backgroundColor: 'var(--ui-bg)',
    backgroundImage: 'none',
    color: 'var(--ui-text)',
    border: '1px solid var(--ui-border-accented)',
    borderRadius: '0.25rem',
    padding: '0.125rem 0.375rem',
  },
  '.cm-panel button[name="close"]': {
    backgroundColor: 'transparent',
    border: 'none',
    color: 'var(--ui-text-muted)',
  },
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
        keymap.of([...searchKeymap, ...defaultKeymap, ...historyKeymap]),
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
