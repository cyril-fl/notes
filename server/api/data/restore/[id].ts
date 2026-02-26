export default defineEventHandler(
  async (event): Promise<ServerResponse<null>> => {
    const id = getRouterParam(event, 'id');

    guard(id, 'Item ID required.', 400);

    const ids = id.split(',');

    const mongo = useStorage('mongodb');

    await Promise.all(
      ids.map(async (id) => {
        const existing = await mongo.getItem<DataSchema>(id);

        const parsed = dataParams.safeParse(existing);

        guard(parsed.data, 'Item not found.', 404);

        const updated = { ...parsed.data, deletedAt: null };
        await mongo.setItem(id, updated);
        await addToIndex(updated);
      })
    );

    return { message: 'Item(s) restored.', data: null };
  }
);
