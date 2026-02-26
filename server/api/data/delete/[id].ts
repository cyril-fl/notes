export default defineEventHandler(
  async (event): Promise<ServerResponse<null>> => {
    const id = getRouterParam(event, 'id');

    guard(id, 'Note ID required.', 400);

    const ids = id.split(',');

    const mongo = useStorage('mongodb');

    await Promise.all(
      ids.map(async (id) => {
        const existing = await mongo.getItem<DataSchema>(id);

        const parsed = dataParams.safeParse(existing);

        guard(parsed.data, 'Note not found.', 404);

        await mongo.removeItem(id);
        await removeFromIndex(id);
      })
    );

    return { message: 'Note deleted successfully.', data: null };
  }
);
