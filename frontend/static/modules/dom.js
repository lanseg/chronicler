function pad(number) {
    return String(number).padStart(2, "0");
}

function formatDateTime(date) {
    if (!date) {
        return "No date";
    }
    return (
        `${pad(date.getDate())}.${pad(date.getMonth() + 1)}.${date.getFullYear()} ` +
        `${pad(date.getHours())}:${pad(date.getMinutes())}`
    );
}

function getSourceName(metadata, parentSrc, source) {
    if (!parentSrc && !source) {
        return "";
    }
    const nameOrId = formatSource(parentSrc ?? source);
    if (metadata != undefined && metadata.get(nameOrId)) {
        return metadata.get(nameOrId).name;
    }
    return nameOrId;
}

function formatSource(source) {
    if (source.sourceType.name === "web") {
        try {
            return new URL(source.url).host;
        } catch {}
        return source.url;
    }
    if (source.senderId) {
        return source.senderId;
    }
    return source.channelId;
}

// DOM manipulation
function createElement(name, attributes, content) {
    const el = document.createElement(name);
    Object.entries(attributes ?? {}).map((obj) => {
        el.setAttribute(obj[0], obj[1]);
    });
    if (typeof content === "string" || content instanceof String) {
        el.innerHTML = content;
    } else if (content && content.nodeType) {
        el.appendChild(content);
    }
    return el;
}

// Common elements
var _togglerCounter = 0;
function createExpander(header, content) {
    const togglerId = `_toggler${_togglerCounter}`;
    const element = createElement(
        "div",
        { class: "section" },
        `<input class="toggler" type="checkbox" id="${togglerId}" />     
        <div class="header">
        <label class="toggler" for="${togglerId}">
        <div class="title"><span class="toggler_status"></span></div>
      </label>
    </div>
    <div class="content"></div>`,
    );

    const head = element.querySelector("div.header .title");
    const body = element.querySelector("div.content");
    if (header) {
        head.appendChild(header);
    }
    if (content) {
        body.appendChild(content);
    }
    _togglerCounter++;
    return element;
}

// File elements
function createRecordSection(title) {
    return createExpander(
        createElement("span", {}, title),
        createElement("div", { class: "content" }),
    );
}

function createAudioElement(title, url) {
    return createElement(
        "figure",
        {},
        `
        <figcaption>${title}</figcaption>
        <audio controls>
        <source src="${url}" >
        </audio>        
    `,
    );
}

function createVideoElement(title, url) {
    return createElement(
        "figure",
        {},
        `
        <figcaption>${title}</figcaption>
        <video controls>
        <source src="${url}" >
        </audio>        
    `,
    );
}

function createImageElement(title, url) {
    return createElement(
        "div",
        { class: "image" },
        `
                          <a href="${url}">
                          <img src="${url}" alt=${title}/>
                          </a>
    `,
    );
}

function createFileElement(title, url) {
    return createElement("div", { class: "file" }, `<a href="${url}">${title}</a>`);
}

function createList(title, recordId, files, elFunc) {
    if (files.length === 0) {
        return document.createDocumentFragment();
    }
    const sectionEl = createRecordSection(`${title} (${files.length})`);
    const section = sectionEl.querySelector(".content");
    const galleryEl = createElement("div", { class: "files" });

    for (const file of files) {
        galleryEl.appendChild(
            elFunc(
                file.name,
                `/chronicler/records/${recordId}?file=${encodeURIComponent(file.fileUrl)}`,
            ),
        );
    }
    section.appendChild(galleryEl);
    return sectionEl;
}

export function createRecordSet(rs, metadata) {
    const recordEl = createElement("div", { id: rs["id"], class: "record" });
    if (!rs.rootRecord) {
        recordEl.innerHTML = `<div class='content error'>No record for id ${rs["id"]}</div>`;
        return recordEl;
    }
    const text = (rs.description ?? "")
        .split("\n")
        .map((s) => `<p>${s.trim()}</p>`)
        .join("<br/>");
    const wrapper = createElement("label", { class: "record_wrapper", for: `${rs.id}_checkbox` });
    wrapper.innerHTML = `
    <input type="checkbox" class="selection_marker" data-record="${rs.id}" id="${rs.id}_checkbox" />
    <div class="record" id="${rs.id}">
      <div class='header'>
        <span class="icon ${rs.rootRecord.source.sourceType.name}">&nbsp;</span>
        <div class="origin">
          <span class="source">${getSourceName(metadata, rs.rootRecord.parent)}</span>
          <span class="sender">${getSourceName(
              metadata,
              rs.rootRecord.source,
              rs.rootRecord.parent,
          )}</span>
        </div>
        <span class="datetime">${formatDateTime(rs.rootRecord.time)}</span>
        <a href="?record_id=${rs.id}">${rs.recordCount}</a>
        <a href="/chronicler/records/${rs.id}?file=record.json">json<a>
      </div>
      <div class="content">${text}</div>
    </div>
    `;
    return wrapper;
}

export function createRecord(rsId, record, metadata) {
    const src = record.source;
    const parentSrc = record.parent;

    const recordName = src ? getSourceName(metadata, src) : "NONE";
    const parentName = parentSrc ? getSourceName(metadata, null, parentSrc) : "NONE";
    const parentMsg = parentSrc ? parentSrc.messageId : "NONE";

    const recordEl = createElement("div", {
        id: record.source.messageId,
        class: "record",
    });

    const text = record.textContent
        .split("\n")
        .map((s) => `<p>${s.trim()}</p>`)
        .join("<br/>");
    recordEl.innerHTML = `<div class='header'>
        <span class="icon ${src.sourceType.name}">&nbsp;</span>
        <span class="datetime">${formatDateTime(record.time)}</span>
        <a href="#${src.messageId}">#</a>
            <span class="username">${recordName}</span>
            <span class="username">→ <a href="#${parentMsg}">${parentName}</a></span>
        </div>
        <div class='content'>${text}</div>`;
    recordEl.appendChild(renderFileList(rsId, record.files));
    recordEl.appendChild(
        createList(
            "Links",
            rsId,
            record.links.map((l) => ({ fileUrl: l, name: l })),
            createFileElement,
        ),
    );
    return recordEl;
}

function renderFileList(rsId, files) {
    const result = document.createDocumentFragment();
    /* --- */
    files.sort((a, b) => {
        return a.name.localeCompare(b.name);
    });
    result.appendChild(
        createList(
            "Images",
            rsId,
            files.filter((f) => f.isImage),
            createImageElement,
        ),
    );
    result.appendChild(
        createList(
            "Audios",
            rsId,
            files.filter((f) => f.isAudio),
            createAudioElement,
        ),
    );
    result.appendChild(
        createList(
            "Videos",
            rsId,
            files.filter((f) => f.isVideo),
            createVideoElement,
        ),
    );
    result.appendChild(
        createList(
            "Documents",
            rsId,
            files.filter((f) => f.isDocument),
            createFileElement,
        ),
    );
    return result;
}

export function createRecordSetSummary(rs) {
    const el = createElement("div", { class: "summary record" });
    el.appendChild(renderFileList(rs.id, rs.allFiles));
    el.appendChild(
        createList(
            "All links",
            rs.id,
            rs.allLinks.map((l) => ({ fileUrl: l, name: l })),
            createFileElement,
        ),
    );
    return el;
}

export function createStatus(metric) {
    const el = document.createElement("div");
    let value = "";
    el.setAttribute("class", "menuitem");
    if (metric.Value.hasOwnProperty("IntValue")) {
        value = metric.Value.IntValue;
    } else if (metric.Value.hasOwnProperty("DoubleValue")) {
        value = metric.Value.DoubleValue;
    } else if (metric.Value.hasOwnProperty("IntRangeValue")) {
        value = metric.Value.IntRangeValue;
    } else if (metric.Value.hasOwnProperty("DoubleRangeValue")) {
        value = metric.Value.DoubleRangeValue;
    } else if (metric.Value.hasOwnProperty("DateTimeValue")) {
        value = new Date(metric.Value.DateTimeValue.timestamp).toLocaleString("ru-RU");
    } else {
        value = JSON.stringify(metric.Value);
    }
    el.innerHTML =
        `<div class="statname">${metric.name}</div>` + `<div class="statvalue">${value}</div>`;
    return el;
}
