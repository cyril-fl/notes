/** The context panel of a note: a Pro feature, beside the note on a wide screen, hidden by the focus mode. */
export function useContextPanel() {
  const { has } = usePro();
  const focus = useState('focus', () => false);
  const wide = useMediaQuery('(min-width: 1024px)');
  return computed(() => has('contextPanel') && !(focus.value && has('focusMode')) && wide.value);
}
