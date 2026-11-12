<script setup lang="ts">
import type { DropdownMenuItem, EditorSuggestionMenuItem, EditorToolbarItem } from '@nuxt/ui';
import { createHandlers } from '@nuxt/ui/runtime/utils/editor.js';
import type { Range } from '@tiptap/core';
import type { Editor } from '@tiptap/vue-3';
import type { FetchError } from 'ofetch';
import type { NoteFile } from '~/types/items';

// The editor itself, its modules, toolbar and menus: the note page and the snippet dialog write in the same one.
// modelValue: the markdown it starts from, read again only when it changes. noteId: the note written, left out of the [[ menu.
// compact: without the margins of a full page. label: names the text for a screen reader, the note page has its title.
const props = defineProps<{
  modelValue: string;
  noteId?: string;
  readonly?: boolean;
  placeholder?: string;
  autofocus?: boolean;
  compact?: boolean;
  label?: string;
}>();
// touch: a change no focus announced, the note page counts it as typing.
const emit = defineEmits<{ 'update:modelValue': [markdown: string]; touch: [] }>();

const { icons } = useIcons();
const { has, ready } = usePro();
const { t } = useI18n();
const api = useApi();
const { confirm } = useDialogs();
const baseURL = useRuntimeConfig().public.apiBase;
const toast = useToast();

const editorRef = useTemplateRef<{ editor?: Editor }>('editor');
defineExpose({ editor: computed(() => editorRef.value?.editor) });

// The note scrolls in its own box, which the menus over the text do not watch (only the window): they follow it here.
function placeBubbles() {
  const editor = editorRef.value?.editor;
  const shown = editor ? parts.value.bubbles.filter((bubble) => bubble.shouldShow({ editor })) : [];
  if (!editor || !shown.length) return;
  const { tr } = editor.state;
  for (const bubble of shown) tr.setMeta(bubble.pluginKey, 'updatePosition');
  editor.view.dispatch(tr);
}

// false: no update event, unlocking must not save the note again.
watch([() => editorRef.value?.editor, () => props.readonly], ([editor, locked]) => editor?.setEditable(!locked, false));

// A dropped image takes no focus: the upload is an interaction, its image an edit.
async function upload(file: File) {
  emit('touch');
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
  useNoteLinks(toRef(() => props.noteId)),
  useNoteSuggestions(),
  useNoteSlash(),
  ...useNoteTables(),
  useNoteCode(),
  useNoteMath(),
  ...useNoteCallouts(),
  useNoteImages(baseURL, upload),
  { extensions: [slashTrigger(openSlash)] },
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
// Not on a touch screen: a long press sends contextmenu, the menu would take the place of the native copy and paste one.
const touch = useMediaQuery('(pointer: coarse)');
const contextMenu = computed(() => (has('contextMenu') && !props.readonly && !touch.value ? slashMenu.value : []));
// The menu keeps the browser from moving the cursor on a right click: it goes where the click was, unless inside the selected text.
function placeCursor(event: MouseEvent) {
  const editor = editorRef.value?.editor;
  const pos = editor?.view.posAtCoords({ left: event.clientX, top: event.clientY })?.pos;
  if (!editor || pos === undefined || !contextMenu.value.length) return;
  const { from, to } = editor.state.selection;
  if (pos < from || pos > to) editor.commands.setTextSelection(pos);
}
// No Ctrl+A nor Ctrl+Z on a touch screen: buttons undo and redo, select the whole note, or empty it like the AC of a calculator.
async function clearNote() {
  if (await confirm({ title: t('note.clear'), description: t('note.clearDescription'), action: t('note.clear') }))
    editorRef.value?.editor?.chain().focus().clearContent(true).run();
}
// On a phone each group of formatting buttons (headings, marks, blocks) folds into one menu, showing the active one's icon.
const phone = useMediaQuery('(max-width: 639px)');
const formatting = computed((): EditorToolbarItem[][] =>
  phone.value
    ? parts.value.toolbar.map((group) => {
        // A group of one (the import button) is already a button: a menu holding it alone would need two taps.
        if (group.length < 2) return group;
        const label = group.map((item) => item.tooltip?.text).join(', ');
        return [
          {
            icon: group[0]?.icon,
            tooltip: { text: label },
            items: group.map((item) => ({ ...item, label: item.tooltip?.text })),
          },
        ];
      })
    : parts.value.toolbar,
);
const toolbar = computed(() =>
  touch.value
    ? [
        [
          { kind: 'undo', icon: icons.undo, tooltip: { text: t('note.undo') }, 'aria-label': t('note.undo') },
          { kind: 'redo', icon: icons.redo, tooltip: { text: t('note.redo') }, 'aria-label': t('note.redo') },
        ],
        ...formatting.value,
        [
          {
            icon: icons.selectAll,
            tooltip: { text: t('note.selectAll') },
            'aria-label': t('note.selectAll'),
            onClick: () => editorRef.value?.editor?.chain().focus().selectAll().run(),
          },
          { icon: icons.clear, tooltip: { text: t('note.clear') }, 'aria-label': t('note.clear'), onClick: clearNote },
        ],
      ]
    : formatting.value,
);
</script>

<template>
  <!-- UEditor hands its listeners to TipTap, not to the page: clicks on links are caught here. -->
  <div v-if="ready" class="flex min-h-0 flex-1 flex-col" @click="parts.onClick">
    <!-- A div takes the menu's listeners: UEditor hands its own to TipTap. -->
    <UContextMenu :items="contextMenu" :disabled="!contextMenu.length">
      <div class="flex min-h-0 flex-1 flex-col" @contextmenu.capture="placeCursor" @scroll.capture="placeBubbles">
        <UEditor
          ref="editor"
          v-slot="{ editor }"
          v-bind="label ? { 'aria-label': label } : {}"
          :model-value="modelValue"
          content-type="markdown"
          :placeholder="placeholder"
          :autofocus="autofocus"
          :image="false"
          :mention="false"
          :starter-kit="parts.starterKit"
          :extensions="parts.extensions"
          :handlers="parts.handlers"
          class="flex min-h-0 w-full flex-1 flex-col"
          :ui="{
            content: compact ? 'min-h-0 overflow-y-auto' : 'min-h-0 overflow-y-auto px-8 pt-1 pb-8',
            // Compact: the text fills the box, so a click anywhere in it writes.
            base: compact ? 'min-h-full px-3 py-2' : 'sm:px-0 px-0',
          }"
          @update:model-value="emit('update:modelValue', cleanMarkdown($event))"
        >
          <!-- Its commands write even in a read-only editor: hidden while locked. -->
          <UEditorToolbar
            v-if="!readonly"
            :editor="editor"
            :items="toolbar"
            class="shrink-0 overflow-x-auto border-b border-muted py-1"
            :class="compact ? '' : 'mx-7'"
          />
          <UEditorToolbar
            v-for="bubble in readonly ? [] : parts.bubbles"
            :key="bubble.pluginKey"
            :editor="editor"
            :items="bubble.items"
            layout="bubble"
            :plugin-key="bubble.pluginKey"
            :options="{
              placement: 'top-start',
              // Above, or below when it fits only there; along a tall image, at the top of what shows of it, gone with it.
              flip: { fallbackStrategy: 'initialPlacement' },
              // 16: tiptap moves it by its offset (8) after the shift, which would leave it against the edge.
              shift: { crossAxis: true, padding: 16 },
              // Stuck over the image (pushed down by the shift), it also moves off its left edge.
              offset: ({ y, placement, rects }) => ({
                mainAxis: 8,
                crossAxis: placement.startsWith('top') && y > rects.reference.y - rects.floating.height ? 8 : 0,
              }),
              hide: true,
            }"
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
  </div>
</template>
