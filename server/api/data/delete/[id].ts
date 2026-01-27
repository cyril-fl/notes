export default defineEventHandler(
  async (event): Promise<ServerResponse<null>> => {
    const id = getRouterParam(event, 'id');

    guard(id, 'Note ID required.', 400);

    const mongo = useStorage('mongodb');

    const existing = await mongo.getItem<DataSchema>(id);

    console.info('[DELETE] - Existing item to delete:', existing);

    const parsed = dataParams.safeParse(existing);

    guard(parsed.data, 'Note not found.', 404);

    await mongo.removeItem(id);

    return { message: 'Note deleted successfully.', data: null };
  }
);
