<template>
  <div class="flex h-dvh gap-4 overflow-hidden bg-muted p-4 text-default">
    <aside class="w-64 shrink-0 overflow-hidden rounded-2xl bg-default">
      <AppSidebar />
    </aside>
    <main class="min-w-0 flex-1 overflow-y-auto rounded-2xl bg-default">
      <slot />
    </main>
  </div>
</template>
