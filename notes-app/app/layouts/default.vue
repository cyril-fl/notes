<template>
  <div class="flex h-dvh gap-2 overflow-hidden bg-muted p-2 text-default">
    <aside class="w-64 shrink-0 overflow-hidden rounded-xl bg-default">
      <AppSidebar />
    </aside>
    <main class="min-w-0 flex-1 overflow-y-auto rounded-xl bg-default">
      <slot />
    </main>
  </div>
</template>
