export default defineEventHandler(
  async (event): Promise<ServerResponse<DataSchema[]>> => {
    const raw = getQuery(event);
    const parsed = searchQuerySchema.safeParse(raw);

    guard(parsed.data, 'Invalid search query.', 400);

    const results = await search(parsed.data.q, {
      type: parsed.data.type,
      tags: parsed.data.tags?.split(','),
    });

    return { message: 'Search results.', data: results };
  }
);
