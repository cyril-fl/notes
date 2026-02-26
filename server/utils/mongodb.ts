import { MongoClient, type Collection, type Document } from 'mongodb';

let _collection: Collection<Document> | null = null;

export async function getMongoCollection(): Promise<Collection<Document>> {
  if (_collection) return _collection;

  const uri = process.env.MONGODB_URI || 'mongodb://localhost:27017/';
  const client = new MongoClient(uri);
  await client.connect();

  _collection = client.db('notes-pad').collection('items');
  return _collection;
}
