"use strict;"

const imgExtensions = ['png', 'jpg', 'svg', 'gif', 'webp'];

export function isImage(path) {
    return imgExtensions.includes(getExtension(path.toLowerCase()));
}

export function getFileName(path) {
    const lastIndex = path.lastIndexOf('/');
    return (lastIndex > -1 && lastIndex < path.length) ?
        path.substring(lastIndex + 1) : "";
}

export function getExtension(path) {
    const lastIndex = path.lastIndexOf('.');
    return (lastIndex > -1 && lastIndex < path.length) ?
        path.substring(lastIndex + 1) : "";
}

export const sourceType = ["unknown", "telegram", "twitter", "web", "youtube"];

export class User {
    constructor(userMetadata) {
        this._userMetadata = userMetadata;
    }

    get id() {
        return this._userMetadata["id"];
    }

    get name() {
        return this._userMetadata["username"] ?? this._userMetadata["id"];
    }

    get quotes() {
        return this._userMetadata["quotes"] ?? [];
    }
}

export class Record {
    constructor(record, user) {
        this.record = record;
        this._images = [];
        this._files = [];
        this._user = user;

        for (const file of record.files ?? []) {
            const fname = getFileName(file.file_url);
            if (isImage(fname)) {
                this._images.push(fname);
            } else {
                this._files.push(fname);
            }
        }
    }

    get user() {
        return this._user;
    }

    get name() {
        return this.user ? this.user.username : this.source.sender_id;
    }

    get images() {
        return this._images;
    }

    get files() {
        return this._files;
    }

    get time() {
        if (!this.record.time) {
            return undefined;
        }
        return this.record.time;
    }
}

export class ChroniclerData {

    users = new Map();

    constructor(data) {
        this.data = data;
        this._records = [];
        this.recordById = new Map();

        for (const user of data.userMetadata ?? []) {
            this.users.set(user.id, new User(user));
        }

        for (const record of data.records ?? []) {
            if (!record.source) {
                continue;
            }
            const recordObj = new Record(
                record,
                this.users.get(record.source.sender_id) ?? new User({ "id": record.source.sender_id })
            );
            this._records.push(recordObj);
            this.recordById.set(record.source.message_id, recordObj);
        }

        for (const recordObj of this.recordById.values()) {
            const parent = recordObj.record.parent;
            if (parent) {
                recordObj.parent = this.recordById.get(parent.message_id);
                continue;
            }

            const source = recordObj.record.source;
            if (source.channel_id) {
                recordObj.parent = this.recordById.get(source.channel_id);
            }

        }
    }

    get records() {
        return this._records;
    }

    getSourceName(record) {
        if (sourceType[record["source"]["type"]] === "web") {
            try {
                return new URL(record["source"]["url"]).host;
            } catch { }
            return rootRecord["source"]["url"];
        }
        if (userById.get(record["source"]["sender_id"])) {
            return userById.get(record["source"]["sender_id"])["username"];
        }
        return record["source"]["sender_id"];
    }
}
