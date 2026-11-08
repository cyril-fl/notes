<script setup lang="ts">
import type { DropdownMenuItem, EditorSuggestionMenuItem } from '@nuxt/ui';
import { createHandlers } from '@nuxt/ui/runtime/utils/editor.js';
import type { Range } from '@tiptap/core';
import type { Editor } from '@tiptap/vue-3';
import type { FetchError } from 'ofetch';
import type { RouteLocationNormalized } from 'vue-router';
import type { Item, ItemDetail, NoteFile, NotesGraph } from '~/types/items';
import type { Changes } from '~/utils/saver';

const { icons } = useIcons();
const { has, ready } = usePro();
const focus = useState('focus', () => false);
const { t } = useI18n();

// start: the starting text of a new note, saved only once typed in.
const props = defineProps<{ noteId?: string; parentId?: string; start?: string }>();
const emit = defineEmits<{ created: [id: string] }>();

const api = useApi();
const { confirm, prompt } = useDialogs();
const { removeItem } = useItems();
const baseURL = useRuntimeConfig().public.apiBase;
const toast = useToast();

const id = ref(props.noteId);
const parent = ref(props.parentId);
const path = ref('');
const title = ref<string>();
useHead({ title: () => displayTitle(title.value, t('untitled')) });
const tags = ref<string[]>([]);
let links = '';
const initialContent = ref(props.start ?? '');
const loading = ref(!!props.noteId);
const notFound = ref(false);
// Locked read-only: the API refuses its content, the editor and the source mode only read.
const readonly = ref(false);

const failed = ref(false);
const saving = ref(false);
const isEmpty = ref(!props.noteId);
// The editor can change its document while loading: only what follows a user interaction is an edit,
// or merely opening a note would rewrite the markdown written with the CLI. Pointerdown too: a task checkbox takes no focus.
const touched = ref(false);
// The version the text is written over (#62): once another device changed the note, the API answers 409 and nothing more is sent.
let version: string | undefined;
// The text on the API as this editor last saw it: a moved note gets a new updated_at, not a new text.
let saved = '';
let stale = false;

const saver = createSaver(send);
const editorRef = useTemplateRef<{ editor?: Editor }>('editor');

async function send(changes: Changes): Promise<boolean | undefined> {
  if (id.value && Object.values(changes).every((value) => value === undefined)) return;
  // The API refuses a note with no content: a new one waits for its first text.
  if (!id.value && !changes.content) return false;
  if (stale) return false;

  saving.value = true;
  try {
    await (id.value ? patchNote(changes) : createNote(changes));
    failed.value = false;
  } catch (error) {
    failed.value = true;
    stale = (error as FetchError).status === 409;
    throw error;
  } finally {
    saving.value = false;
  }
}

async function createNote(changes: Changes) {
  const body = { ...changes, parent_id: parent.value };
  const item = await api<Item>('/api/notes', { method: 'POST', body });
  id.value = item.id;
  emit('created', item.id);
  const note = await api<ItemDetail>(`/api/notes/${item.id}`);
  remember(note);
  await readTags(changes, note);
}

async function patchNote(changes: Changes) {
  const body = { ...changes, updated_at: version };
  const note = await api<ItemDetail>(`/api/notes/${id.value}`, {
    method: 'PATCH',
    body,
    conflict: t('note.changedElsewhere'),
  });
  remember(note);
  await readTags(changes, note);
}

function remember(note: ItemDetail) {
  version = note.item.updated_at ?? note.item.created_at; // A note never edited has no updated_at.
  saved = note.content ?? '';
}

// The API turns the #tags of the text into tags and its [[links]] into note ids: the sidebar, the related notes and the graph follow when they or the title change.
async function readTags(changes: Changes, note: ItemDetail) {
  if (changes.content === undefined) return;
  showResolvedLinks(changes.content, note.content ?? '');
  const changed =
    note.tags?.join() !== tags.value.join() ||
    linkedNotes(note.content ?? '') !== links ||
    note.item.title !== title.value;
  tags.value = note.tags ?? [];
  links = linkedNotes(note.content ?? '');
  title.value = note.item.title;
  if (changed) await refreshNuxtData();
}

// The API points the typed [[links]] to their note: the editor shows them resolved, without touching the cursor or the undo history.
function showResolvedLinks(sent: string, saved: string) {
  const editor = editorRef.value?.editor;
  const links = resolvedLinks(sent, saved);
  if (!editor || !links.size) return;
  const { tr } = editor.state;
  editor.state.doc.descendants((node, pos) => {
    const link = node.type.name === 'wikiLink' && !node.attrs.id && links.get(node.attrs.target.toLowerCase());
    if (link) tr.setNodeMarkup(pos, undefined, link);
  });
  if (tr.docChanged) editor.view.dispatch(tr.setMeta('addToHistory', false));
}

async function load() {
  if (!props.noteId) return;
  try {
    const note = await api<ItemDetail>(`/api/notes/${props.noteId}`, { quiet: true });
    initialContent.value = note.content ?? '';
    tags.value = note.tags ?? [];
    links = linkedNotes(initialContent.value);
    title.value = note.item.title;
    readonly.value = !!note.readonly;
    remember(note);
    place(note);
    isEmpty.value = false;
  } catch {
    notFound.value = true;
  } finally {
    loading.value = false;
  }
}
load();

function place(note: ItemDetail) {
  path.value = note.path.slice(0, note.path.lastIndexOf('/')) || '/';
  parent.value = note.item.parent_id ?? undefined;
}
// The note moved from the Cmd K palette: its path follows when the views read again, without holding them up, and its version when its text is the one saved.
const unhook = useNuxtApp().hook('app:data:refresh', () => {
  if (!id.value) return;
  api<ItemDetail>(`/api/notes/${id.value}`, { quiet: true }).then(
    (note) => {
      place(note);
      if (note.content === saved) remember(note);
    },
    () => {},
  );
});
onBeforeUnmount(unhook);

// The editor writes blank lines as non-breaking spaces, and & < > as HTML entities.
// Spaces first: a typed "&nbsp;" is written "&amp;nbsp;" and must come back as typed.
function clean(markdown: string) {
  return decodeEntities(markdown.replace(/&nbsp;/gi, ' ').replace(/\u00a0/g, ' ')).trim();
}

function save(content: string) {
  isEmpty.value = !content;
  saver.push({ content: content || undefined });
}

function onContent(markdown: string) {
  if (touched.value && !readonly.value) save(clean(markdown));
}

// What is typed before goes first: the API refuses it once locked.
async function toggleLock() {
  await saver.flush();
  const note = await api<ItemDetail>(`/api/notes/${id.value}`, {
    method: 'PATCH',
    body: { readonly: !readonly.value },
  });
  readonly.value = !!note.readonly;
}
// false: no update event, unlocking must not save the note again.
watch([() => editorRef.value?.editor, readonly], ([editor, locked]) => editor?.setEditable(!locked, false));

// Pro: the markdown as text, in place of the editor; undefined in the editor. Back in the editor, it starts from that text.
const source = ref<string>();
function toggleSource() {
  if (source.value !== undefined) {
    initialContent.value = source.value;
    source.value = undefined;
    return;
  }
  const markdown = editorRef.value?.editor?.getMarkdown();
  source.value = markdown === undefined ? initialContent.value : clean(markdown);
}
function onSource(text: string) {
  source.value = text;
  if (has('wordCount')) words.value = countWords(text);
  save(text.trim());
}

// The Cmd K palette renames the open note here, so the editor follows: its first heading takes the title, or one is added on top.
// TODO: the marks and #tags of that heading are replaced too, keep them if renamed headings lose tags.
const renaming = useState<(() => Promise<void>) | undefined>('rename-note');
async function rename() {
  const next = (
    await prompt({ title: t('items.renameNote'), value: displayTitle(title.value, ''), action: t('common.rename') })
  )?.trim();
  if (!next || next === title.value) return;
  if (readonly.value) return void toast.add({ title: t('note.locked'), color: 'warning' });
  touched.value = true;
  if (source.value !== undefined) return onSource(retitle(source.value, next));
  const editor = editorRef.value?.editor;
  if (!editor) return;
  const { tr, doc, schema } = editor.state;
  const first = outline.value[0];
  const node = first && doc.nodeAt(first.pos);
  if (first && node) tr.insertText(next, first.pos + 1, first.pos + node.nodeSize - 1);
  else tr.insert(0, schema.nodes.heading!.create({ level: 1 }, schema.text(next)));
  editor.view.dispatch(tr);
}
renaming.value = rename;
onBeforeUnmount(() => renaming.value === rename && (renaming.value = undefined));

// TODO: a keepalive body is limited to 64 KiB, a longer note loses its last 1.5 s when the tab closes; beacon has the same limit.
function saveOnClose() {
  if (id.value && isEmpty.value) {
    saver.take();
    fetch(`${baseURL}/api/notes/${id.value}`, { method: 'DELETE', keepalive: true });
    return;
  }
  if (saver.isBusy() && !id.value) return; // The note is being created: a second POST would duplicate it.
  const changes = saver.take();
  if (!changes || (!id.value && !changes.content)) return;

  const url = `${baseURL}/api/notes${id.value ? `/${id.value}` : ''}`;
  const body = JSON.stringify(id.value ? { ...changes, updated_at: version } : { ...changes, parent_id: parent.value });
  fetch(url, {
    method: id.value ? 'PATCH' : 'POST',
    body,
    keepalive: true,
    headers: { 'Content-Type': 'application/json' },
  });
}
useEventListener(window, 'pagehide', saveOnClose);

async function beforeLeave(to: RouteLocationNormalized) {
  if (to.params.id === id.value) return true; // The new note getting its address.
  if (isEmpty.value) {
    saver.take();
    return true;
  }
  await saver.flush();
  if (!failed.value) return true;
  const leave = await confirm({
    title: t('note.unsaved'),
    description: t('note.unsavedDescription'),
    action: t('note.leaveAnyway'),
  });
  return leave === true;
}
onBeforeRouteLeave(beforeLeave);
onBeforeRouteUpdate(beforeLeave);

// A note left empty goes to the trash, with its last saved content. Not while typing: clearing it to rewrite it must not close it.
onBeforeUnmount(async () => {
  if (!id.value || !isEmpty.value) return;
  await saver.flush(); // A save in progress would fail on a removed note.
  await removeItem({ id: id.value, kind: 'note' });
});

// A dropped image takes no focus: the upload is an interaction, its image an edit.
async function upload(file: File) {
  touched.value = true;
  try {
    const saved = await api<NoteFile>('/api/files', {
      method: 'POST',
      body: file,
      query: { name: file.name },
      quiet: true,
    });
    return `/api/files/${saved.id}`;
  } catch (error) {
    const status = (error as FetchError).status;
    const key = status === 413 ? 'note.imageTooLarge' : status === 415 ? 'note.imageType' : 'note.imageFailed';
    toast.add({ title: t(key, { name: file.name }), color: 'error' });
    return null;
  }
}

// Merged once the account is known: UEditor reads its extensions when created, so it waits for ready.
const modules = [
  useNoteFormatting(),
  useNoteLinks(id),
  useNoteSuggestions(),
  useNoteSlash(),
  ...useNoteTables(),
  useNoteCode(),
  ...useNoteCallouts(),
  { extensions: [noteImage(baseURL, upload), slashTrigger(openSlash)] },
];
const parts = computed(() => mergeModules(modules, has));

// The / menu, its items of a same group (headings, lists, callouts) in a submenu; opened where the / was typed, the / goes away with the pick.
const slashAt = ref<{ range: Range; left: number; top: number } | null>(null);
const slashMenu = computed(() => {
  const editor = editorRef.value?.editor;
  const slash = parts.value.menus.find((menu) => menu.pluginKey === 'slashMenu');
  if (!editor || !slash) return [];
  const handlers = { ...createHandlers(), ...parts.value.handlers };
  const items: DropdownMenuItem[] = [];
  for (const item of toValue(slash.items) as (EditorSuggestionMenuItem & { kind: string; group?: string })[]) {
    const entry = {
      label: item.label,
      icon: item.icon,
      onSelect: () => {
        if (slashAt.value) editor.chain().focus().deleteRange(slashAt.value.range).run();
        handlers[item.kind]?.execute(editor, item).run();
      },
    };
    if (!item.group) items.push(entry);
    else {
      const group = items.find((other) => other.label === item.group && other.children);
      if (group) (group.children as DropdownMenuItem[]).push(entry);
      else items.push({ label: item.group, icon: item.icon, children: [entry] });
    }
  }
  return items;
});
function openSlash(range: Range, rect: DOMRect | null) {
  if (rect && slashMenu.value.length) slashAt.value = { range, left: rect.left, top: rect.bottom };
}
const slashOpen = computed({
  get: () => !!slashAt.value,
  set: (open) => {
    if (open) return;
    slashAt.value = null;
    editorRef.value?.editor?.commands.focus();
  },
});
// The menu has the focus: a backspace there takes back the / and closes it, a second / is written (//), as in the text.
function slashKey(event: KeyboardEvent) {
  const editor = editorRef.value?.editor;
  if (!editor || !slashAt.value || (event.key !== 'Backspace' && event.key !== '/')) return;
  // chain().focus() focuses the editor as soon as it is built, run or not: only for these two keys.
  const chain = editor.chain().focus();
  if (event.key === 'Backspace') chain.deleteRange(slashAt.value.range).run();
  else chain.insertContent('/').run();
  event.preventDefault();
  slashAt.value = null;
}
useEventListener(window, 'keydown', slashKey);
// The same menu on a right click or a long press.
const contextMenu = computed(() => (has('contextMenu') && !readonly.value ? slashMenu.value : []));
// The menu keeps the browser from moving the cursor on a right click: it goes where the click was, unless inside the selected text.
function placeCursor(event: MouseEvent) {
  const editor = editorRef.value?.editor;
  const pos = editor?.view.posAtCoords({ left: event.clientX, top: event.clientY })?.pos;
  if (!editor || pos === undefined || !contextMenu.value.length) return;
  const { from, to } = editor.state.selection;
  if (pos < from || pos > to) editor.commands.setTextSelection(pos);
}

// The headings of the note, for the context panel: a click scrolls the text to one, without moving the cursor.
const panel = useContextPanel();
const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
const nearGraph = computed(() => (graph.value && id.value ? neighborhood(graph.value, id.value) : undefined));
const outline = ref<{ level: number; text: string; pos: number }[]>([]);
function readOutline(editor: Editor) {
  const headings: typeof outline.value = [];
  editor.state.doc.forEach((node, pos) => {
    if (node.type.name === 'heading' && node.textContent)
      headings.push({ level: node.attrs.level, text: node.textContent, pos });
  });
  outline.value = headings;
}
const words = ref(0);
function readText(editor: Editor) {
  readOutline(editor);
  if (has('wordCount')) words.value = countWords(editor.getText());
}
watch(
  () => editorRef.value?.editor,
  (editor) => {
    if (!editor) return;
    readText(editor);
    editor.on('update', () => readText(editor));
  },
);
function scrollTo(pos: number) {
  (editorRef.value?.editor?.view.nodeDOM(pos) as HTMLElement | null)?.scrollIntoView({
    behavior: 'smooth',
    block: 'start',
  });
}

const status = computed(() => {
  if (failed.value) return t('note.notSaved');
  if (isEmpty.value && id.value) return t('note.emptyWillTrash');
  if (saving.value) return t('note.saving');
  return '';
});
</script>

<template>
  <UEmpty
    v-if="notFound"
    :icon="icons.noteNotFound"
    :title="$t('note.notFound')"
    :description="$t('note.notFoundDescription')"
    :actions="[{ label: $t('common.backToNotes'), icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
  />

  <!-- UEditor hands its listeners to TipTap, not to the page: clicks on links are caught here.
       Only the text scrolls: a sticky toolbar inside the scroll moves with Safari's elastic bounce. -->
  <article
    v-else-if="!loading && ready"
    class="flex min-h-0 flex-1 flex-col pt-4"
    @focusin="touched = true"
    @pointerdown.capture="touched = true"
    @click="parts.onClick"
  >
    <!-- A div takes the menu's listeners: UEditor hands its own to TipTap. -->
    <UTextarea
      v-if="source !== undefined"
      :model-value="source"
      variant="none"
      autofocus
      :readonly="readonly"
      :aria-label="$t('note.source')"
      class="flex min-h-0 flex-1"
      :ui="{ base: 'h-full resize-none px-8 pb-8 font-mono text-sm' }"
      @update:model-value="onSource"
    />
    <UContextMenu v-else :items="contextMenu" :disabled="!contextMenu.length">
      <div class="flex min-h-0 flex-1 flex-col" @contextmenu.capture="placeCursor">
        <UEditor
          ref="editor"
          v-slot="{ editor }"
          :model-value="initialContent"
          content-type="markdown"
          :placeholder="$t('note.placeholder')"
          :autofocus="!noteId"
          :image="false"
          :mention="false"
          :starter-kit="parts.starterKit"
          :extensions="parts.extensions"
          :handlers="parts.handlers"
          class="flex min-h-0 w-full flex-1 flex-col"
          :ui="{ content: 'min-h-0 overflow-y-auto px-8 pb-8', base: 'sm:px-0 px-0' }"
          @update:model-value="onContent"
        >
          <!-- Its commands write even in a read-only editor: hidden while locked. -->
          <UEditorToolbar
            v-if="!readonly"
            :editor="editor"
            :items="parts.toolbar"
            class="mx-7 shrink-0 border-b border-muted py-1"
          />
          <UEditorToolbar
            v-for="bubble in readonly ? [] : parts.bubbles"
            :key="bubble.pluginKey"
            :editor="editor"
            :items="bubble.items"
            layout="bubble"
            :plugin-key="bubble.pluginKey"
            :options="{ placement: 'top-start' }"
            :should-show="bubble.shouldShow"
            class="rounded-md border border-default bg-default p-1 shadow-sm"
          />
          <UEditorSuggestionMenu
            v-for="{ items, ...menu } in parts.menus.filter((menu) => menu.pluginKey !== 'slashMenu')"
            :key="menu.pluginKey"
            :editor="editor"
            v-bind="menu"
            :items="toValue(items)"
          />
        </UEditor>
      </div>
    </UContextMenu>
    <UDropdownMenu
      v-model:open="slashOpen"
      :items="slashMenu"
      :content="{ align: 'start', onCloseAutoFocus: (event: Event) => event.preventDefault() }"
    >
      <span class="fixed size-0" :style="{ left: `${slashAt?.left}px`, top: `${slashAt?.top}px` }" />
    </UDropdownMenu>

    <footer class="shrink-0 border-t border-default px-8">
      <RelatedNotes v-if="id && !panel" :key="id" :note-id="id" />
      <div class="flex h-9 items-center gap-2 text-xs text-muted">
        <PathNav :path="path" />
        <span class="ml-auto shrink-0" :class="{ 'text-error': failed }" role="status">{{ status }}</span>
        <!-- The word count and the outline belong to the editor: hidden in the source mode. Dropping `source === undefined` shows the count, which follows the text; the outline would also need its headings read from the text. -->
        <span v-if="has('wordCount') && source === undefined" class="shrink-0 tabular-nums">{{
          $t('note.words', words)
        }}</span>
        <UTooltip v-if="id && has('noteLock')" :text="readonly ? $t('note.unlock') : $t('note.lock')">
          <UButton
            :icon="readonly ? icons.lock : icons.unlock"
            color="neutral"
            variant="ghost"
            size="xs"
            :aria-label="$t('note.lock')"
            :aria-pressed="readonly"
            @click="toggleLock"
          />
        </UTooltip>
        <UTooltip v-if="has('sourceMode')" :text="source === undefined ? $t('note.source') : $t('note.editor')">
          <UButton
            :icon="icons.source"
            color="neutral"
            variant="ghost"
            size="xs"
            :aria-label="$t('note.source')"
            :aria-pressed="source !== undefined"
            @click="toggleSource"
          />
        </UTooltip>
        <UTooltip v-if="has('focusMode')" :text="focus ? $t('note.unfocus') : $t('note.focus')" :kbds="['meta', '.']">
          <UButton
            :icon="focus ? icons.unfocus : icons.focus"
            color="neutral"
            variant="ghost"
            size="xs"
            :aria-label="focus ? $t('note.unfocus') : $t('note.focus')"
            :aria-pressed="focus"
            @click="focus = !focus"
          />
        </UTooltip>
      </div>
    </footer>

    <!-- Drawn by the layout beside the note, like the sidebar. -->
    <Teleport v-if="panel" to="#context-panel" defer>
      <nav v-if="source === undefined" :aria-label="$t('note.outline')">
        <h2 class="mb-2 text-xs font-medium text-muted">{{ $t('note.outline') }}</h2>
        <p v-if="!outline.length" class="text-xs text-dimmed">{{ $t('note.noOutline') }}</p>
        <button
          v-for="heading in outline"
          :key="heading.pos"
          type="button"
          class="block w-full truncate rounded py-1 text-left text-sm text-muted hover:text-default"
          :style="{ paddingLeft: `${(heading.level - 1) * 0.75}rem` }"
          @click="scrollTo(heading.pos)"
        >
          {{ heading.text }}
        </button>
      </nav>
      <RelatedNotes v-if="id" :key="id" :note-id="id" default-open />
      <!-- At the bottom of the panel, kept in sight while the rest scrolls. -->
      <section v-if="has('graph') && nearGraph?.links.length" class="sticky bottom-0 mt-auto -mb-4 bg-default pb-4">
        <div class="mb-2 flex items-center justify-between">
          <h2 class="text-xs font-medium text-muted">{{ $t('common.graph') }}</h2>
          <UButton
            :icon="icons.expand"
            :to="{ path: '/graph', query: { note: id } }"
            :aria-label="$t('related.showGraph')"
            color="neutral"
            variant="ghost"
            size="xs"
          />
        </div>
        <GraphView :graph="nearGraph" :selected="id" mini class="h-56!" />
      </section>
    </Teleport>
  </article>
</template>
