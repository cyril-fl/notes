export default defineEventHandler(
  async (event): Promise<ServerResponse<DataSchema>> => {
    const data = await readBody<RequestBody<DraftData>>(event);

    const parsed = dataDraft.safeParse(data);

    guard(parsed.data, 'Invalid data.', 400);

    const mongo = useStorage('mongodb');

    const now = new Date();

    console.log('[DATA API - CREATE] parsed', parsed);

    const id = generateId(parsed.data.type);

    const note: DataSchema = {
      id,
      createdAt: now,
      updatedAt: now,
      ...parsed.data,
    };

    console.log('[DATA API - CREATE] note', note);

    await mongo.setItem(id, note);

    setResponseStatus(event, 201);

    return { message: 'Note created successfully.', data: note };
  }
);
