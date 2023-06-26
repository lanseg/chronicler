"use strict";

import { ChroniclerData } from "./wrappers.js";

export async function getRecordSets() {
    return fetch("/chronicler/")
        .then(response => response.text())
        .then(text => JSON.parse(text))
}

export async function getRecord(id) {
    return fetch(`/chronicler/${id}`)
        .then(response => response.text())
        .then(text => JSON.parse(text))
        .then(data => new ChroniclerData(data));
}
