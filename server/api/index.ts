export default defineEventHandler(async () => {
  console.log('Hello World');

  // const mongo = useStorage('mongodb');
  // console.log(mongo);

  return {
    message: 'Hello World',
  };
});
