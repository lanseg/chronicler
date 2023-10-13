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

function createElement(name, attributes) {
    const el = document.createElement(name);
    Object.entries(attributes ?? {}).map((obj) => {
        el.setAttribute(obj[0], obj[1]);
    });
    return el;
}

function createAudioPlaylist(recordId, audios) {
    if (audios.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement("div", { class: "fileset" });
    const galleryEl = createElement("div", { class: "files" });

    imagesEl.innerHTML += `<div class="title">Audio</div>`;
    for (const file of audios) {
        galleryEl.innerHTML += `
                        <figure>
                          <figcaption>${file.name}</figcaption>
                          <audio controls>
                            <source src="chronicler/${recordId}?file=${encodeURIComponent(
                                file.fileUrl,
                            )}" >
                          </audio>
                        </figure>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}

function createVideoPlaylist(recordId, videos) {
    if (videos.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement("div", { class: "fileset" });
    const galleryEl = createElement("div", { class: "gallery" });

    imagesEl.innerHTML += `<div class="title">Video</div>`;
    for (const file of videos) {
        galleryEl.innerHTML += `
                        <figure>
                          <figcaption>${file.name}</figcaption>
                          <video controls>
                            <source src="chronicler/${recordId}?file=${encodeURIComponent(
                                file.fileUrl,
                            )}" >
                          </video>
                        </figure>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}

function createGallery(recordId, images) {
    if (images.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement("div", { class: "fileset" });
    const galleryEl = createElement("div", { class: "gallery" });

    imagesEl.innerHTML += `<div class="title">Images</div>`;
    for (const file of images) {
        galleryEl.innerHTML += `<div class="image">
                          <a href="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}">
                          <img src="chronicler/${recordId}?file=${encodeURIComponent(
                              file.fileUrl,
                          )}" />
                          </a>
                      </div>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}

function createFileList(recordId, title, files) {
    if (files.length === 0) {
        return document.createDocumentFragment();
    }
    const filesEl = createElement("div", { class: "fileset " });
    const setEl = createElement("div", { class: "files" });

    filesEl.innerHTML += `<div class="title">${title}</div>`;
    for (const file of files) {
        setEl.innerHTML += `<div class="file">
                     <a href="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}">${
                         file.name
                     }</a>
                 </div>`;
    }
    filesEl.appendChild(setEl);
    return filesEl;
}

export function createRecordSet(rs, metadata) {
    const recordEl = createElement("div", { id: rs["id"], class: "record" });
    if (!rs.rootRecord) {
        recordEl.innerHTML = `<div class='content error'>No record for id ${rs["id"]}</div>`;
        return recordEl;
    }
    const srcName = getSourceName(metadata, rs.rootRecord.source, rs.rootRecord.parent);
    const timeLabel = formatDateTime(rs.rootRecord.time);
    recordEl.innerHTML = `<div class='header'>
        <span class="icon ${rs.rootRecord.source.sourceType.name}">&nbsp;</span>
        <span class="sourcename">${srcName}</span>
        <span class="datetime">${timeLabel}</span>
        <a href="?record_id=${rs.id}">${rs.recordCount}</a>
        <a href="/chronicler/${rs.id}?file=record.json">json<a>
      </div>
      <div class="content">${rs.description}</div>`;
    return recordEl;
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

    recordEl.innerHTML = `<div class='header'>
        <span class="icon ${src.sourceType.name}">&nbsp;</span>
        <span class="datetime">${formatDateTime(record.time)}</span>
        <a href="#${src.messageId}">#</a>
            <span class="username">${recordName}</span>
            <span class="username">→ <a href="#${parentMsg}">${parentName}</a></span>
        </div>
        <div class='content'>${record.textContent.replaceAll("\n", "<br/>")}</div>`;

    /* --- */
    record.files.sort((a, b) => {
        return a.name.localeCompare(b.name);
    });
    recordEl.appendChild(
        createGallery(
            rsId,
            record.files.filter((f) => f.isImage),
        ),
    );
    recordEl.appendChild(
        createAudioPlaylist(
            rsId,
            record.files.filter((f) => f.isAudio),
        ),
    );
    recordEl.appendChild(
        createVideoPlaylist(
            rsId,
            record.files.filter((f) => f.isVideo),
        ),
    );
    recordEl.appendChild(
        createFileList(
            rsId,
            "Documents",
            record.files.filter((f) => f.isDocument),
        ),
    );
    recordEl.appendChild(createFileList(rsId, "All files", record.files));
    return recordEl;
}
