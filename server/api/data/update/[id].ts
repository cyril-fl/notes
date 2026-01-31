import { defu } from 'defu';

export default defineEventHandler(
  async (event): Promise<ServerResponse<DraftData>> => {
    const id = getRouterParam(event, 'id');
    guard(id, 'ID required.', 400); // i18n

    const mongo = useStorage('mongodb');

    const existing = await mongo.getItem<DataSchema>(id);
    const parsedExisting = dataParams.safeParse(existing);
    guard(parsedExisting.data, 'Note not found.', 404);

    const body = await readBody<RequestBody<DraftData>>(event);

    const parsedBody = dataDraftPartial.safeParse(body);
    guard(parsedBody.data, 'Invalid data.', 400);

    const now = new Date();

    // TODO check defu  voir si les children sont bien merge
    const raw = defu(
      {
        ...parsedBody.data,
        lastEdited: now,
      },
      parsedExisting.data
    );

    const updated = dataParams.safeParse(raw);
    guard(updated.data, 'Invalid data.', 400);

    await mongo.setItem(id, updated.data);

    return { message: 'Successfully updated note.', data: updated.data };
  }
);
