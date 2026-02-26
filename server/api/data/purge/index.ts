export default defineEventHandler(async (): Promise<ServerResponse<null>> => {
  const mongo = useStorage('mongodb');

  const keys = await mongo.keys();

  await Promise.all(
    keys.map(async (key) => {
      const raw = await mongo.getItem<DataSchema>(key);
      const parsed = dataParams.safeParse(raw);
      if (!parsed.success) return;

      if (parsed.data.deletedAt) {
        await mongo.removeItem(key);
        await removeFromIndex(parsed.data.id);
      }
    })
  );

  return { message: 'Trash emptied.', data: null };
});
