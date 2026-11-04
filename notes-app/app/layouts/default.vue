<template>
  <div class="flex h-dvh overflow-hidden bg-default text-default">
    <aside class="w-64 shrink-0 border-r border-default bg-muted">
      <AppSidebar />
    </aside>
    <main class="min-w-0 flex-1 overflow-y-auto">
      <slot />
    </main>
  </div>
</template>
