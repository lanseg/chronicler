"use strict";

export async function getRecordSets() {
    return fetch("/chronicler/records")
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
