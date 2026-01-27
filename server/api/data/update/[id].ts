import { defu } from 'defu';

export default defineEventHandler(
  async (event): Promise<ServerResponse<DraftData>> => {
    const id = getRouterParam(event, 'id');
    guard(id, 'ID required.', 400); // i18n

    console.log('[DATA API - UPDATE] id', id);

    const mongo = useStorage('mongodb');

    const existing = await mongo.getItem<DataSchema>(id);
    const parsedExisting = dataParams.safeParse(existing);
    guard(parsedExisting.data, 'Note not found.', 404);

    console.log('[DATA API - UPDATE] parsedExisting', parsedExisting);
    const body = await readBody<RequestBody<DraftData>>(event);

    console.log('[DATA API - UPDATE] body', body);
    const parsedBody = dataDraftPartial.safeParse(body); // TODO fail ici
    guard(parsedBody.data, 'Invalid data.', 400);

    console.log('[DATA API - UPDATE] parsedBody', parsedBody);

    const now = new Date();

    const raw = defu(
      {
        ...parsedBody.data,
        lastEdited: now,
      },
      parsedExisting.data
    );

    console.log('[DATA API - UPDATE] raw', raw);

    const updated = dataParams.safeParse(raw);
    guard(updated.data, 'Invalid data.', 400);

    console.log('[DATA API - UPDATE] updated', updated);

    await mongo.setItem(id, updated.data);

    return { message: 'Successfully updated note.', data: updated.data };
  }
);
