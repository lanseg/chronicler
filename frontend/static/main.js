"use strict";

/* */
const imgExtensions = ['png', 'jpg', 'svg', 'gif', 'webp'];
const sourceType = ["unknown", "telegram", "twitter", "web", "youtube"];

/* */
Map.prototype.getOrElse = function (key, value) {
    return this.has(key) ? this.get(key) : value;
}

function pad(number, len = 2) {
    return String(number).padStart(2, '0');
}

function formatDateTime(date) {
    return `${pad(date.getDate())}.${pad(date.getMonth() + 1)}.${date.getFullYear()} ` +
        `${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function isImage(path) {
    return imgExtensions.includes(getExtension(path.toLowerCase()));
}

function getExtension(path) {
    const lastIndex = path.lastIndexOf('.');
    return (lastIndex > -1 && lastIndex < path.length) ?
        path.substring(lastIndex + 1) : "";
}

function getFileName(path) {
    const lastIndex = path.lastIndexOf('/');
    return (lastIndex > -1 && lastIndex < path.length) ?
        path.substring(lastIndex + 1) : "";
}

function createElement(name, attributes) {
    const el = document.createElement(name);
    Object.entries(attributes ?? {}).map(obj => {
        el.setAttribute(obj[0], obj[1]);
    });
    return el;
}

/* */
async function getRecordSets() {
    return fetch("/chronicler/")
        .then(response => response.text())
        .then(text => JSON.parse(text))
}

async function getRecord(id) {
    return fetch(`/chronicler/${id}`)
        .then(response => response.text())
        .then(text => JSON.parse(text))
        .then(data => new ChroniclerData(data));
}

class Record {
    constructor(record) {
        this.record = record;
        this._images = [];
        this._files = [];

        for (const file of record.files ?? []) {
            const fname = getFileName(file.file_url);
            if (isImage(fname)) {
                this._images.push(fname);
            } else {
                this._files.push(fname);
            }
        }
    }

    get images() {
        return this._images;
    }

    get files() {
        return this._files;
    }

    get date() {
        return new Date(this.record.time * 1000);
    }
}

class ChroniclerData {

    constructor(data) {
        this.data = data;
        this._records = [];
        this.userById = new Map();
        this.recordById = new Map();

        for (const user of data.userMetadata ?? []) {
            this.userById.set(user.id, user);
        }

        for (const record of data.records) {
            const recordObj = new Record(record);
            this._records.push(recordObj);
            if (!record.source) {
                continue;
            }
            this.recordById.set(record.source.message_id, recordObj);
        }
    }

    get records() {
        return this._records;
    }

    getUserName(userId) {
        return this.userById.has(userId) ? this.userById.get(userId).username : userId;
    }

    getParent(record) {
        if (record.parent) {
            return this.recordById.get(record.parent.message_id);
        }
        if (record.source.channel_id) {
            return this.recordById.get(record.source.channel_id);
        }
        return undefined;
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
