export function pad(number) {
    return String(number).padStart(2, '0');
}

export function formatDateTime(date) {
    if (!date) {
        return "No date";
    }
    return `${pad(date.getDate())}.${pad(date.getMonth() + 1)}.${date.getFullYear()} ` +
        `${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function createElement(name, attributes) {
    const el = document.createElement(name);
    Object.entries(attributes ?? {}).map(obj => {
        el.setAttribute(obj[0], obj[1]);
    });
    return el;
}

export function createAudioPlaylist(recordId, audios) {
    if (audios.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement('div', { 'class': 'fileset' });
    const galleryEl = createElement('div', { 'class': 'files' });

    imagesEl.innerHTML += `<div class="title">Audio</div>`;
    for (const file of audios) {
        galleryEl.innerHTML += `
                        <figure>
                          <figcaption>${file.name}</figcaption>
                          <audio controls>
                            <source src="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}" >
                          </audio>
                        </figure>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}

export function createVideoPlaylist(recordId, videos) {
    if (videos.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement('div', { 'class': 'fileset' });
    const galleryEl = createElement('div', { 'class': 'gallery' });

    imagesEl.innerHTML += `<div class="title">Video</div>`;
    for (const file of videos) {
        galleryEl.innerHTML += `
                        <figure>
                          <figcaption>${file.name}</figcaption>
                          <video controls>
                            <source src="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}" >
                          </video>
                        </figure>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}


export function createGallery(recordId, images) {
    if (images.length === 0) {
        return document.createDocumentFragment();
    }
    const imagesEl = createElement('div', { 'class': 'fileset' });
    const galleryEl = createElement('div', { 'class': 'gallery' });

    imagesEl.innerHTML += `<div class="title">Images</div>`;
    for (const file of images) {
        galleryEl.innerHTML += `<div class="image">
                          <a href="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}">
                          <img src="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}" />
                          </a>
                      </div>`;
    }
    imagesEl.appendChild(galleryEl);
    return imagesEl;
}

export function createFileList(recordId, files) {
    if (files.length === 0) {
        return document.createDocumentFragment();
    }
    const filesEl = createElement('div', { 'class': 'fileset ' });
    const setEl = createElement('div', { 'class': 'files' });

    filesEl.innerHTML += `<div class="title">All files</div>`;
    for (const file of files) {
        setEl.innerHTML += `<div class="file">
                     <a href="chronicler/${recordId}?file=${encodeURIComponent(file.fileUrl)}">${file.name}</a>
                 </div>`;
    }
    filesEl.appendChild(setEl);
    return filesEl;
}
