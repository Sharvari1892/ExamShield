import { openDB } from 'idb';

export const dbPromise = openDB('examshield', 1, {
  upgrade(db) {
    db.createObjectStore('session');
  },
});
