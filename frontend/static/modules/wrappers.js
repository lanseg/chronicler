"use strict;";

const imgExtensions = ["png", "jpg", "svg", "gif", "webp"];
const audioExtensions = ["wav", "mp3", "ogg", "oga"];
const videoExtensions = ["webm", "mp4", "mov"];
const docExtensions = ["pdf"];

function getExtension(path) {
    const lastIndex = path.lastIndexOf(".");
    return lastIndex > -1 && lastIndex < path.length ? path.substring(lastIndex + 1) : "";
}

function getFileName(path) {
    const nameStart = path.lastIndexOf("/") + 1;
    const nameEnd = path.indexOf("?", nameStart);
    return path.substring(nameStart, nameEnd == -1 ? path.length : nameEnd);
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
        return this._fileObj["file_url"] ?? this._fileObj["local_url"];
    }

    get name() {
        if (!this.fileUrl) {
            return "";
        }
        return getFileName(this.fileUrl);
    }

    get isAudio() {
        if (!this.fileUrl) {
            return false;
        }
        return audioExtensions.includes(getExtension(this.name.toLowerCase()));
    }

    get isVideo() {
        if (!this.fileUrl) {
            return false;
        }
        return videoExtensions.includes(getExtension(this.name.toLowerCase()));
    }

    get isImage() {
        if (!this.fileUrl) {
            return false;
        }
        return imgExtensions.includes(getExtension(this.name.toLowerCase()));
    }

    get isDocument() {
        if (!this.fileUrl) {
            return false;
        }
        return docExtensions.includes(getExtension(this.name.toLowerCase()));
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

    get id() {
        return (
            this.messageId ?? this.channelId ?? this.senderId ?? (this.url ? btoa(this.url) : null)
        );
    }
}

/** **/
export class SourceMetadata {
    constructor(metadataObj) {
        this._metadataObj = metadataObj;
    }

    get id() {
        return this._metadataObj["id"];
    }

    get name() {
        return this._metadataObj["username"] ?? this._metadataObj["id"];
    }

    get quotes() {
        return this._metadataObj["quotes"] ?? [];
    }
}

/** **/
export class Record {
    constructor(recordObj, source, parent) {
        this._recordObj = recordObj;

        this.time = recordObj["time"] ? new Date(recordObj["time"] * 1000) : null;
        this.source = source;
        this.parent = parent;

        this._allFiles = [];
        for (const file of recordObj["files"] ?? []) {
            this._allFiles.push(new File(file));
        }
    }

    get textContent() {
        return this._recordObj["text_content"] ?? "";
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
        this._recordById = new Map();
        this.sourceMetadata = new Map();
        this.records = [];

        for (let srcmd of recordSetObj["userMetadata"] ?? []) {
            const md = new SourceMetadata(srcmd);
            this.sourceMetadata.set(md.id, md);
        }

        for (const record of recordSetObj["records"] ?? []) {
            if (!record["source"]) {
                this.records.push(new Record(record, null));
                continue;
            }
            const source = new Source(record["source"]);
            const parent = record["parent"] ? new Source(record["parent"]) : null;
            const newRecord = new Record(record, source, parent);
            this.records.push(newRecord);
            this._recordById.set(source.id, newRecord);
        }

        for (const record of this.records) {
            if (!record.parentSource) {
                continue;
            }
            record.parent = this._recordById.get(record.parentSource.id);
            record.parentUser = this._userById.get(record.parentSource.id);
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
        this.sourceMetadata = new Map();
        this.recordSets = [];

        for (let srcmd of recordListObj["user_metadata"] ?? []) {
            const md = new SourceMetadata(srcmd);
            this.sourceMetadata.set(md.id, md);
        }

        for (const rs of recordListObj["record_sets"] ?? []) {
            if (rs["root_record"]) {
                const src = rs["root_record"]["source"]
                    ? new Source(rs["root_record"]["source"])
                    : null;
                const parent = rs["root_record"]["parent"]
                    ? new Source(rs["root_record"]["parent"])
                    : null;
                this.recordSets.push(
                    new RecordSetInfo(rs, new Record(rs["root_record"], src, parent)),
                );
            } else {
                this.recordSets.push(new RecordSetInfo(rs, null));
            }
        }
    }
}
