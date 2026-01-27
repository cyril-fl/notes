export default defineEventHandler(
  async (event): Promise<ServerResponse<DataSchema>> => {
    const id = getRouterParam(event, 'id');

    guard(id, 'ID required.', 400);

    const mongo = useStorage('mongodb');

    const data = await mongo.getItem<DataSchema>(id);

    guard(data, 'Note not found.', 404);

    return { message: 'Note fetched successfully.', data };
  }
);
