<script setup lang="ts">
import type { EditorToolbarItem } from '@nuxt/ui';
import type { Editor } from '@tiptap/vue-3';
import { TaskItem, TaskList } from '@tiptap/extension-list';
import type { RouteLocationNormalized } from 'vue-router';
import type { Item, ItemDetail, NotesGraph } from '~/types/items';
import type { Changes } from '~/utils/saver';

const { icons } = useIcons();

const props = defineProps<{ noteId?: string; parentId?: string }>();
const emit = defineEmits<{ created: [id: string] }>();

const ALLOWED_LINKS = ['http', 'https', 'mailto'];

const api = useApi();
const { prompt, confirm } = useDialogs();
const { removeItem } = useItems();
const baseURL = useRuntimeConfig().public.apiBase;
const toast = useToast();

const id = ref(props.noteId);
const parent = ref(props.parentId);
const path = ref('');
const tags = ref<string[]>([]);
const initialContent = ref('');
const loading = ref(!!props.noteId);
const notFound = ref(false);

const failed = ref(false);
const saving = ref(false);
const isEmpty = ref(!props.noteId);
// The editor can change its document while loading: only what follows a user interaction is an edit,
// or merely opening a note would rewrite the markdown written with the CLI. Pointerdown too: a task checkbox takes no focus.
const touched = ref(false);

const saver = createSaver(send);

async function send(changes: Changes): Promise<boolean | void> {
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
  await readTags(changes);
}

async function patchNote(changes: Changes) {
  await api(`/api/notes/${id.value}`, { method: 'PATCH', body: changes });
  await readTags(changes);
}

// The API turns the #tags of the text into tags: the sidebar and the related notes follow when they change.
async function readTags(changes: Changes) {
  if (changes.content === undefined) return;
  const note = await api<ItemDetail>(`/api/notes/${id.value}`);
  const changed = note.tags?.join() !== tags.value.join();
  tags.value = note.tags ?? [];
  if (changed) await refreshNuxtData();
}

async function load() {
  if (!props.noteId) return;
  try {
    const note = await api<ItemDetail>(`/api/notes/${props.noteId}`, { quiet: true });
    initialContent.value = note.content ?? '';
    tags.value = note.tags ?? [];
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
  const content = decodeEntities(markdown.replace(/&nbsp;/gi, ' ').replace(/ /g, ' ')).trim();
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
  fetch(url, { method: id.value ? 'PATCH' : 'POST', body, keepalive: true, headers: { 'Content-Type': 'application/json' } });
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
    title: "La note n'est pas enregistrée",
    description: 'Tes dernières modifications seront perdues.',
    action: 'Quitter quand même',
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

async function askLink(editor: Editor) {
  const current = editor.getAttributes('link').href as string | undefined;
  if (current) return editor.chain().focus().unsetLink().run();

  const href = await prompt({ title: 'Ajouter un lien', placeholder: 'https://', action: 'Ajouter' });
  if (href) editor.chain().focus().setLink({ href: /^[a-z]+:/i.test(href) ? href : `https://${href}` }).run();
}

// The notes offered after [[: picked, the link points to the note's id right away.
const { data: graph } = useLazyAsyncData('graph', () => api<NotesGraph>('/api/graph/notes'));
const noteItems = computed(() =>
  (graph.value?.notes ?? [])
    .filter((note) => note.id !== id.value && note.title)
    .map((note) => ({ kind: 'wikiLink', label: note.title, description: note.path, icon: icons.note, id: note.id }))
);

function openLink(event: MouseEvent) {
  const link = (event.target as HTMLElement).closest<HTMLElement>('a[data-wikilink]');
  if (!link) return;
  event.preventDefault();
  if (link.dataset.note) return navigateTo(`/notes/${link.dataset.note}`);
  toast.add({
    title: 'Lien en attente',
    description: `Il sera relié dès qu'une seule note s'appellera « ${link.dataset.wikilink} ». Écris son chemin (/Dossier/Titre) si plusieurs portent ce titre.`,
    color: 'warning',
  });
}

const handlers = {
  wikiLink: {
    canExecute: () => true,
    execute: (editor: Editor, item: { label: string; id: string }) =>
      editor.chain().focus().insertContent([{ type: 'wikiLink', attrs: { target: item.label, id: item.id } }, { type: 'text', text: ' ' }]),
    isActive: () => false,
    isDisabled: () => false,
  },
  link: {
    canExecute: () => true,
    execute: (editor: Editor) => {
      askLink(editor);
      return editor.chain();
    },
    isActive: (editor: Editor) => editor.isActive('link'),
    isDisabled: (editor: Editor) => editor.state.selection.empty && !editor.isActive('link'),
  },
};

const toolbar: EditorToolbarItem[][] = [
  [
    { kind: 'heading', level: 1, icon: icons.heading1, tooltip: { text: 'Titre 1' } },
    { kind: 'heading', level: 2, icon: icons.heading2, tooltip: { text: 'Titre 2' } },
    { kind: 'heading', level: 3, icon: icons.heading3, tooltip: { text: 'Titre 3' } },
  ],
  [
    { kind: 'mark', mark: 'bold', icon: icons.bold, tooltip: { text: 'Gras' } },
    { kind: 'mark', mark: 'italic', icon: icons.italic, tooltip: { text: 'Italique' } },
    { kind: 'mark', mark: 'strike', icon: icons.strike, tooltip: { text: 'Barré' } },
    { kind: 'mark', mark: 'code', icon: icons.code, tooltip: { text: 'Code' } },
    { kind: 'link', icon: icons.link, tooltip: { text: 'Lien' } },
  ],
  [
    { kind: 'bulletList', icon: icons.list, tooltip: { text: 'Liste' } },
    { kind: 'orderedList', icon: icons.orderedList, tooltip: { text: 'Liste numérotée' } },
    { kind: 'taskList', icon: icons.taskList, tooltip: { text: 'Tâches' } },
    { kind: 'blockquote', icon: icons.quote, tooltip: { text: 'Citation' } },
  ],
];

const starterKit = {
  underline: false,
  link: {
    openOnClick: false,
    isAllowedUri: (url: string, ctx: { defaultValidate: (url: string) => boolean }) =>
      ALLOWED_LINKS.includes((url.split(':')[0] ?? '').toLowerCase()) && ctx.defaultValidate(url),
  },
} as object;
const extensions = [TaskList, TaskItem.configure({ nested: true }), WikiLink];

const status = computed(() => {
  if (failed.value) return 'Non enregistré';
  if (isEmpty.value && id.value) return 'Vide : ira à la corbeille en la quittant';
  if (saving.value) return 'Enregistrement…';
  return '';
});
</script>

<template>
  <UEmpty
    v-if="notFound"
    :icon="icons.noteNotFound"
    title="Cette note est introuvable"
    description="Elle n'existe pas ou elle est dans la corbeille."
    :actions="[{ label: 'Retour aux notes', icon: icons.home, color: 'neutral', variant: 'subtle', to: '/' }]"
  />

  <!-- UEditor hands its listeners to TipTap, not to the page: clicks on links are caught here. -->
  <article v-else-if="!loading" class="flex flex-1 flex-col p-8" @focusin="touched = true" @pointerdown.capture="touched = true" @click="openLink">
    <UEditor
      v-slot="{ editor }"
      :model-value="initialContent"
      content-type="markdown"
      placeholder="Commencez par un titre…"
      :autofocus="!noteId"
      :image="false"
      :mention="false"
      :starter-kit="starterKit"
      :extensions="extensions"
      :handlers="handlers"
      class="mt-2 w-full"
      :ui="{ base: 'sm:px-0 px-0' }"
      @update:model-value="onContent"
    >
      <UEditorToolbar
        :editor="editor"
        :items="toolbar"
        class="sticky top-0 z-10 -mx-1 border-b border-muted bg-default py-1"
      />
      <UEditorSuggestionMenu
        :editor="editor"
        :items="noteItems"
        char="[["
        plugin-key="wikiLinkMenu"
        :filter-fields="['label', 'description']"
        :suggestion="{ allowSpaces: true }"
      />
    </UEditor>

    <!-- mt-auto + sticky: pinned to the bottom of the page, short note or long. -->
    <footer class="sticky bottom-0 z-10 -mx-8 -mb-8 mt-auto border-t border-default bg-default px-8">
      <RelatedNotes v-if="id" :key="id" :note-id="id" />
      <div class="flex h-9 items-center gap-2 text-xs text-muted">
        <PathNav :path="path" />
        <span class="ml-auto shrink-0" :class="{ 'text-error': failed }" role="status">{{ status }}</span>
      </div>
    </footer>
  </article>
</template>
