<script setup lang="ts">
import type { Choice, MergePart } from '~/utils/merge';

// After a 409 with changes in the same place (#78): the merged text, each conflict at its place with the text typed here,
// the other device's and a choice. Closes with the text chosen; the text typed here is already kept in the history.
const props = defineProps<{ parts: MergePart[] }>();
const emit = defineEmits<{ close: [text: string] }>();

const phone = useMediaQuery('(max-width: 639px)');
const conflicts = props.parts.filter((part) => !('lines' in part));
const choices = ref<(Choice | undefined)[]>(conflicts.map(() => undefined));
const left = computed(() => choices.value.filter((choice) => !choice).length);
const options: { choice: Choice; label: string }[] = [
  { choice: 'mine', label: 'conflict.mine' },
  { choice: 'theirs', label: 'conflict.theirs' },
  { choice: 'both', label: 'conflict.both' },
];

// The index of each conflict among the parts, for its choice.
const indexOf = (part: MergePart) => conflicts.indexOf(part);
const save = () => emit('close', mergedText(props.parts, choices.value as Choice[]));
const keepTheirs = () =>
  emit(
    'close',
    mergedText(
      props.parts,
      conflicts.map(() => 'theirs'),
    ),
  );
</script>

<template>
  <UModal
    :title="$t('conflict.title', { count: left })"
    :description="$t('conflict.description')"
    :fullscreen="phone"
    :dismissible="false"
    :close="false"
    :ui="{ content: 'sm:max-w-3xl' }"
  >
    <template #body>
      <div class="flex flex-col gap-2 text-sm">
        <template v-for="(part, index) in parts" :key="index">
          <pre v-if="'lines' in part" class="font-sans whitespace-pre-wrap">{{ part.lines.join('\n') }}</pre>
          <section
            v-else
            class="flex flex-col gap-2 rounded-md border p-3"
            :class="choices[indexOf(part)] ? 'border-default' : 'border-warning'"
            :aria-label="$t('conflict.block', { number: indexOf(part) + 1 })"
          >
            <span class="text-xs text-muted">{{ $t('conflict.here') }}</span>
            <pre class="font-sans whitespace-pre-wrap">{{ part.mine.join('\n') }}</pre>
            <span class="text-xs text-muted">{{ $t('conflict.elsewhere') }}</span>
            <pre class="font-sans whitespace-pre-wrap">{{ part.theirs.join('\n') }}</pre>
            <div class="flex flex-wrap gap-2">
              <UButton
                v-for="option in options"
                :key="option.choice"
                :label="$t(option.label)"
                size="xs"
                :color="choices[indexOf(part)] === option.choice ? 'primary' : 'neutral'"
                :variant="choices[indexOf(part)] === option.choice ? 'solid' : 'outline'"
                :aria-pressed="choices[indexOf(part)] === option.choice"
                @click="choices[indexOf(part)] = option.choice"
              />
            </div>
          </section>
        </template>
      </div>
    </template>
    <template #footer>
      <div class="flex w-full flex-wrap justify-between gap-2">
        <UButton :label="$t('conflict.keepTheirs')" color="neutral" variant="ghost" @click="keepTheirs" />
        <UButton :label="$t('conflict.save')" :disabled="left > 0" @click="save" />
      </div>
    </template>
  </UModal>
</template>
