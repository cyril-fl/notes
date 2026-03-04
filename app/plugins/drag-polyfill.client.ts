import { polyfill } from 'mobile-drag-drop';
import { scrollBehaviourDragImageTranslateOverride } from 'mobile-drag-drop/scroll-behaviour';

export default defineNuxtPlugin(() => {
  polyfill({
    dragImageTranslateOverride: scrollBehaviourDragImageTranslateOverride,
  });
});
