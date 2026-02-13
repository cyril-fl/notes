export function useFolder() {
  const isEditing = useState('isEditing', () => false);

  /* TODO Créer des hook pour gerer le is edition, lajout de folder ext

  voir si des inject son necessaire mais je ne pense pas .

  utiliser un Editable de teka reka pour editer
  */
  return {
    isEditing,
  };
}
