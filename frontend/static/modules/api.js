"use strict";

export async function getRecordSets(offset = 0, size = 10) {
    return fetch(`/chronicler/records?offset=${offset}&size=${size}`)
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}

export async function getRecord(id) {
    return fetch(`/chronicler/records/${id}`)
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}

export async function deleteRecordSets(ids) {
    return fetch(`chronicler/records/delete?ids=${ids.join(",")}`)
        .then((response) => response.text())
        .then((text) => JSON.parse(text));
}
