export default defineEventHandler(async (): Promise<ServerResponse<null>> => {
  const mongo = useStorage('mongodb');

  await mongo.clear();

  return { message: 'Notes flushed successfully.', data: null };
});
