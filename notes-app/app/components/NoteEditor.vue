<script setup lang="ts">
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
const { confirm } = useDialogs();
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

const failed = ref(false);
const saving = ref(false);
const isEmpty = ref(!props.noteId);
// The editor can change its document while loading: only what follows a user interaction is an edit,
// or merely opening a note would rewrite the markdown written with the CLI. Pointerdown too: a task checkbox takes no focus.
const touched = ref(false);

const saver = createSaver(send);
const editorRef = useTemplateRef<{ editor?: Editor }>('editor');

async function send(changes: Changes): Promise<boolean | undefined> {
  if (id.value && Object.values(changes).every((value) => value === undefined)) return;
  // The API refuses a note with no content: a new one waits for its first text.
  if (!id.value && !changes.content) return false;

  saving.value = true;
  try {
    await (id.value ? patchNote(changes) : createNote(changes));
    failed.value = false;
  } catch (error) {
    failed.value = true;
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
  await readTags(changes, await api<ItemDetail>(`/api/notes/${item.id}`));
}

async function patchNote(changes: Changes) {
  await readTags(changes, await api<ItemDetail>(`/api/notes/${id.value}`, { method: 'PATCH', body: changes }));
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
    path.value = note.path.slice(0, note.path.lastIndexOf('/')) || '/';
    parent.value = note.item.parent_id ?? undefined;
    isEmpty.value = false;
  } catch {
    notFound.value = true;
  } finally {
    loading.value = false;
  }
}
load();

function onContent(markdown: string) {
  if (!touched.value) return;
  // The editor writes blank lines as non-breaking spaces, and & < > as HTML entities.
  // Spaces first: a typed "&nbsp;" is written "&amp;nbsp;" and must come back as typed.
  const content = decodeEntities(markdown.replace(/&nbsp;/gi, ' ').replace(/\u00a0/g, ' ')).trim();
  isEmpty.value = !content;
  saver.push({ content: content || undefined });
}

// shortcut: a keepalive body is limited to 64 KiB, a longer note loses its last 1.5 s when the tab closes; beacon has the same limit.
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
  const body = JSON.stringify(id.value ? changes : { ...changes, parent_id: parent.value });
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
  useNoteTables(),
  { extensions: [noteImage(baseURL, upload)] },
];
const parts = computed(() => mergeModules(modules, has));

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
watch(
  () => editorRef.value?.editor,
  (editor) => {
    if (!editor) return;
    readOutline(editor);
    editor.on('update', () => readOutline(editor));
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
      <UEditorToolbar :editor="editor" :items="parts.toolbar" class="mx-7 shrink-0 border-b border-muted py-1" />
      <UEditorToolbar
        v-for="bubble in parts.bubbles"
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
        v-for="{ items, ...menu } in parts.menus"
        :key="menu.pluginKey"
        :editor="editor"
        v-bind="menu"
        :items="toValue(items)"
      />
    </UEditor>

    <footer class="shrink-0 border-t border-default px-8">
      <RelatedNotes v-if="id && !panel" :key="id" :note-id="id" />
      <div class="flex h-9 items-center gap-2 text-xs text-muted">
        <PathNav :path="path" />
        <span class="ml-auto shrink-0" :class="{ 'text-error': failed }" role="status">{{ status }}</span>
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
      <nav :aria-label="$t('note.outline')">
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
        <h2 class="mb-2 text-xs font-medium text-muted">{{ $t('common.graph') }}</h2>
        <GraphView :graph="nearGraph" :selected="id" mini class="h-56!" />
      </section>
    </Teleport>
  </article>
</template>
