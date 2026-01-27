export default defineEventHandler(
  async (): Promise<ServerResponse<DataSchema[]>> => {
    const mongo = useStorage('mongodb');

    const keys = await mongo.keys();

    const data = await Promise.all(
      keys.map(async (key) => await mongo.getItem<DataSchema>(key))
    );

    const filtered = data.filter(
      (d): d is DataSchema => dataParams.safeParse(d).success
    );

    return { message: 'Notes fetched successfully.', data: filtered };
  }
);
