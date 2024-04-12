"use strict";

export async function getRecordSets(offset = 0, size = 10, query = "") {
    return fetch(
        `/chronicler/records?offset=${encodeURI(offset)}&size=${encodeURI(size)}&query=${encodeURI(query)}`,
    )
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}

export async function getRecord(id) {
    return fetch(`/chronicler/records/${encodeURI(id)}`)
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}

export async function deleteRecordSets(ids) {
    return fetch(`chronicler/records/delete?ids=${encodeURI(ids.join(","))}`)
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}
