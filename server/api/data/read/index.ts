export default defineEventHandler(
  async (): Promise<ServerResponse<DataSchema[]>> => {
    const mongo = useStorage('mongodb');

    const keys = await mongo.keys();
    const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;
    const now = Date.now();

    const data = await Promise.all(
      keys.map(async (key) => {
        const raw = await mongo.getItem<DataSchema>(key);
        const parsed = dataParams.safeParse(raw);
        if (!parsed.success) return null;

        const item = parsed.data;

        // Auto-purge expired soft-deleted items (> 30 days)
        if (item.deletedAt) {
          const deletedTime = new Date(item.deletedAt).getTime();
          if (now - deletedTime > THIRTY_DAYS_MS) {
            await mongo.removeItem(key);
            await removeFromIndex(item.id);
            return null;
          }
        }

        return item;
      })
    );

    const filtered = data.filter((d): d is DataSchema => d !== null);

    return { message: 'Notes fetched successfully.', data: filtered };
  }
);
