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

/** **/
export class SourceType {

    static byId = new Map();

    constructor(id, name) {
        this.id = id;
        this.name = name;
        SourceType.byId.set(id, this);
    }

    static UNKNOWN = new SourceType(0, "unknown");
    static TELEGRAM = new SourceType(1, "telegram");
    static TWITTER = new SourceType(2, "twitter");
    static WEB = new SourceType(3, "web");
    static YOUTUBE = new SourceType(4, "youtube");
}

/** **/
export class File {
    constructor(fileObj) {
        this._fileObj = fileObj;
    }

    get fileId() {
        return this._fileObj["file_id"];
    }

    get fileUrl() {
        return this._fileObj["file_url"];
    }
}


/** **/
export class Source {

    constructor(sourceObj) {
        this._sourceObj = sourceObj;
    }

    get senderId() {
        return this._sourceObj["sender_id"];
    }

    get channelId() {
        return this._sourceObj["channel_id"];
    }

    get messageId() {
        return this._sourceObj["message_id"];
    }

    get url() {
        return this._sourceObj["url"];
    }

    get sourceType() {
        return SourceType.byId.get(this._sourceObj["type"] ?? 0);
    }
}


/** **/
export class UserMetadata {
    constructor(userObj) {
        this._userObj = userObj;
    }

    get id() {
        return this._userObj["id"];
    }

    get name() {
        return this._userObj["username"] ?? this._userObj["id"];
    }

    get quotes() {
        return this._userMetadata["quotes"] ?? [];
    }
}

/** **/
export class Record {

    constructor(recordObj, user) {
        this._recordObj = recordObj;

        this.time = recordObj["time"] ? new Date(recordObj["time"] * 1000) : null;
        this.user = user;
        if (recordObj["source"]) {
            this.source = new Source(recordObj["source"]);
        }
        if (recordObj["parent"]) {
            this.parent = new Source(recordObj["parent"]);
        }

        this._allFiles = [];
        for (const file of recordObj["files"] ?? []) {
            this._allFiles.push(new File(file));
        }
    }

    get textContent() {
        return this._recordObj["text_content"] ?? "";
    }

    get name() {
        return this.user ? this.user.username : this.source.sender_id;
    }

    get images() {
        return [];
    }

    get files() {
        return this._allFiles;
    }
}

/** **/
export class Request {
    constructor(requestObj) {
        this._requestObj = requestObj;

        this.source = new Source(requestObj["source"]);
    }
}

/** **/
export class RecordSet {
    constructor(recordSetObj) {
        this._recordSetObj = recordSetObj;
        this._userById = new Map();
        this.userMetadata = [];
        this.records = [];

        for (let user of recordSetObj["userMetadata"] ?? []) {
            const md = new UserMetadata(user);
            this._userById.set(md.id, md);
            this.userMetadata.push(md);
        }

        for (const record of recordSetObj["records"] ?? []) {
            this.records.push(new Record(record, record["source"] ? this._userById.get(record["source"]["sender_id"]) : null));
        }
    }

    get id() {
        return this._recordSetObj["id"];
    }
}

/** **/
export class RecordSetInfo {

    constructor(rsInfoObj, record) {
        this._rsInfoObj = rsInfoObj;
        this.rootRecord = record;
    }

    get id() {
        return this._rsInfoObj["id"];
    }

    get description() {
        return this._rsInfoObj["description"];
    }

    get recordCount() {
        return this._rsInfoObj["record_count"];
    }

}

export class RecordListResponse {
    constructor(recordListObj) {
        this._recordListObj = recordListObj;
        this._userById = new Map();
        this.userMetadata = [];
        this.recordSets = [];

        for (let user of recordListObj["user_metadata"] ?? []) {
            const md = new UserMetadata(user);
            this._userById.set(md.id, md);
            this.userMetadata.push(md);
        }

        for (const rs of recordListObj["record_sets"] ?? []) {
            if (rs["root_record"]) {
                const rootRecord = new Record(rs["root_record"],
                    rs["root_record"]["source"] ?
                        this._userById.get(rs["root_record"]["source"]["sender_id"]) : null);
                this.recordSets.push(new RecordSetInfo(rs, rootRecord));
            } else {
                this.recordSets.push(new RecordSetInfo(rs, null));
            }
        }
    }
}
