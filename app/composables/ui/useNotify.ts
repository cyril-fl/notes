export function useNotify() {
  const { icons } = useIcons();

  function error(message: string) {
    const toast = useToast();
    toast.add({
      title: message,
      icon: icons.warning,
      color: 'error',
      duration: 5000,
    });
  }

  return { error };
}
