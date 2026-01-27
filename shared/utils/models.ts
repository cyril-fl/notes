import twitter from 'twitter-text';
import { ItemType } from '../types/models';
import type { FolderSchema, ItemSchema, NoteSchema } from '../types/models';
import type { CalendarDate } from '@internationalized/date';
import { today, getLocalTimeZone } from '@internationalized/date';

/* Models */
// Base Item
export abstract class Item {
  public readonly id: string;

  abstract title: string;
  public createdAt: CalendarDate;
  public updatedAt: CalendarDate;

  abstract readonly type: ItemType;

  private _path: string[];

  constructor(data: ItemSchema) {
    const now = today(getLocalTimeZone());

    this.id = data.id;
    this.createdAt = now;
    this.updatedAt = now;
    this._path = data.path;
  }

  // Getters
  public get ancestor(): string | null {
    return this._path.at(-1) ?? null;
  }

  public get ancestors(): string[] {
    return this._path;
  }

  public get path(): string[] {
    return [...this._path, this.id];
  }

  public set path(path: string[]) {
    this._path = path;
  }
}

// Folder Item
export class Folder extends Item {
  public readonly type = ItemType.FOLDER;

  public title: string;
  public childrenIds: string[];

  constructor(data: FolderSchema) {
    super(data);
    this.title = data.title;
    this.childrenIds = data.childrenIds;
  }
}

// Note Item
export class Note extends Item {
  public readonly type = ItemType.NOTE;

  public content: string;

  constructor(data: NoteSchema) {
    super(data);
    this.content = data.content;
  }

  // Getters
  public get title(): string {
    return extractTitleFromContent(this.content);
  }

  public get hashtags(): string[] {
    return twitter.extractHashtags(this.content);
  }

  public get mentions(): string[] {
    return twitter.extractMentions(this.content);
  }
}
